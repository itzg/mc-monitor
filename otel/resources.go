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
	logger  *zap.Logger
}

type OpenTelemetryMetricResourceOptions func(r *OpenTelemetryMetricResource)

func withServerEdition(edition utils.ServerEdition) OpenTelemetryMetricResourceOptions {
	return func(r *OpenTelemetryMetricResource) {
		r.edition = edition
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
	if err != nil || info.Players.Max == 0 {
		Metrics().RecordHealth(false, buildMetricAttributes(r.host, r.port, r.edition, ""))
		return
	}

	Metrics().RecordResponseTime(elapsed.Seconds(), buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	Metrics().RecordHealth(true, buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	Metrics().RecordPlayersOnlineCount(info.Players.Online, buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	Metrics().RecordPlayersMaxCount(info.Players.Max, buildMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
}
