package main

import (
	"os"

	"github.com/lucheng0127/bumblebee/pkg/cli"

	log "github.com/sirupsen/logrus"
	rCli "github.com/urfave/cli/v2"
)

func main() {
	app := &rCli.App{
		Commands: []*rCli.Command{
			cli.NewApiServerCmd(),
			cli.NewUserCmd(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
