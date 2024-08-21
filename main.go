package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/zapconfigs"
	"go.uber.org/zap"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&statusCmd{}, "status")
	subcommands.Register(&statusBedrockCmd{}, "status")
	subcommands.Register(&gatherTelegrafCmd{}, "monitoring")
	subcommands.Register(&exportPrometheusCmd{}, "monitoring")
	subcommands.Register(&collectOpenTelemetryCmd{}, "monitoring")

	var config GlobalConfig
	err := flagsfiller.Parse(&config, flagsfiller.WithEnv(""))
	if err != nil {
		log.Fatal(err)
	}

	var logger *zap.Logger
	if config.Debug {
		zapConfig := zap.Config{
			Encoding:      "console",
			EncoderConfig: zapconfigs.NewDebugEncoderConfig(),
			Level:         zap.NewAtomicLevelAt(zap.DebugLevel),
			// output to stderr so that scripts grabbing output don't get logs
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		var err error
		logger, err = zapConfig.Build()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		logger = zapconfigs.NewDefaultLogger()
	}
	defer logger.Sync()

	os.Exit(int(subcommands.Execute(context.Background(), logger)))
}

type GlobalConfig struct {
	Debug bool `usage:"enable debug logging"`
}

type versionCmd struct{}

func (c *versionCmd) Name() string {
	return "version"
}

func (c *versionCmd) Synopsis() string {
	return "Show version and exit"
}

func (c *versionCmd) Usage() string {
	return ""
}

func (c *versionCmd) SetFlags(*flag.FlagSet) {
}

func (c *versionCmd) Execute(context.Context, *flag.FlagSet, ...interface{}) subcommands.ExitStatus {
	fmt.Printf("%s commit=%s date=%s\n", version, commit, date)
	return subcommands.ExitSuccess
}
