package otel

import (
	"strconv"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/itzg/mc-monitor/utils"
	"go.uber.org/zap"
)

type Resource interface {
	Execute()
}

type OpenTelemetryMetricResource struct {
	host    string
	port    uint16
	edition utils.ServerEdition
	pinger  mcpinger.Pinger
	metrics *ServerMetrics
	logger  *zap.Logger
}

type OpenTelemetryMetricResourceOptions func(r *OpenTelemetryMetricResource)

func withServerEdition(edition utils.ServerEdition) OpenTelemetryMetricResourceOptions {
	return func(r *OpenTelemetryMetricResource) {
		r.edition = edition
	}
}

func withServerMetrics(logger *zap.Logger) OpenTelemetryMetricResourceOptions {
	return func(r *OpenTelemetryMetricResource) {
		r.metrics = NewServerMetrics(logger)
	}
}

func withLogger(logger *zap.Logger) OpenTelemetryMetricResourceOptions {
	return func(r *OpenTelemetryMetricResource) {
		r.logger = logger
	}
}

func newOpenTelemetryMetricResource(host string, port uint16, options ...OpenTelemetryMetricResourceOptions) (
	*OpenTelemetryMetricResource,
	error,
) {
	resource := &OpenTelemetryMetricResource{
		host:   host,
		port:   port,
		pinger: mcpinger.New(host, port),
	}

	for _, option := range options {
		option(resource)
	}

	return resource, nil
}

func (r *OpenTelemetryMetricResource) Execute() {
	r.logger.Debug("pinging", zap.String("host", r.host), zap.String("port", strconv.Itoa(int(r.port))))
	startTime := time.Now()
	info, err := r.pinger.Ping()
	elapsed := time.Now().Sub(startTime)
	r.logger.Debug("ping returned", zap.Error(err), zap.Any("info", info))
	r.logger.Debug("measured elapsed time", zap.Float64("elapsed", elapsed.Seconds()))

	if r.metrics != nil {
		if err != nil || info.Players.Max == 0 {
			r.metrics.RecordHealth(false, buildMetricAttributes(r.host, r.port, r.edition, ""))
			return
		}

		r.metrics.RecordResponseTime(elapsed.Seconds(), buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
		r.metrics.RecordHealth(true, buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
		r.metrics.RecordPlayersOnlineCount(info.Players.Online, buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
		r.metrics.RecordPlayersMaxCount(info.Players.Max, buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	}
}
