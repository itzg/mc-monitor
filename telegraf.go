package main

import (
	"errors"
	mcpinger "github.com/Raqbit/mc-pinger"
	lpsender "github.com/itzg/line-protocol-sender"
	"go.uber.org/zap"
	"log"
	"strconv"
	"time"
)

const (
	MetricName = "minecraft_status"

	TagHost    = "host"
	TagPort    = "port"
	TagStatus  = "status"
	TagVersion = "version"

	FieldError        = "error"
	FieldOnline       = "online"
	FieldMax          = "max"
	FieldResponseTime = "response_time"

	StatusError   = "error"
	StatusSuccess = "success"
)

type TelegrafGatherer struct {
	host     string
	port     string
	pinger   mcpinger.Pinger
	logger   *zap.Logger
	lpClient lpsender.Client
}

func NewTelegrafGatherer(host string, port uint16, lpClient lpsender.Client, logger *zap.Logger) *TelegrafGatherer {
	return &TelegrafGatherer{
		host:     host,
		port:     strconv.FormatInt(int64(port), 10),
		pinger:   mcpinger.New(host, uint16(port)),
		lpClient: lpClient,
		logger:   logger,
	}
}

func (g *TelegrafGatherer) Gather() {
	g.logger.Debug("gathering", zap.String("host", g.host), zap.String("port", g.port))
	startTime := time.Now()
	info, err := g.pinger.Ping()
	elapsed := time.Now().Sub(startTime)

	if err != nil {
		g.sendFailedMetrics(err, elapsed)
	} else if info.Players.Max == 0 {
		g.sendFailedMetrics(errors.New("server not ready"), elapsed)
	} else {
		err := g.sendInfoMetrics(info, elapsed)
		if err != nil {
			log.Printf("failed to send metrics: %s", err)
		}
	}
}

func (g *TelegrafGatherer) sendInfoMetrics(info *mcpinger.ServerInfo, elapsed time.Duration) error {
	m := lpsender.NewSimpleMetric(MetricName)

	m.AddTag(TagHost, g.host)
	m.AddTag(TagPort, g.port)
	m.AddTag(TagStatus, StatusSuccess)
	m.AddTag(TagVersion, info.Version.Name)

	m.AddField(FieldResponseTime, elapsed.Seconds())
	m.AddField(FieldOnline, uint64(info.Players.Online))
	m.AddField(FieldMax, uint64(info.Players.Max))

	g.lpClient.Send(m)

	return nil
}

func (g *TelegrafGatherer) sendFailedMetrics(err error, elapsed time.Duration) {
	m := lpsender.NewSimpleMetric(MetricName)

	m.AddTag(TagHost, g.host)
	m.AddTag(TagPort, g.port)
	m.AddTag(TagStatus, StatusError)

	m.AddField(FieldError, err.Error())
	m.AddField(FieldResponseTime, elapsed.Seconds())

	g.lpClient.Send(m)

	return
}
