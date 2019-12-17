package main

import (
	"fmt"
	"github.com/Raqbit/mc-pinger"
	"github.com/itzg/go-flagsfiller"
	"log"
)

var config struct {
	Host string `default:"localhost"`
	Port int    `default:"25565"`
}

func main() {
	err := flagsfiller.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	pinger := mcpinger.New(config.Host, uint16(config.Port))

	info, err := pinger.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("version=%s online=%d max=%d motd='%s'",
		info.Version.Name, info.Players.Online, info.Players.Max, info.Description.Text)
}
