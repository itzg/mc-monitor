package main

import (
	"fmt"
	"net"
	"strconv"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/itzg/mc-monitor/slp"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
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

type pingOptions interface {
	GetHost() string
	GetPort() uint16
	GetTimeout() time.Duration
	GetUseProxy() bool
	GetProxyVersion() byte
}

func javaPingerOptions(opt pingOptions) []mcpinger.McPingerOption {
	var opts []mcpinger.McPingerOption
	if t := opt.GetTimeout(); t > 0 {
		opts = append(opts, mcpinger.WithTimeout(t))
	}
	if opt.GetUseProxy() {
		opts = append(opts, mcpinger.WithProxyProto(opt.GetProxyVersion()))
	}
	return opts
}

func pingJavaServer(opt pingOptions) (*mcpinger.ServerInfo, error) {
	opts := javaPingerOptions(opt)
	pinger := mcpinger.New(opt.GetHost(), opt.GetPort(), opts...)
	return pinger.Ping()
}

// pingJavaServerList uses the legacy server list ping of 1.4 to 1.6
func pingJavaServerList(opt pingOptions) (*mcpinger.ServerInfo, error) {
	response, err := slp.ServerListPing(opt.GetHost(), int(opt.GetPort()), opt.GetTimeout())
	if err != nil {
		return nil, err
	}
	return legacyServerInfo(response.ServerVersion, response.CurrentPlayerCount, response.MaxPlayers)
}

// pingJavaOldServerList uses the old server list ping of b1.8 to 1.3, which doesn't report the server version
func pingJavaOldServerList(opt pingOptions) (*mcpinger.ServerInfo, error) {
	response, err := slp.OldServerListPing(opt.GetHost(), int(opt.GetPort()), opt.GetTimeout())
	if err != nil {
		return nil, err
	}
	return legacyServerInfo("", response.CurrentPlayerCount, response.MaxPlayers)
}

// legacyServerInfo maps the fields of a legacy server list response onto the ones used from a status response
func legacyServerInfo(version string, playersOnline string, playersMax string) (*mcpinger.ServerInfo, error) {
	online, err := strconv.ParseInt(playersOnline, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid online player count: %w", err)
	}
	maxPlayers, err := strconv.ParseInt(playersMax, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid max player count: %w", err)
	}

	var info mcpinger.ServerInfo
	info.Version.Name = version
	info.Players.Online = int32(online)
	info.Players.Max = int32(maxPlayers)
	return &info, nil
}

type specificPromCollector interface {
	Collect(metrics chan<- prometheus.Metric)
	SetTimeout(t time.Duration)
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

func newPromCollectors(servers []string, bedrockOptions BedrockOptions, useProxy bool, proxyVersion uint, logger *zap.Logger) (promCollectors, error) {
	var collectors []specificPromCollector

	if useProxy && proxyVersion != 1 && proxyVersion != 2 {
		return nil, fmt.Errorf("proxy version must be 1 or 2")
	}

	javaCollectors, err := createJavaPromCollectors(servers, useProxy, byte(proxyVersion), logger)
	if err != nil {
		return nil, err
	}
	collectors = append(collectors, javaCollectors...)

	bedrockCollectors, err := createBedrockPromCollectors(bedrockOptions, logger)
	if err != nil {
		return nil, err
	}
	collectors = append(collectors, bedrockCollectors...)

	return collectors, nil
}

func createJavaPromCollectors(servers []string, useProxy bool, proxyVersion byte, logger *zap.Logger) (collectors []specificPromCollector, err error) {
	for _, server := range servers {
		host, port, err := SplitHostPort(server, DefaultJavaPort)
		if err != nil {
			return nil, fmt.Errorf("failed to process server entry '%s': %w", server, err)
		}
		collectors = append(collectors, newPromJavaCollector(host, port, useProxy, proxyVersion, logger))
	}
	return
}

func createBedrockPromCollectors(bedrockOptions BedrockOptions, logger *zap.Logger) (collectors []specificPromCollector, err error) {
	for _, server := range bedrockOptions.Servers {
		host, port, err := SplitHostPort(server, DefaultBedrockPort)
		if err != nil {
			return nil, fmt.Errorf("failed to process bedrock server entry '%s': %w", server, err)
		}
		collectors = append(collectors, newPromBedrockCollector(
			host,
			port,
			bedrockOptions.Protocol,
			bedrockOptions.NethernetMode,
			logger,
		))
	}
	return
}

// useLegacyPing switches the Java collectors from the status protocol to one of the legacy server list pings
func (c promCollectors) useLegacyPing(useServerListPing bool, useOldServerListPing bool) error {
	if !useServerListPing && !useOldServerListPing {
		return nil
	}
	if useServerListPing && useOldServerListPing {
		return fmt.Errorf("only one of server list ping or old server list ping can be used")
	}

	for _, entry := range c {
		javaCollector, ok := entry.(*promJavaCollector)
		if !ok {
			continue
		}
		if javaCollector.useProxy {
			return fmt.Errorf("proxy protocol is not supported with server list ping")
		}
		javaCollector.useServerListPing = useServerListPing
		javaCollector.useOldServerListPing = useOldServerListPing
	}
	return nil
}

func newPromJavaCollector(host string, port uint16, useProxy bool, proxyVersion byte, logger *zap.Logger) specificPromCollector {
	return &promJavaCollector{
		host:         host,
		port:         port,
		logger:       logger,
		useProxy:     useProxy,
		proxyVersion: proxyVersion,
	}
}

type promJavaCollector struct {
	host         string
	port         uint16
	logger       *zap.Logger
	timeout      time.Duration
	useProxy     bool
	proxyVersion byte

	useServerListPing    bool
	useOldServerListPing bool
}

func (c *promJavaCollector) GetHost() string {
	return c.host
}

func (c *promJavaCollector) GetPort() uint16 {
	return c.port
}

func (c *promJavaCollector) GetTimeout() time.Duration {
	return c.timeout
}

func (c *promJavaCollector) GetUseProxy() bool {
	return c.useProxy
}

func (c *promJavaCollector) GetProxyVersion() byte {
	return c.proxyVersion
}

func (c *promJavaCollector) SetTimeout(t time.Duration) {
	c.timeout = t
}

func (c *promJavaCollector) Collect(metrics chan<- prometheus.Metric) {
	c.logger.Debug("pinging", zap.String("host", c.host), zap.String("port", strconv.Itoa(int(c.port))))
	startTime := time.Now()
	info, err := c.ping()
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

func (c *promJavaCollector) ping() (*mcpinger.ServerInfo, error) {
	switch {
	case c.useServerListPing:
		return pingJavaServerList(c)
	case c.useOldServerListPing:
		return pingJavaOldServerList(c)
	default:
		return pingJavaServer(c)
	}
}

func (c *promJavaCollector) sendMetric(metrics chan<- prometheus.Metric, desc *prometheus.Desc,
	version string, value float64) {

	metric, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, value,
		c.host, strconv.Itoa(int(c.port)), string(JavaEdition), version)
	if err != nil {
		c.logger.Error("failed to build metric", zap.Error(err), zap.String("name", desc.String()))
	} else {
		metrics <- metric
	}
}

type promBedrockCollector struct {
	host     string
	port     uint16
	protocol string
	mode     string
	logger   *zap.Logger
	timeout  time.Duration
}

func (c *promBedrockCollector) GetHost() string {
	return c.host
}

func (c *promBedrockCollector) GetPort() uint16 {
	return c.port
}

func (c *promBedrockCollector) GetTimeout() time.Duration {
	return c.timeout
}

func (c *promBedrockCollector) SetTimeout(t time.Duration) {
	c.timeout = t
}

func newPromBedrockCollector(host string, port uint16, protocol string, mode string, logger *zap.Logger) *promBedrockCollector {
	return &promBedrockCollector{host: host, port: port, protocol: protocol, mode: mode, logger: logger}
}

func (c *promBedrockCollector) Collect(metrics chan<- prometheus.Metric) {
	address := net.JoinHostPort(c.host, strconv.Itoa(int(c.port)))
	c.logger.Debug("collecting", zap.String("address", address))

	info, err := PingBedrockServer(address, c.mode, c.protocol, c.timeout, c.logger)
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
		c.host, strconv.Itoa(int(c.port)), string(BedrockEdition), version)
	if err != nil {
		c.logger.Error("failed to build metric", zap.Error(err), zap.String("name", desc.String()))
	} else {
		metrics <- metric
	}
}
