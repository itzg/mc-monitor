package main

import (
	"flag"
	"testing"
	"time"

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
