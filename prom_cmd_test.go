package main

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/google/subcommands"
	"github.com/itzg/zapconfigs"
	"github.com/stretchr/testify/require"
)

func TestExportPrometheusProxyFlags(t *testing.T) {
	t.Setenv("EXPORT_USE_PROXY", "false")
	t.Setenv("EXPORT_PROXY_VERSION", "1")

	cmd := &exportPrometheusCmd{}
	flags := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	cmd.SetFlags(flags)

	err := flags.Parse([]string{"-servers", "localhost", "-use-proxy", "-proxy-version", "2"})

	require.NoError(t, err)
	require.Equal(t, []string{"localhost"}, cmd.Servers)
	require.True(t, cmd.UseProxy)
	require.Equal(t, uint(2), cmd.ProxyVersion)
	require.Equal(t, time.Minute, cmd.Timeout)
}

func TestExportPrometheusProxyEnvironment(t *testing.T) {
	t.Setenv("EXPORT_USE_PROXY", "true")
	t.Setenv("EXPORT_PROXY_VERSION", "2")

	cmd := &exportPrometheusCmd{}
	flags := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	cmd.SetFlags(flags)

	require.True(t, cmd.UseProxy)
	require.Equal(t, uint(2), cmd.ProxyVersion)
}

func TestExportPrometheusExitsOnContextCancel(t *testing.T) {
	cmd := &exportPrometheusCmd{}
	flags := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
	cmd.SetFlags(flags)
	require.NoError(t, flags.Parse([]string{"-servers", "localhost", "-port", "0"}))

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan subcommands.ExitStatus, 1)
	go func() {
		done <- cmd.Execute(ctx, flags, zapconfigs.NewDefaultLogger())
	}()

	// Give the server a moment to start listening before asking it to stop.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case status := <-done:
		require.Equal(t, subcommands.ExitSuccess, status)
	case <-time.After(2 * time.Second):
		t.Fatal("Execute did not return after context cancellation -- SIGTERM would hang indefinitely")
	}
}

