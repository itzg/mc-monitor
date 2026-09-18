package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPingBedrockNetherNet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/join", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"Hello world","protocol":2193,"version":"1.26.50","level":"Bedrock level","players":0,"maxPlayers":10,"gameType":0}`))
	}))
	defer ts.Close()

	logger := zap.NewNop()
	info, err := PingBedrockNetherNet(ts.URL, 5*time.Second, logger)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, "Hello world", info.ServerName)
	assert.Equal(t, "2193", info.ProtocolVersion)
	assert.Equal(t, "1.26.50", info.Version)
	assert.Equal(t, "Bedrock level", info.LevelName)
	assert.Equal(t, 0, info.Players)
	assert.Equal(t, 10, info.MaxPlayers)
	assert.Equal(t, "Survival", info.GameMode)
	assert.True(t, info.Rtt > 0)
}

func TestPingBedrockNetherNet_Variants(t *testing.T) {
	tests := []struct {
		name         string
		jsonBody     string
		expectedMode string
		expectedProt string
	}{
		{
			name:         "creative mode",
			jsonBody:     `{"name":"Creative Server","protocol":2193,"version":"1.26.50","level":"Flat","players":3,"maxPlayers":20,"gameType":1}`,
			expectedMode: "Creative",
			expectedProt: "2193",
		},
		{
			name:         "adventure mode with string protocol",
			jsonBody:     `{"name":"Adventure Server","protocol":"2200","version":"1.26.50","level":"Quest","players":1,"maxPlayers":5,"gameType":2}`,
			expectedMode: "Adventure",
			expectedProt: "2200",
		},
		{
			name:         "spectator mode",
			jsonBody:     `{"name":"Spectator Server","protocol":2193,"version":"1.26.50","level":"Arena","players":0,"maxPlayers":5,"gameType":6}`,
			expectedMode: "Spectator",
			expectedProt: "2193",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.jsonBody))
			}))
			defer ts.Close()

			info, err := PingBedrockNetherNet(ts.URL, 5*time.Second, zap.NewNop())
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMode, info.GameMode)
			assert.Equal(t, tt.expectedProt, info.ProtocolVersion)
		})
	}
}

func TestPingBedrockNetherNet_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	info, err := PingBedrockNetherNet(ts.URL, 5*time.Second, zap.NewNop())
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "status 404")
}

func TestPingBedrockNetherNet_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	info, err := PingBedrockNetherNet(ts.URL, 5*time.Second, zap.NewNop())
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "Failed to decode response")
}
