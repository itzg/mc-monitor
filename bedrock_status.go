package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"go.uber.org/zap"

	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type statusBedrockCmd struct {
	Host string `default:"localhost"`
	Port int    `default:"19132"`

	Protocol string `usage:"protocol to use: auto, nethernet, or raknet" default:"auto"`

	RetryInterval time.Duration `usage:"if retry-limit is non-zero, status will be retried at this interval" default:"10s"`
	RetryLimit    int           `usage:"if non-zero, failed status will be retried this many times before exiting"`
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

func (c *statusBedrockCmd) Execute(_ context.Context, _ *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	logger := args[0].(*zap.Logger)
	address := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	if c.RetryInterval <= 0 {
		c.RetryInterval = 1 * time.Second
	}

	for {
		var info *BedrockServerInfo
		var err error
		switch strings.ToLower(c.Protocol) {
		case "nethernet":
			info, err = PingBedrockNetherNet(address, 0, logger)
		case "raknet":
			info, err = PingBedrockServer(address, 0, logger)
		case "auto", "":
			info, err = PingBedrockNetherNet(address, 3*time.Second, logger)
			if err != nil {
				logger.Debug("NetherNet probe failed, falling back to RakNet", zap.Error(err))
				info, err = PingBedrockServer(address, 0, logger)
			}
		default:
			logger.Fatal("Unknown protocol", zap.String("protocol", c.Protocol))
			return subcommands.ExitUsageError
		}

		if err != nil {
			if c.RetryLimit > 0 {
				c.RetryLimit--
				time.Sleep(c.RetryInterval)
				continue
			}
			logger.Fatal("Failed to ping Bedrock server", zap.Error(err))
			return subcommands.ExitFailure
		}

		fmt.Printf("%s : version=%s online=%d max=%d",
			address,
			info.Version, info.Players, info.MaxPlayers)

		return subcommands.ExitSuccess
	}
}
