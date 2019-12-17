package main

import (
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/go-mc-status"
	"log"
)

var config struct {
	Address string `default:"localhost"`
	Port    int    `default:"25565"`
}

func main() {
	err := flagsfiller.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	client := mcstatus.NewClient(config.Address, uint16(config.Port))
	err = client.Connect()
	if err != nil {
		log.Fatal(err)
	}

	state, err := client.Handshake()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n", state)
}
