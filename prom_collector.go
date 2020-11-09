package main

import (
	"fmt"
	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"net"
	"strconv"
	"time"
)

const (
	promLabelHost    = "server_host"
	promLabelPort    = "server_port"
	promLabelEdition = "server_edition"
	promLabelVersion = "server_version"
)

var (
	promVariableLabels = []string{promLabelHost, promLabelPort, promLabelEdition, promLabelVersion}
	promDescHealthy    = prometheus.NewDesc("minecraft_status_healthy",
		"Indicates if the server is healthy (1) or not (0)",
		promVariableLabels, nil)
	promDescResponseTime = prometheus.NewDesc("minecraft_status_response_time_seconds",
		"Amount of time it took for server to respond",
		promVariableLabels, nil)
	promDescPlayersOnline = prometheus.NewDesc("minecraft_status_players_online_count",
		"Number of players currently online",
		promVariableLabels, nil)
	promDescPlayersMax = prometheus.NewDesc("minecraft_status_players_max_count",
		"Maximum number of players allowed by the server",
		promVariableLabels, nil)
)

type specificPromCollector interface {
	Collect(metrics chan<- prometheus.Metric)
}

type promCollectors []specificPromCollector

func (promCollectors) Describe(descs chan<- *prometheus.Desc) {
	descs <- promDescHealthy
	descs <- promDescResponseTime
	descs <- promDescPlayersOnline
	descs <- promDescPlayersMax
}

func (c promCollectors) Collect(metrics chan<- prometheus.Metric) {
	for _, entry := range c {
		entry.Collect(metrics)
	}
}

func newPromCollectors(servers []string, bedrockServers []string, logger *zap.Logger) (promCollectors, error) {
	var collectors []specificPromCollector

	javaCollectors, err := createPromCollectors(servers, JavaEdition, logger)
	if err != nil {
		return nil, err
	}
	collectors = append(collectors, javaCollectors...)

	bedrockCollectors, err := createPromCollectors(bedrockServers, BedrockEdition, logger)
	if err != nil {
		return nil, err
	}
	collectors = append(collectors, bedrockCollectors...)

	return collectors, nil
}

func createPromCollectors(servers []string, edition ServerEdition, logger *zap.Logger) (collectors []specificPromCollector, err error) {
	for _, server := range servers {
		switch edition {

		case JavaEdition:
			host, port, err := SplitHostPort(server, DefaultJavaPort)
			if err != nil {
				return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
			}
			collectors = append(collectors, newPromJavaCollector(host, port, logger))

		case BedrockEdition:
			host, port, err := SplitHostPort(server, DefaultBedrockPort)
			if err != nil {
				return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
			}
			collectors = append(collectors, newPromBedrockCollector(host, port, logger))
		}
	}
	return
}

func newPromJavaCollector(host string, port uint16, logger *zap.Logger) specificPromCollector {
	return &promJavaCollector{
		host:   host,
		port:   strconv.Itoa(int(port)),
		pinger: mcpinger.New(host, port),
		logger: logger,
	}
}

type promJavaCollector struct {
	host   string
	port   string
	pinger mcpinger.Pinger
	logger *zap.Logger
}

func (c *promJavaCollector) Collect(metrics chan<- prometheus.Metric) {
	c.logger.Debug("pinging", zap.String("host", c.host), zap.String("port", c.port))
	startTime := time.Now()
	info, err := c.pinger.Ping()
	elapsed := time.Now().Sub(startTime)

	if err != nil {
		c.sendMetric(metrics, promDescHealthy, "", 0)
	} else {
		c.sendMetric(metrics, promDescResponseTime, info.Version.Name, elapsed.Seconds())
		if info.Players.Max == 0 { // when server responds to ping but is not fully ready
			c.sendMetric(metrics, promDescHealthy, info.Version.Name, 0)
		} else {
			c.sendMetric(metrics, promDescHealthy, info.Version.Name, 1)
			c.sendMetric(metrics, promDescPlayersOnline, info.Version.Name, float64(info.Players.Online))
			c.sendMetric(metrics, promDescPlayersMax, info.Version.Name, float64(info.Players.Max))
		}

	}
}

func (c *promJavaCollector) sendMetric(metrics chan<- prometheus.Metric, desc *prometheus.Desc,
	version string, value float64) {

	metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, value,
		c.host, c.port, string(JavaEdition), version)
	if err != nil {
		c.logger.Error("failed to build metric", zap.Error(err), zap.String("name", desc.String()))
	} else {
		metrics <- metric
	}
}

type promBedrockCollector struct {
	host   string
	port   string
	logger *zap.Logger
}

func newPromBedrockCollector(host string, port uint16, logger *zap.Logger) *promBedrockCollector {
	return &promBedrockCollector{host: host, port: strconv.Itoa(int(port)), logger: logger}
}

func (c *promBedrockCollector) Collect(metrics chan<- prometheus.Metric) {
	c.logger.Debug("pinging", zap.String("host", c.host), zap.String("port", c.port))

	info, err := PingBedrockServer(net.JoinHostPort(c.host, c.port))
	if err != nil {
		c.sendMetric(metrics, promDescHealthy, "", 0)
	} else {
		c.sendMetric(metrics, promDescResponseTime, info.Version, info.Rtt.Seconds())
		c.sendMetric(metrics, promDescHealthy, info.Version, 1)
		c.sendMetric(metrics, promDescPlayersOnline, info.Version, float64(info.Players))
		c.sendMetric(metrics, promDescPlayersMax, info.Version, float64(info.MaxPlayers))
	}
}

func (c *promBedrockCollector) sendMetric(metrics chan<- prometheus.Metric,
	desc *prometheus.Desc, version string, value float64) {

	metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, value,
		c.host, c.port, string(BedrockEdition), version)
	if err != nil {
		c.logger.Error("failed to build metric", zap.Error(err), zap.String("name", desc.String()))
	} else {
		metrics <- metric
	}
}
