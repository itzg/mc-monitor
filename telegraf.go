package main

import (
	"bytes"
	mcpinger "github.com/Raqbit/mc-pinger"
	protocol "github.com/influxdata/line-protocol"
	"go.uber.org/zap"
	"log"
	"net"
	"strconv"
	"time"
)

type TelegrafGatherer struct {
	host             string
	port             string
	telegrafEndpoint string
	pinger           mcpinger.Pinger
	logger           *zap.Logger
}

func NewTelegrafGatherer(host string, port int, telegrafEndpoint string, logger *zap.Logger) *TelegrafGatherer {
	return &TelegrafGatherer{
		host:             host,
		port:             strconv.FormatInt(int64(port), 10),
		pinger:           mcpinger.New(host, uint16(port)),
		telegrafEndpoint: telegrafEndpoint,
		logger:           logger,
	}
}

func (g *TelegrafGatherer) Gather() {
	g.logger.Debug("gathering", zap.String("host", g.host), zap.String("port", g.port))
	startTime := time.Now()
	info, err := g.pinger.Ping()
	elapsed := time.Now().Sub(startTime)

	if err != nil {
		err := g.sendFailedMetrics(err, elapsed)
		if err != nil {
			log.Printf("failed to send metrics: %s", err)
		}
	} else {
		err := g.sendInfoMetrics(info, elapsed)
		if err != nil {
			log.Printf("failed to send metrics: %s", err)
		}
	}
}

func (g *TelegrafGatherer) sendInfoMetrics(info *mcpinger.ServerInfo, elapsed time.Duration) error {
	m := NewSimpleMetric(MetricName)

	m.AddTag(TagHost, g.host)
	m.AddTag(TagPort, g.port)
	m.AddTag(TagStatus, StatusSuccess)

	m.AddField(FieldResponseTime, elapsed.Seconds())
	m.AddField(FieldOnline, uint64(info.Players.Online))
	m.AddField(FieldMax, uint64(info.Players.Max))

	var buf bytes.Buffer
	encoder := protocol.NewEncoder(&buf)
	_, err := encoder.Encode(m)
	if err != nil {
		return err
	}

	err = g.sendLine(buf.Bytes())
	if err != nil {
		return nil
	}

	return nil
}

func (g *TelegrafGatherer) sendFailedMetrics(err error, elapsed time.Duration) error {
	m := NewSimpleMetric(MetricName)

	m.AddTag(TagHost, g.host)
	m.AddTag(TagPort, g.port)
	m.AddTag(TagStatus, StatusError)

	m.AddField(FieldError, err.Error())
	m.AddField(FieldResponseTime, elapsed.Seconds())

	var buf bytes.Buffer
	encoder := protocol.NewEncoder(&buf)
	_, err = encoder.Encode(m)
	if err != nil {
		return err
	}

	err = g.sendLine(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (g *TelegrafGatherer) sendLine(lineBytes []byte) error {
	g.logger.Debug("sending metrics",
		zap.String("endpoint", g.telegrafEndpoint),
		zap.ByteString("line", lineBytes))
	conn, err := net.Dial("tcp", g.telegrafEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Printf("failed to close line protocol connection: %s", closeErr)
		}
	}()

	_, err = conn.Write(lineBytes)
	if err != nil {
		return err
	}

	return nil
}

type SimpleMetric struct {
	name   string
	tags   []*protocol.Tag
	fields []*protocol.Field
}

func NewSimpleMetric(name string) *SimpleMetric {
	return &SimpleMetric{name: name}
}

func (m *SimpleMetric) Time() time.Time {
	return time.Now()
}

func (m *SimpleMetric) Name() string {
	return m.name
}

func (m *SimpleMetric) TagList() []*protocol.Tag {
	return m.tags
}

func (m *SimpleMetric) FieldList() []*protocol.Field {
	return m.fields
}

func (m *SimpleMetric) AddTag(key, value string) {
	m.tags = append(m.tags, &protocol.Tag{
		Key:   key,
		Value: value,
	})
}

func (m *SimpleMetric) AddField(key string, value interface{}) {
	m.fields = append(m.fields, &protocol.Field{
		Key:   key,
		Value: value,
	})
}
