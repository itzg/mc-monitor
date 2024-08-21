package main

import (
	"strconv"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"go.uber.org/zap"
)

type otelResource interface {
	Execute()
}

type OTelMetricResource struct {
	host    string
	port    uint16
	edition ServerEdition
	pinger  mcpinger.Pinger
	logger  *zap.Logger
}

type OTelMetricResourceOptions func(r *OTelMetricResource)

func withServerEdition(edition ServerEdition) OTelMetricResourceOptions {
	return func(r *OTelMetricResource) {
		r.edition = edition
	}
}

func withLogger(logger *zap.Logger) OTelMetricResourceOptions {
	return func(r *OTelMetricResource) {
		r.logger = logger
	}
}

func newOpenTelemetryMetricResource(host string, port uint16, options ...OTelMetricResourceOptions) (*OTelMetricResource, error) {
	resource := &OTelMetricResource{
		host:   host,
		port:   port,
		pinger: mcpinger.New(host, port),
	}

	for _, option := range options {
		option(resource)
	}

	return resource, nil
}

func (r *OTelMetricResource) Execute() {
	r.logger.Debug("pinging", zap.String("host", r.host), zap.String("port", strconv.Itoa(int(r.port))))
	startTime := time.Now()
	info, err := r.pinger.Ping()
	elapsed := time.Now().Sub(startTime)
	if err != nil || info.Players.Max == 0 {
		Metrics().RecordHealth(false, getOTelMetricAttributes(r.host, r.port, r.edition, ""))
		return
	}

	Metrics().RecordResponseTime(elapsed.Seconds(), getOTelMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	Metrics().RecordHealth(true, getOTelMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	Metrics().RecordPlayersOnlineCount(float64(info.Players.Online), getOTelMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
	Metrics().RecordPlayersMaxCount(float64(info.Players.Max), getOTelMetricAttributes(r.host, r.port, r.edition, info.Version.Name))
}
