package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"
	"unicode/utf16"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestJavaPingerOptions(t *testing.T) {
	tests := []struct {
		name               string
		useProxy           bool
		proxyVersion       byte
		expectProxy        bool
		expectProxyVersion byte
	}{
		{name: "disabled", useProxy: false, proxyVersion: 2},
		{name: "version one", useProxy: true, proxyVersion: 1, expectProxy: true, expectProxyVersion: 1},
		{name: "version two", useProxy: true, proxyVersion: 2, expectProxy: true, expectProxyVersion: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &promJavaCollector{
				timeout:      5 * time.Second,
				useProxy:     tt.useProxy,
				proxyVersion: tt.proxyVersion,
			}
			pinger := mcpinger.New("localhost", DefaultJavaPort, javaPingerOptions(collector)...)
			value := reflect.ValueOf(pinger).Elem()

			assert.Equal(t, 5*time.Second, time.Duration(value.FieldByName("Timeout").Int()))
			assert.Equal(t, tt.expectProxy, value.FieldByName("UseProxy").Bool())
			assert.Equal(t, uint64(tt.expectProxyVersion), value.FieldByName("ProxyVersion").Uint())
		})
	}
}

func TestNewPromCollectorsPropagatesProxyConfig(t *testing.T) {
	collectors, err := newPromCollectors(
		[]string{"java.example.com"},
		[]string{"bedrock.example.com"},
		true,
		2,
		zap.NewNop(),
	)

	require.NoError(t, err)
	require.Len(t, collectors, 2)

	javaCollector, ok := collectors[0].(*promJavaCollector)
	require.True(t, ok)
	assert.True(t, javaCollector.useProxy)
	assert.Equal(t, byte(2), javaCollector.proxyVersion)

	_, ok = collectors[1].(*promBedrockCollector)
	require.True(t, ok)
}

func TestNewPromCollectorsRejectsInvalidProxyVersion(t *testing.T) {
	_, err := newPromCollectors([]string{"java.example.com"}, nil, true, 3, zap.NewNop())

	require.EqualError(t, err, "proxy version must be 1 or 2")
}

// startLegacyServer accepts connections like a legacy Java server and answers each server list ping
// with a kick packet containing the given UTF-16BE encoded content.
func startLegacyServer(t *testing.T, content string) (host string, port uint16) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	encoded := utf16.Encode([]rune(content))
	response := new(bytes.Buffer)
	response.WriteByte(0xff)
	require.NoError(t, binary.Write(response, binary.BigEndian, uint16(len(encoded))))
	require.NoError(t, binary.Write(response, binary.BigEndian, encoded))

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			packetId := make([]byte, 1)
			_, err = io.ReadFull(conn, packetId)
			if err == nil && packetId[0] == 0xfe {
				_, _ = conn.Write(response.Bytes())
			}
			// the response is terminated by closing our side, but the remainder of the ping still needs
			// to be consumed to avoid resetting the connection
			_ = conn.(*net.TCPConn).CloseWrite()
			_, _ = io.Copy(io.Discard, conn)
			_ = conn.Close()
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.IP.String(), uint16(addr.Port)
}

