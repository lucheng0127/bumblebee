package cli

import (
	"os"
	"strings"

	"github.com/lucheng0127/bumblebee/pkg/apiserver"
	"github.com/lucheng0127/bumblebee/pkg/config"
	log "github.com/sirupsen/logrus"
	rCli "github.com/urfave/cli/v2"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func run(cCtx *rCli.Context) error {
	debug := false

	// Setup log
	logLevel := cCtx.String("log-level")
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		debug = true
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	log.SetOutput(os.Stdout)

	// Load config
	cfgFile := cCtx.String("config-file")
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return err
	}

	// Validate config
	if errs := cfg.Validate(); len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}

	// Setup singal handler
	ctx := signals.SetupSignalHandler()

	// Create apiserver and launch
	svc, err := apiserver.NewApiServer(cfg, debug, cCtx.String("database"))
	if err != nil {
		return err
	}

	if err := svc.PreRun(ctx); err != nil {
		return err
	}

	return svc.Run(ctx)
}

func NewApiServerCmd() *rCli.Command {
	return &rCli.Command{
		Name:   "apiserver",
		Usage:  "run apiserver",
		Action: run,
		Flags: []rCli.Flag{
			&rCli.StringFlag{
				Name:     "config-file",
				Aliases:  []string{"c"},
				Usage:    "config file of apiserver",
				Required: true,
			},
			&rCli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "log level: debug, info, warn, error",
				Value:   "info",
			},
			&rCli.StringFlag{
				Name:    "database",
				Aliases: []string{"d"},
				Usage:   "sqlite database file",
				Value:   "bumblebee.db",
			},
		},
	}
}
