package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent"
	"github.com/mistifyio/mistify-agent/config"
	flag "github.com/spf13/pflag"
)

func main() {
	var address, logLevel, configFile string

	flag.StringVarP(&address, "address", "a", ":8080", "listen address")
	flag.StringVarP(&logLevel, "log-level", "l", "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.StringVarP(&configFile, "config-file", "c", "", "config file")
	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "log.ParseLevel",
		}).Fatal(err)
	}
	log.SetLevel(level)

	if configFile == "" {
		// TODO: allow a config directory as well
		log.Fatal("need a config file")
	}

	conf := config.NewConfig()

	if err := conf.AddConfig(configFile); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "config.AddConfig",
		}).Fatal(err)
	}

	if err = conf.Fixup(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "config.Fixup",
		}).Fatal(err)
	}

	ctx, err := agent.NewContext(conf)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "agent.NewContext",
		}).Fatal(err)
	}

	if err = ctx.RunGuests(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "agent.Context.RunGuests",
		}).Fatal(err)
	}

	if err = agent.Run(ctx, address); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "agent.Run",
		}).Fatal(err)
	}
}
