package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent"
	"github.com/mistifyio/mistify-agent/config"
)

func main() {
	var address, logLevel, configFile string
	var h bool

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.StringVar(&address, []string{"a", "#address", "-address"}, ":8080", "listen address")
	flag.StringVar(&logLevel, []string{"l", "#log-level", "-log-level"}, "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.StringVar(&configFile, []string{"c", "#config-file", "-config-file"}, "", "config file")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

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
