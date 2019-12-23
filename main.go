package main

import (
	"context"
	"flag"
	"github.com/google/subcommands"
	"os"
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&statusCmd{}, "")
	subcommands.Register(&gatherTelegrafCmd{}, "monitoring")

	flag.Parse()

	os.Exit(int(subcommands.Execute(context.Background())))
}
