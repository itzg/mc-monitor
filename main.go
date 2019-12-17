package main

import (
	"fmt"
	"github.com/Raqbit/mc-pinger"
	"github.com/itzg/go-flagsfiller"
	"log"
	"time"
)

var config struct {
	Host   string `default:"localhost" usage:"hostname of the Minecraft server"`
	Port   int    `default:"25565" usage:"port of the Minecraft server"`
	Gather struct {
		Interval        time.Duration `default:"1m" usage:"when gather endpoint configured, gathers and sends at this interval"`
		TelegrafAddress string        `usage:"[host:port] of telegraf accepting Influx line protocol"`
	}
}

const (
	MetricName = "minecraft_status"

	TagHost   = "host"
	TagPort   = "port"
	TagStatus = "status"

	FieldError        = "error"
	FieldOnline       = "online"
	FieldResponseTime = "response_time"

	StatusError   = "error"
	StatusSuccess = "success"
)

func main() {
	err := flagsfiller.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	pinger := mcpinger.New(config.Host, uint16(config.Port))

	if config.Gather.TelegrafAddress != "" {
		gatherer := NewTelegrafGatherer(config.Host, config.Port, config.Gather.TelegrafAddress)
		gatherer.Start(pinger, config.Gather.Interval)
	} else {
		// one shot
		info, err := pinger.Ping()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("version=%s online=%d max=%d motd='%s'",
			info.Version.Name, info.Players.Online, info.Players.Max, info.Description.Text)
	}
}
