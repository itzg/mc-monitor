package otel

import (
	"context"
	"strconv"

	"github.com/itzg/mc-monitor/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serverHostAttribute    = "server_host"
	serverPortAttribute    = "server_port"
	serverEditionAttribute = "server_edition"
	serverVersionAttribute = "server_version"
)

var (
	meter = otel.GetMeterProvider().Meter("minecraft")
)

type serverMetrics struct {
}

func Metrics() *serverMetrics {
	return &serverMetrics{}
}

func (m *serverMetrics) RecordHealth(healthy bool, attributes []attribute.KeyValue) {
	_, err := meter.Int64ObservableGauge(
		"minecraft_status_healthy",
		metric.WithDescription("Indicates if the server is healthy (1) or not (0)"),
		metric.WithUnit("1"),
		metric.WithInt64Callback(
			func(ctx context.Context, observer metric.Int64Observer) error {
				if healthy {
					observer.Observe(1, metric.WithAttributes(attributes...))
				} else {
					observer.Observe(0, metric.WithAttributes(attributes...))
				}
				return nil
			},
		),
	)
	handleError("Error creating healthy metric", err)
}

func (m *serverMetrics) RecordResponseTime(responseTime float64, attributes []attribute.KeyValue) {
	_, err := meter.Float64ObservableGauge(
		"minecraft_status_response_time",
		metric.WithDescription("The response time of the server"),
		metric.WithUnit("ms"),
		metric.WithFloat64Callback(
			func(ctx context.Context, observer metric.Float64Observer) error {
				observer.Observe(responseTime, metric.WithAttributes(attributes...))
				return nil
			},
		),
	)
	handleError("Error creating response time metric", err)
}

func (m *serverMetrics) RecordPlayersOnlineCount(playersOnlineCount float64, attributes []attribute.KeyValue) {
	_, err := meter.Float64ObservableGauge(
		"minecraft_players_online_count",
		metric.WithDescription("The number of players currently online on the server"),
		metric.WithUnit("1"),
		metric.WithFloat64Callback(
			func(ctx context.Context, observer metric.Float64Observer) error {
				observer.Observe(playersOnlineCount, metric.WithAttributes(attributes...))
				return nil
			},
		),
	)
	handleError("Error creating players online count metric", err)
}

func (m *serverMetrics) RecordPlayersMaxCount(playersMaxCount float64, attributes []attribute.KeyValue) {
	_, err := meter.Float64ObservableGauge(
		"minecraft_players_max_count",
		metric.WithDescription("The maximum number of players that can be online on the server"),
		metric.WithUnit("1"),
		metric.WithFloat64Callback(
			func(ctx context.Context, observer metric.Float64Observer) error {
				observer.Observe(playersMaxCount, metric.WithAttributes(attributes...))
				return nil
			},
		),
	)
	handleError("Error creating players max count metric", err)
}

func handleError(msg string, err error) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}

func getOTelMetricAttributes(host string, port uint16, edition utils.ServerEdition, version string) []attribute.KeyValue {
	var attributes = make([]attribute.KeyValue, 0)

	if host != "" {
		attributes = append(attributes, attribute.String(serverHostAttribute, host))
	}

	if port != 0 {
		attributes = append(attributes, attribute.String(serverPortAttribute, strconv.Itoa(int(port))))
	}

	if edition != "" {
		attributes = append(attributes, attribute.String(serverEditionAttribute, string(edition)))
	}

	if version != "" {
		attributes = append(attributes, attribute.String(serverVersionAttribute, version))
	}

	return attributes
}
