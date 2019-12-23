package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type gatherTelegrafCmd struct {
	Interval        time.Duration `default:"1m" usage:"gathers and sends metrics at this interval"`
	Servers         []string      `usage:"one or more [host:port] addresses of servers to monitor"`
	TelegrafAddress string        `default:"localhost:8094" usage:"[host:port] of telegraf accepting Influx line protocol"`
}

func (c *gatherTelegrafCmd) Name() string {
	return "gather-for-telegraf"
}

func (c *gatherTelegrafCmd) Synopsis() string {
	return "Periodically gathers to status of one or more Minecraft servers and sends metrics to telegraf over TCP using Influx line protocol"
}

func (c *gatherTelegrafCmd) Usage() string {
	return ""
}

func (c *gatherTelegrafCmd) SetFlags(flags *flag.FlagSet) {
	filler := flagsfiller.New()
	err := filler.Fill(flags, c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *gatherTelegrafCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {

	if len(c.Servers) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "requires at least one server")
		return subcommands.ExitUsageError
	}

	if c.TelegrafAddress == "" {
		_, _ = fmt.Fprintln(os.Stderr, "requires TelegrafAddress")
		return subcommands.ExitUsageError
	}

	fmt.Printf("monitoring %v at interval %s and reporting to %s\n", c.Servers, c.Interval, c.TelegrafAddress)

	ticker := time.NewTicker(c.Interval)

	gatherers := c.createGatherers()

	for {
		select {
		case <-ctx.Done():
			return subcommands.ExitSuccess

		case <-ticker.C:
			for _, gatherer := range gatherers {
				gatherer.Gather()
			}
		}
	}
}

func (c *gatherTelegrafCmd) createGatherers() []*TelegrafGatherer {
	gatherers := make([]*TelegrafGatherer, 0, len(c.Servers))

	for _, addr := range c.Servers {
		parts := strings.SplitN(addr, ":", 2)
		if len(parts) == 2 {
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				log.Printf("WARN: unable to process %s: %s\n", addr, err)
			} else {
				gatherers = append(gatherers, NewTelegrafGatherer(parts[0], port, c.TelegrafAddress))
			}
		} else {
			gatherers = append(gatherers, NewTelegrafGatherer(parts[0], DefaultPort, c.TelegrafAddress))
		}
	}

	return gatherers
}
