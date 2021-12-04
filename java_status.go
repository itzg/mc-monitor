package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/itzg/mc-monitor/slp"
	"go.uber.org/zap"
	"log"
	"os"
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
)

type statusCmd struct {
	Host string `default:"localhost" usage:"hostname of the Minecraft server" env:"MC_HOST"`
	Port int    `default:"25565" usage:"port of the Minecraft server" env:"MC_PORT"`

	UseServerListPing bool `usage:"indicates the legacy, server list ping should be used for pre-1.12"`

	RetryInterval time.Duration `usage:"if retry-limit is non-zero, status will be retried at this interval" default:"10s"`
	RetryLimit    int           `usage:"if non-zero, failed status will be retried this many times before exiting"`
	Timeout       time.Duration `usage:"the timeout the ping can take as a maximum" default:"60s"`

	UseProxy     bool `usage:"supports contacting Bungeecord when proxy_protocol enabled"`
	ProxyVersion byte `usage:"version of PROXY protocol to use" default:"1"`

	ShowPlayerCount bool `usage:"show just the online player count"`
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

func (c *statusCmd) Execute(ctx context.Context, fs *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	logger := args[0].(*zap.Logger)

	if c.UseServerListPing {
		return c.ExecuteServerListPing()
	}

	var options []mcpinger.McPingerOption
	if c.Timeout > 0 {
		options = append(options, mcpinger.WithTimeout(c.Timeout))
	}
	if c.UseProxy {
		options = append(options, mcpinger.WithProxyProto(c.ProxyVersion))
	}

	if c.RetryInterval <= 0 {
		c.RetryInterval = 1 * time.Second
	}

	for {
		logger.Debug("pinging")
		pinger := mcpinger.New(c.Host, uint16(c.Port), options...)
		info, err := pinger.Ping()
		logger.Debug("ping returned", zap.Error(err), zap.Any("info", info))

		if err != nil {
			if c.RetryLimit > 0 {
				c.RetryLimit--
				time.Sleep(c.RetryInterval)
				continue
			}
			_, _ = fmt.Fprintf(os.Stderr, "failed to ping %s:%d : %s", c.Host, c.Port, err)
			return subcommands.ExitFailure
		}

		// While server is starting up it will answer pings, but respond with empty JSON object.
		// As such, we'll sanity check the max players value to see if a zero-value has been
		// provided for info.
		if info.Players.Max == 0 {
			if c.RetryLimit > 0 {
				c.RetryLimit--
				time.Sleep(c.RetryInterval)
				continue
			}
			_, _ = fmt.Fprintf(os.Stderr, "server not ready %s:%d", c.Host, c.Port)
			return subcommands.ExitFailure
		}

		if c.ShowPlayerCount {
			fmt.Printf("%d\n", info.Players.Online)
		} else {
			fmt.Printf("%s:%d : version=%s online=%d max=%d motd='%s'\n",
				c.Host, c.Port,
				info.Version.Name, info.Players.Online, info.Players.Max, info.Description.Text)
		}

		return subcommands.ExitSuccess
	}
}

func (c *statusCmd) ExecuteServerListPing() subcommands.ExitStatus {
	err := retry.Do(func() error {
		response, err := slp.ServerListPing(c.Host, c.Port, c.Timeout)
		if err != nil {
			return err
		}

		if response.MaxPlayers == "0" {
			return errors.New("server not ready")
		}

		if c.ShowPlayerCount {
			fmt.Printf("%s\n", response.CurrentPlayerCount)
		} else {
			fmt.Printf("%s:%d : version=%s online=%s max=%s motd='%s'\n",
				c.Host, c.Port,
				response.ServerVersion, response.CurrentPlayerCount, response.MaxPlayers, response.MessageOfTheDay)
		}

		return nil
	}, retry.Delay(c.RetryInterval), retry.Attempts(uint(c.RetryLimit+1)))

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to ping %s:%d : %s", c.Host, c.Port, err)
		return subcommands.ExitFailure
	}

	// regular output is within Do function
	return subcommands.ExitSuccess
}
