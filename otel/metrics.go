package otel

import (
	"strconv"

	"github.com/itzg/mc-monitor/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

const (
	serverHostAttribute    = "server_host"
	serverPortAttribute    = "server_port"
	serverEditionAttribute = "server_edition"
	serverVersionAttribute = "server_version"
)

type ServerMetrics struct {
	healthy            bool
	responseTime       float64
	playersOnlineCount int64
	playersMaxCount    int64
	logger             *zap.Logger
}

func NewServerMetrics(logger *zap.Logger) *ServerMetrics {
	return &ServerMetrics{
		healthy:            false,
		responseTime:       0.0,
		playersOnlineCount: 0,
		playersMaxCount:    0,
		logger:             logger,
	}
}

func (m *ServerMetrics) RecordHealth(healthy bool, attributes []attribute.KeyValue) {
	m.healthy = healthy
	NewInt64ObservableGauge(
		"minecraft_status_healthy",
		"Indicates if the server is healthy (1) or not (0)",
		func() int64 {
			if m.healthy {
				m.logger.Debug("Server is healthy")
				return int64(1)
			}
			m.logger.Debug("Server is not healthy")
			return int64(0)
		},
		attributes,
	)
}

func (m *ServerMetrics) RecordResponseTime(responseTime float64, attributes []attribute.KeyValue) {
	m.responseTime = responseTime
	NewFloat64ObservableGauge(
		"minecraft_status_response_time",
		"The response time of the server",
		func() float64 {
			m.logger.Debug("Response time", zap.Float64("responseTime", m.responseTime))
			return m.responseTime
		},
		attributes,
	)
}

func (m *ServerMetrics) RecordPlayersOnlineCount(playersOnlineCount int32, attributes []attribute.KeyValue) {
	m.playersOnlineCount = int64(playersOnlineCount)
	NewInt64ObservableGauge(
		"minecraft_status_players_online_count",
		"The number of players currently online on the server",
		func() int64 {
			m.logger.Debug("PlayersOnlineCount", zap.Int64("playersOnlineCount", m.playersOnlineCount))
			return m.playersOnlineCount
		},
		attributes,
	)
}

func (m *ServerMetrics) RecordPlayersMaxCount(playersMaxCount int32, attributes []attribute.KeyValue) {
	m.playersMaxCount = int64(playersMaxCount)
	NewInt64ObservableGauge(
		"minecraft_status_players_max_count",
		"The maximum number of players that can be online on the server",
		func() int64 {
			m.logger.Debug("PlayersMaxCount", zap.Int64("playersMaxCount", m.playersMaxCount))
			return m.playersMaxCount
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
