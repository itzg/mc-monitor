package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/subcommands"
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/zapconfigs"
	"go.uber.org/zap"
	"log"
	"os"
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

	var config GlobalConfig
	err := flagsfiller.Parse(&config, flagsfiller.WithEnv(""))
	if err != nil {
		log.Fatal(err)
	}

	var logger *zap.Logger
	if config.Debug {
		logger = zapconfigs.NewDebugLogger()
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

func (c *versionCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	fmt.Printf("%s commit=%s date=%s\n", version, commit, date)
	return subcommands.ExitSuccess
}
