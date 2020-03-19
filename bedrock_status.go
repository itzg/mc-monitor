package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"log"
	"net"
	"strconv"
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

func (c *statusBedrockCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	address := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	info, err := PingBedrockServer(address)
	if err != nil {
		log.Fatal(err)
		return subcommands.ExitFailure
	}

	fmt.Printf("%s : version=%s online=%d max=%d",
		address,
		info.Version, info.Players, info.MaxPlayers)

	return subcommands.ExitSuccess
}
