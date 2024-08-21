package otel

import (
	"strconv"

	"github.com/itzg/mc-monitor/utils"
	"go.opentelemetry.io/otel/attribute"
)

const (
	serverHostAttribute    = "server_host"
	serverPortAttribute    = "server_port"
	serverEditionAttribute = "server_edition"
	serverVersionAttribute = "server_version"
)

type serverMetrics struct {
}

func Metrics() *serverMetrics {
	return &serverMetrics{}
}

func (m *serverMetrics) RecordHealth(healthy bool, attributes []attribute.KeyValue) {
	_ = NewInt64ObservableGauge(
		"minecraft_status_healthy",
		"Indicates if the server is healthy (1) or not (0)",
		func() int64 {
			if healthy {
				return int64(1)
			}
			return int64(0)
		},
		attributes,
	)
}

func (m *serverMetrics) RecordResponseTime(responseTime float64, attributes []attribute.KeyValue) {
	_ = NewFloat64ObservableGauge(
		"minecraft_status_response_time",
		"The response time of the server",
		func() float64 {
			return responseTime
		},
		attributes,
	)
}

func (m *serverMetrics) RecordPlayersOnlineCount(playersOnlineCount int32, attributes []attribute.KeyValue) {
	_ = NewInt64ObservableGauge(
		"minecraft_players_online_count",
		"The number of players currently online on the server",
		func() int64 {
			return int64(playersOnlineCount)
		},
		attributes,
	)
}

func (m *serverMetrics) RecordPlayersMaxCount(playersMaxCount int32, attributes []attribute.KeyValue) {
	_ = NewInt64ObservableGauge(
		"minecraft_players_max_count",
		"The maximum number of players that can be online on the server",
		func() int64 {
			return int64(playersMaxCount)
		},
		attributes,
	)
}

func buildMetricAttributes(host string, port uint16, edition utils.ServerEdition, version string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(serverHostAttribute, host),
		attribute.String(serverPortAttribute, strconv.Itoa(int(port))),
		attribute.String(serverEditionAttribute, string(edition)),
		attribute.String(serverVersionAttribute, version),
	}
}
