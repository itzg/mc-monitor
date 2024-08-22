package otel

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/mc-monitor/utils"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
)

type CollectOpenTelemetryCmd struct {
	Servers        []string      `usage:"one or more [host:port] addresses of Java servers to monitor, when port is omitted 25565 is used"`
	BedrockServers []string      `usage:"one or more [host:port] addresses of Bedrock servers to monitor, when port is omitted 19132 is used"`
	Interval       time.Duration `default:"10s" usage:"Collect and sends OpenTelemetry data at this interval"`
	OtelCollector  Collector     `group:"exporter" namespace:"exporter" usage:"Open Telemetry OtelCollector configurations"`
	logger         *zap.Logger
}

type Collector struct {
	Endpoint string        `default:"localhost:4317" usage:"OpenTelemetry gRPC endpoint to export data"`
	Timeout  time.Duration `default:"35s" usage:"Timeout for collecting OpenTelemetry data"`
}

// ShutdownFunc is a function that can be called to shut down the Open Telemetry provider components
type ShutdownFunc func() error

var _histogramBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10, 25, 50, 100}

func (c *CollectOpenTelemetryCmd) Name() string {
	return "collect-otel"
}

func (c *CollectOpenTelemetryCmd) Synopsis() string {
	return "Starts collecting telemetry data using OpenTelemetry"
}

func (c *CollectOpenTelemetryCmd) Usage() string {
	return ""
}

func (c *CollectOpenTelemetryCmd) SetFlags(f *flag.FlagSet) {
	filler := flagsfiller.New(flagsfiller.WithEnv("Export"))
	err := filler.Fill(f, c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *CollectOpenTelemetryCmd) Execute(ctx context.Context, _ *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// Validate the command line arguments
	if (len(c.Servers) + len(c.BedrockServers)) == 0 {
		utils.PrintUsageError("requires at least one server")
		return subcommands.ExitUsageError
	}

	if c.OtelCollector.Endpoint == "" {
		utils.PrintUsageError("the open telemetry endpoint must be set")
		return subcommands.ExitUsageError
	}

	// Start the OpenTelemetry meter provider
	meterShutdownFunc, err := c.startMeterProvider(ctx)
	if err != nil {
		utils.PrintUsageError(fmt.Sprintf("failed to start meter provider: %v", err))
		return subcommands.ExitFailure
	}

	// Set the logger for the OpenTelemetry components
	c.logger = args[0].(*zap.Logger).Named("otel")

	// Create the  resources to be monitored
	resources, err := c.initializeMetricResources()
	if err != nil {
		utils.PrintUsageError(fmt.Sprintf("failed to create metric checker: %v", err))
		return subcommands.ExitFailure
	}

	// Start the observing loop
	ticker := time.NewTicker(c.Interval)

	for {
		select {
		case <-ctx.Done():
			if err := meterShutdownFunc(); err != nil {
				return subcommands.ExitFailure
			}

			return subcommands.ExitSuccess

		case <-ticker.C:
			c.logger.Info("collecting OpenTelemetry data")

			for _, r := range resources {
				go r.Execute()
			}
		}
	}
}

// startMeterProvider constructs and starts the exporter that will be sending telemetry data from a meter provider that is set
func (c *CollectOpenTelemetryCmd) startMeterProvider(ctx context.Context) (ShutdownFunc, error) {
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(c.OtelCollector.Endpoint), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
				metric.WithTimeout(c.OtelCollector.Timeout),
				metric.WithInterval(c.Interval),
			),
		),
		metric.WithView(
			metric.NewView(
				metric.Instrument{
					Name: "*",
					Kind: metric.InstrumentKindHistogram,
				},
				metric.Stream{Aggregation: metric.AggregationExplicitBucketHistogram{Boundaries: _histogramBuckets}},
			),
		),
	)

	otel.SetMeterProvider(meterProvider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(c.Interval))
	if err != nil {
		return nil, err
	}

	return func() error {
		return meterProvider.Shutdown(ctx)
	}, nil
}

// initializeMetricResources creates the OpenTelemetry Metric resources for the given servers
func (c *CollectOpenTelemetryCmd) initializeMetricResources() (
	[]Resource,
	error,
) {
	resources := make([]Resource, 0)

	for _, server := range c.Servers {
		host, port, err := utils.SplitHostPort(server, utils.DefaultJavaPort)
		if err != nil {
			return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
		}
		c.logger.Info("adding Java server", zap.String("host", host), zap.Uint16("port", port))

		resource, err := newOpenTelemetryMetricResource(
			host,
			port,
			withServerEdition(utils.JavaEdition),
			withLogger(c.logger),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Java resource: %w", err)
		}
		resources = append(resources, resource)
	}

	for _, server := range c.BedrockServers {
		host, port, err := utils.SplitHostPort(server, utils.DefaultBedrockPort)
		if err != nil {
			return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
		}
		c.logger.Info("adding Bedrock server", zap.String("host", host), zap.Uint16("port", port))

		resource, err := newOpenTelemetryMetricResource(
			host,
			port,
			withServerEdition(utils.BedrockEdition),
			withLogger(c.logger),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Bedrock resource: %w", err)
		}
		resources = append(resources, resource)
	}

	return resources, nil
}
