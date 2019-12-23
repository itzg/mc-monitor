package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Raqbit/mc-pinger"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"log"
	"os"
)

type statusCmd struct {
	Host string `default:"localhost" usage:"hostname of the Minecraft server" env:"MC_HOST"`
	Port int    `default:"25565" usage:"port of the Minecraft server" env:"MC_PORT"`
}

func (c *statusCmd) Name() string {
	return "status"
}

func (c *statusCmd) Synopsis() string {
	return "Retrieves and displays the status of the given Minecraft server"
}

func (c *statusCmd) Usage() string {
	return ""
}

func (c *statusCmd) SetFlags(flags *flag.FlagSet) {
	filler := flagsfiller.New()
	err := filler.Fill(flags, c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *statusCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	pinger := mcpinger.New(c.Host, uint16(c.Port))

	info, err := pinger.Ping()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to ping %s:%d : %s", c.Host, c.Port, err)
		return subcommands.ExitFailure
	}

	fmt.Printf("%s:%d : version=%s online=%d max=%d motd='%s'",
		c.Host, c.Port,
		info.Version.Name, info.Players.Online, info.Players.Max, info.Description.Text)

	return subcommands.ExitSuccess
}
