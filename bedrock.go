package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"github.com/sandertv/go-raknet"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type statusBedrockCmd struct {
	Host string `default:"localhost"`
	Port int    `default:"19132"`
}

func (c *statusBedrockCmd) Name() string {
	return "status-bedrock"
}

func (c *statusBedrockCmd) Synopsis() string {
	return "Retrieves and displays the status of the given Minecraft Bedrock Dedicated server"
}

func (c *statusBedrockCmd) Usage() string {
	return ""
}

func (c *statusBedrockCmd) SetFlags(flags *flag.FlagSet) {
	filler := flagsfiller.New()
	err := filler.Fill(flags, c)
	if err != nil {
		log.Fatal(err)
	}
}

type bedrockPong struct {
	ServerName      string
	ProtocolVersion string
	Version         string
	Players         int
	MaxPlayers      int
	LevelName       string
	GameMode        string
	Difficulty      string
	Rtt             time.Duration
}

func (c *statusBedrockCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	address := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	start := time.Now()
	response, err := raknet.Ping(address)
	rtt := time.Now().Sub(start)
	if err != nil {
		log.Printf("ERR: failed to query bedrock server %s: %s", address, err.Error())
		return subcommands.ExitFailure
	}

	parts := strings.Split(string(response), ";")
	var info = &bedrockPong{
		Rtt:             rtt,
		ServerName:      parts[1],
		ProtocolVersion: parts[2],
		Version:         parts[3],
		Players:         safeParseInt(parts[4]),
		MaxPlayers:      safeParseInt(parts[5]),
		LevelName:       parts[7],
		GameMode:        parts[8],
	}
	if len(parts) >= 10 {
		info.Difficulty = parts[9]
	}

	fmt.Printf("%s : version=%s online=%d max=%d",
		address,
		info.Version, info.Players, info.MaxPlayers)

	return subcommands.ExitSuccess
}

func safeParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	} else {
		return i
	}
}
