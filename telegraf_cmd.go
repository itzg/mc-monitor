package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	lpsender "github.com/itzg/line-protocol-sender"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

type gatherTelegrafCmd struct {
	Interval        time.Duration `default:"1m" usage:"gathers and sends metrics at this interval"`
	Servers         []string      `usage:"one or more [host:port] addresses of servers to monitor"`
	TelegrafAddress string        `default:"localhost:8094" usage:"[host:port] of telegraf accepting Influx line protocol"`
	logger          *zap.Logger
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

func (c *gatherTelegrafCmd) SetFlags(f *flag.FlagSet) {
	filler := flagsfiller.New(flagsfiller.WithEnv("Gather"))
	err := filler.Fill(f, c)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *gatherTelegrafCmd) Execute(ctx context.Context, _ *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {

	if len(c.Servers) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "requires at least one server")
		return subcommands.ExitUsageError
	}

	if c.TelegrafAddress == "" {
		_, _ = fmt.Fprintln(os.Stderr, "requires TelegrafAddress")
		return subcommands.ExitUsageError
	}

	c.logger = args[0].(*zap.Logger).Named("gather")

	c.logger.Info("starting monitoring",
		zap.Strings("servers", c.Servers),
		zap.Duration("interval", c.Interval),
		zap.String("telegrafAddress", c.TelegrafAddress))

	ticker := time.NewTicker(c.Interval)

	gatherers, err := c.createGatherers()
	if err != nil {
		c.logger.Error("failed to setup gatherers", zap.Error(err))
		return subcommands.ExitFailure
	}

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

func (c *gatherTelegrafCmd) createGatherers() ([]*TelegrafGatherer, error) {
	gatherers := make([]*TelegrafGatherer, 0, len(c.Servers))

	lpClient, err := lpsender.NewClient(context.Background(), lpsender.Config{
		Endpoint:  c.TelegrafAddress,
		BatchSize: len(c.Servers),
		ErrorListener: func(err error) {
			c.logger.Error("failed to send metrics", zap.Error(err))
		},
	})
	if err != nil {
		return nil, err
	}

	for _, addr := range c.Servers {
		host, port, err := SplitHostPort(addr, DefaultJavaPort)
		if err != nil {
			return nil, err
		}
		gatherers = append(gatherers, NewTelegrafGatherer(host, port, lpClient, c.logger))
	}

	return gatherers, nil
}
