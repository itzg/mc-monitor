package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/zap"
)

type collectOpenTelemetryCmd struct {
	Servers        []string      `usage:"one or more [host:port] addresses of Java servers to monitor, when port is omitted 25565 is used"`
	BedRockServers []string      `usage:"one or more [host:port] addresses of Bedrock servers to monitor, when port is omitted 19132 is used"`
	Interval       time.Duration `default:"10s" usage:"Collect and sends OpenTelemetry data at this interval"`
	Exporter       struct {
		Endpoint string        `default:"localhost:4317" usage:"OpenTelemetry gRPC endpoint to export data"`
		Timeout  time.Duration `default:"35s" usage:"Timeout for collecting OpenTelemetry data"`
	} `group:"exporter" namespace:"exporter" usage:"Open Telemetry Exporter configurations"`
	logger *zap.Logger
}

// ShutdownFunc is a function that can be called to shut down the Open Telemetry provider components
type ShutdownFunc func() error

var _histogramBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10, 25, 50, 100}

func (c *collectOpenTelemetryCmd) Name() string {
	return "collect-for-opentelemetry"
}

func (c *collectOpenTelemetryCmd) Synopsis() string {
	return "Starts collecting telemetry data using OpenTelemetry"
}

func (c *collectOpenTelemetryCmd) Usage() string {
	return ""
}

func (c *collectOpenTelemetryCmd) SetFlags(f *flag.FlagSet) {
	filler := flagsfiller.New(flagsfiller.WithEnv("OpenTelemetry"))
	err := filler.Fill(f, c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *collectOpenTelemetryCmd) Execute(ctx context.Context, _ *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// Validate the command line arguments
	if (len(c.Servers) + len(c.BedRockServers)) == 0 {
		printUsageError("requires at least one server")
		return subcommands.ExitUsageError
	}

	if c.Exporter.Endpoint == "" {
		printUsageError("the open telemetry endpoint must be set")
		return subcommands.ExitUsageError
	}

	// Start the OpenTelemetry meter provider
	meterShutdownFunc, err := c.startMeterProvider(ctx)
	if err != nil {
		printUsageError(fmt.Sprintf("failed to start meter provider: %v", err))
		return subcommands.ExitFailure
	}

	// Set the logger for the OpenTelemetry components
	c.logger = args[0].(*zap.Logger).Named("otel")

	// Create the  resources to be monitored
	resources := make([]otelResource, 0)
	metricChecker, err := c.initializeMetricResources()
	if err != nil {
		printUsageError(fmt.Sprintf("failed to create metric checker: %v", err))
		return subcommands.ExitFailure
	}
	resources = append(resources, metricChecker...)

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
func (c *collectOpenTelemetryCmd) startMeterProvider(ctx context.Context) (ShutdownFunc, error) {
	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(c.Exporter.Endpoint), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
				metric.WithTimeout(c.Exporter.Timeout),
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
func (c *collectOpenTelemetryCmd) initializeMetricResources() (
	[]otelResource,
	error,
) {
	resources := make([]otelResource, 0)

	for _, server := range c.Servers {
		host, port, err := SplitHostPort(server, DefaultJavaPort)
		if err != nil {
			return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
		}
		c.logger.Info("adding Java server", zap.String("host", host), zap.Uint16("port", port))

		resource, err := newOpenTelemetryMetricResource(
			host,
			port,
			withServerEdition(JavaEdition),
			withLogger(c.logger),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Java resource: %w", err)
		}
		resources = append(resources, resource)
	}

	for _, server := range c.BedRockServers {
		host, port, err := SplitHostPort(server, DefaultBedrockPort)
		if err != nil {
			return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
		}
		c.logger.Info("adding Bedrock server", zap.String("host", host), zap.Uint16("port", port))

		resource, err := newOpenTelemetryMetricResource(
			host,
			port,
			withServerEdition(BedrockEdition),
			withLogger(c.logger),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Bedrock resource: %w", err)
		}
		resources = append(resources, resource)
	}

	return resources, nil
}
