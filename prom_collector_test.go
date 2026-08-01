package main

import (
	"reflect"
	"testing"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
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
