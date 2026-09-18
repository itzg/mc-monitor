package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/subcommands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStatusBedrockCmd_Flags(t *testing.T) {
	cmd := &statusBedrockCmd{}
	flags := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	cmd.SetFlags(flags)

	err := flags.Parse([]string{
		"-host", "bedrock.example.com",
		"-port", "19133",
		"-protocol", "nethernet",
		"-retry-interval", "5s",
		"-retry-limit", "3",
	})
	require.NoError(t, err)

	assert.Equal(t, "bedrock.example.com", cmd.Host)
	assert.Equal(t, 19133, cmd.Port)
	assert.Equal(t, "nethernet", cmd.Protocol)
	assert.Equal(t, 5*time.Second, cmd.RetryInterval)
	assert.Equal(t, 3, cmd.RetryLimit)
}

func TestStatusBedrockCmd_Execute_Nethernet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"Hello world","protocol":2193,"version":"1.26.50","level":"Bedrock level","players":0,"maxPlayers":10,"gameType":0}`))
	}))
	defer ts.Close()

	parts := strings.Split(strings.TrimPrefix(ts.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	cmd := &statusBedrockCmd{
		Host:     host,
		Port:     port,
		Protocol: "nethernet",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := zap.NewNop()
	status := cmd.Execute(context.Background(), nil, logger)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	assert.Equal(t, subcommands.ExitSuccess, status)
	assert.Equal(t, host+":"+strconv.Itoa(port)+" : version=1.26.50 online=0 max=10", buf.String())
}

func TestStatusBedrockCmd_Execute_Auto(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"Hello world","protocol":2193,"version":"1.26.50","level":"Bedrock level","players":0,"maxPlayers":10,"gameType":0}`))
	}))
	defer ts.Close()

	parts := strings.Split(strings.TrimPrefix(ts.URL, "http://"), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	cmd := &statusBedrockCmd{
		Host:     host,
		Port:     port,
		Protocol: "auto",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := zap.NewNop()
	status := cmd.Execute(context.Background(), nil, logger)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	assert.Equal(t, subcommands.ExitSuccess, status)
	assert.Equal(t, host+":"+strconv.Itoa(port)+" : version=1.26.50 online=0 max=10", buf.String())
}