func TestPromJavaCollectorLegacyPing(t *testing.T) {
	tests := []struct {
		name                 string
		useServerListPing    bool
		useOldServerListPing bool
		content              string
		expectVersion        string
		expectResponseTime   bool
		expectValues         map[string]float64
	}{
		{
			name:               "server list ping",
			useServerListPing:  true,
			content:            "§1\x0078\x001.6.4\x00A Minecraft Server\x003\x0020",
			expectVersion:      "1.6.4",
			expectResponseTime: true,
			expectValues: map[string]float64{
				"minecraft_status_healthy":              1,
				"minecraft_status_players_online_count": 3,
				"minecraft_status_players_max_count":    20,
			},
		},
		{
			name:                 "old server list ping",
			useOldServerListPing: true,
			content:              "A Minecraft Server§3§20",
			// the server version is not reported by the old server list ping
			expectVersion:      "",
			expectResponseTime: true,
			expectValues: map[string]float64{
				"minecraft_status_healthy":              1,
				"minecraft_status_players_online_count": 3,
				"minecraft_status_players_max_count":    20,
			},
		},
		{
			name:               "server not ready",
			useServerListPing:  true,
			content:            "§1\x0078\x001.6.4\x00A Minecraft Server\x000\x000",
			expectVersion:      "1.6.4",
			expectResponseTime: true,
			expectValues: map[string]float64{
				"minecraft_status_healthy": 0,
			},
		},
		{
			name:              "invalid player count",
			useServerListPing: true,
			content:           "§1\x0078\x001.6.4\x00A Minecraft Server\x00many\x0020",
			expectVersion:     "",
			expectValues: map[string]float64{
				"minecraft_status_healthy": 0,
			},
		},
		{
			name:                 "old server list ping with invalid response",
			useOldServerListPing: true,
			content:              "A Minecraft Server",
			expectVersion:        "",
			expectValues: map[string]float64{
				"minecraft_status_healthy": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := startLegacyServer(t, tt.content)

			collectors := promCollectors{newPromJavaCollector(host, port, false, 1, zap.NewNop())}
			collectors[0].SetTimeout(5 * time.Second)
			require.NoError(t, collectors.useLegacyPing(tt.useServerListPing, tt.useOldServerListPing))

			registry := prometheus.NewRegistry()
			require.NoError(t, registry.Register(collectors))
			families, err := registry.Gather()
			require.NoError(t, err)

			values := make(map[string]float64)
			for _, family := range families {
				require.Len(t, family.GetMetric(), 1)
				metric := family.GetMetric()[0]

				labels := make(map[string]string)
				for _, label := range metric.GetLabel() {
					labels[label.GetName()] = label.GetValue()
				}
				assert.Equal(t, map[string]string{
					promLabelHost:    host,
					promLabelPort:    strconv.Itoa(int(port)),
					promLabelEdition: string(JavaEdition),
					promLabelVersion: tt.expectVersion,
				}, labels)

				values[family.GetName()] = metric.GetGauge().GetValue()
			}

			// the response time varies, so only its presence can be expected
			_, hasResponseTime := values["minecraft_status_response_time_seconds"]
			assert.Equal(t, tt.expectResponseTime, hasResponseTime)
			delete(values, "minecraft_status_response_time_seconds")

			assert.Equal(t, tt.expectValues, values)
		})
	}
}

func TestUseLegacyPing(t *testing.T) {
	newCollectors := func(useProxy bool) promCollectors {
		return promCollectors{
			newPromJavaCollector("java.example.com", DefaultJavaPort, useProxy, 1, zap.NewNop()),
			&promBedrockCollector{},
		}
	}

	t.Run("disabled", func(t *testing.T) {
		collectors := newCollectors(true)

		require.NoError(t, collectors.useLegacyPing(false, false))

		javaCollector := collectors[0].(*promJavaCollector)
		assert.False(t, javaCollector.useServerListPing)
		assert.False(t, javaCollector.useOldServerListPing)
	})

	t.Run("server list ping", func(t *testing.T) {
		collectors := newCollectors(false)

		require.NoError(t, collectors.useLegacyPing(true, false))

		javaCollector := collectors[0].(*promJavaCollector)
		assert.True(t, javaCollector.useServerListPing)
		assert.False(t, javaCollector.useOldServerListPing)
	})

	t.Run("old server list ping", func(t *testing.T) {
		collectors := newCollectors(false)

		require.NoError(t, collectors.useLegacyPing(false, true))

		javaCollector := collectors[0].(*promJavaCollector)
		assert.False(t, javaCollector.useServerListPing)
		assert.True(t, javaCollector.useOldServerListPing)
	})

	t.Run("rejects both", func(t *testing.T) {
		err := newCollectors(false).useLegacyPing(true, true)

		require.EqualError(t, err, "only one of server list ping or old server list ping can be used")
	})

	t.Run("rejects proxy", func(t *testing.T) {
		err := newCollectors(true).useLegacyPing(true, false)

		require.EqualError(t, err, "proxy protocol is not supported with server list ping")
	})
}
