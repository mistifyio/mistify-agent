package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent"
	"github.com/mistifyio/mistify-agent/config"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/spf13/pflag"
)

func main() {
	var address, logLevel, configFile string

	flag.StringVarP(&address, "address", "a", ":8080", "listen address")
	flag.StringVarP(&logLevel, "log-level", "l", "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.StringVarP(&configFile, "config-file", "c", "", "config file")
	flag.Parse()

	if err := logx.DefaultSetup(logLevel); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
			"level": logLevel,
		}).Fatal("failed to set up logging")
	}

	if configFile == "" {
		// TODO: allow a config directory as well
		log.Fatal("need a config file")
	}

	conf := config.NewConfig()

	if err := conf.AddConfig(configFile); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "config.AddConfig",
		}).Fatal("failed to add config")
	}

	if err := conf.Fixup(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "config.Fixup",
		}).Fatal("failed to fix up config")
	}

	ctx, err := agent.NewContext(conf)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "agent.NewContext",
		}).Fatal("failed to create agent context")
	}

	if err := ctx.CreateJobLog(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "ctx.CreateJobLog",
		}).Fatal("failed to create job log")
	}

	if err = ctx.RunGuests(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "agent.Context.RunGuests",
		}).Fatal("failed to run guests")
	}

	if err = agent.Run(ctx, address); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "agent.Run",
		}).Fatal("failed to run agent")
	}
}
