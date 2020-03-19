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
)

const promExportPath = "/metrics"

type exportPrometheusCmd struct {
	Servers []string `usage:"one or more [host:port] addresses of servers to monitor, when port is omitted 19132 is used"`
	Port    int      `usage:"HTTP port where Prometheus metrics are exported" default:"8080"`
	Edition string   `usage:"The edition of Minecraft server, java or bedrock" default:"java"`
	logger  *zap.Logger
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

func (c *exportPrometheusCmd) Execute(_ context.Context, _ *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(c.Servers) == 0 {
		printUsageError("requires at least one server")
		return subcommands.ExitUsageError
	}
	if !ValidEdition(c.Edition) {
		printUsageError("invalid edition")
		return subcommands.ExitUsageError
	}

	logger := args[0].(*zap.Logger)

	collectors, err := newPromCollectors(c.Servers, ServerEdition(c.Edition), logger)
	if err != nil {
		log.Fatal(err)
	}

	err = prometheus.Register(collectors)
	if err != nil {
		log.Fatal(err)
	}

	exportAddress := ":" + strconv.Itoa(c.Port)

	logger.Info("exporting metrics for prometheus",
		zap.String("address", exportAddress),
		zap.String("path", promExportPath),
		zap.Strings("servers", c.Servers))

	http.Handle(promExportPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(exportAddress, nil))

	// never actually returns from ListenAndServe, so just satisfy return value
	return subcommands.ExitFailure
}
