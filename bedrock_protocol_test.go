package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newNetherNetTestServer(t *testing.T) (host string, port uint16) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"Hello world","protocol":2193,"version":"1.26.63","level":"Bedrock level","players":3,"maxPlayers":10,"gameType":0}`))
	}))
	t.Cleanup(ts.Close)

	host, portStr, err := net.SplitHostPort(ts.Listener.Addr().String())
	require.NoError(t, err)
	p, err := strconv.ParseUint(portStr, 10, 16)
	require.NoError(t, err)
	return host, uint16(p)
}

func TestNormalizeBedrockProtocol(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", BedrockProtocolAuto},
		{"auto", BedrockProtocolAuto},
		{"NetherNet", BedrockProtocolNetherNet},
		{"raknet", BedrockProtocolRakNet},
	}
	for _, tt := range tests {
		got, err := normalizeBedrockProtocol(tt.in)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
	}

	_, err := normalizeBedrockProtocol("udp")
	assert.Error(t, err)
}

func TestPingBedrockAutoUsesNetherNet(t *testing.T) {
	host, port := newNetherNetTestServer(t)

	info, err := PingBedrock(net.JoinHostPort(host, strconv.Itoa(int(port))), BedrockProtocolAuto, 5*time.Second, zap.NewNop())
	require.NoError(t, err)
	assert.Equal(t, "1.26.63", info.Version)
	assert.Equal(t, 3, info.Players)
}

func TestPromBedrockCollectorNetherNet(t *testing.T) {
	host, port := newNetherNetTestServer(t)

	for _, protocol := range []string{BedrockProtocolAuto, BedrockProtocolNetherNet} {
		t.Run(protocol, func(t *testing.T) {
			collector := newPromBedrockCollector(host, port, protocol, zap.NewNop())
			collector.SetTimeout(5 * time.Second)

			values := collectBedrockMetrics(t, collector)
			assert.Equal(t, 1.0, values[promDescHealthy.String()])
			assert.Equal(t, 3.0, values[promDescPlayersOnline.String()])
			assert.Equal(t, 10.0, values[promDescPlayersMax.String()])
		})
	}
}

func collectBedrockMetrics(t *testing.T, collector *promBedrockCollector) map[string]float64 {
	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	values := map[string]float64{}
	for metric := range ch {
		var m dto.Metric
		require.NoError(t, metric.Write(&m))
		values[metric.Desc().String()] = m.GetGauge().GetValue()
	}
	return values
}
