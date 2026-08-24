package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strconv"
	"time"
)

const promExportPath = "/metrics"

type exportPrometheusCmd struct {
	Servers        []string      `usage:"one or more [host:port] addresses of Java servers to monitor, when port is omitted 25565 is used"`
	BedrockServers []string      `usage:"one or more [host:port] addresses of Bedrock servers to monitor, when port is omitted 19132 is used"`
	Port           int           `usage:"HTTP port where Prometheus metrics are exported" default:"8080"`
	Timeout        time.Duration `usage:"timeout when checking each servers" default:"60s" env:"TIMEOUT"`
	UseProxy       bool          `usage:"supports contacting servers when proxy_protocol is enabled"`
	ProxyVersion   uint          `usage:"version of PROXY protocol to use" default:"1"`
	logger         *zap.Logger
}

func (c *exportPrometheusCmd) Name() string {
	return "export-for-prometheus"
}

func (c *exportPrometheusCmd) Synopsis() string {
	return "Registers an HTTP metrics endpoints for Prometheus export"
}

func (c *exportPrometheusCmd) Usage() string {
	return ""
}

func (c *exportPrometheusCmd) SetFlags(f *flag.FlagSet) {
	filler := flagsfiller.New(flagsfiller.WithEnv("Export"))
	err := filler.Fill(f, c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *exportPrometheusCmd) Execute(ctx context.Context, _ *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if (len(c.Servers) + len(c.BedrockServers)) == 0 {
		printUsageError("requires at least one server")
		return subcommands.ExitUsageError
	}

	logger := args[0].(*zap.Logger)

	collectors, err := newPromCollectors(c.Servers, c.BedrockServers, c.UseProxy, c.ProxyVersion, logger)
	if err != nil {
		log.Fatal(err)
	}

	for i := range collectors {
		collectors[i].SetTimeout(c.Timeout)
	}

	err = prometheus.Register(collectors)
	if err != nil {
		log.Fatal(err)
	}

	exportAddress := ":" + strconv.Itoa(c.Port)

	logger.Info("exporting metrics for prometheus",
		zap.String("address", exportAddress),
		zap.String("path", promExportPath),
	)

	mux := http.NewServeMux()
	mux.Handle(promExportPath, promhttp.Handler())
	server := &http.Server{Addr: exportAddress, Handler: mux}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to cleanly shut down metrics server", zap.Error(err))
			return subcommands.ExitFailure
		}

		return subcommands.ExitSuccess

	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}

		return subcommands.ExitSuccess
	}
}
