package main

import (
	"github.com/mistifyio/mistify-agent"
	"github.com/mistifyio/mistify-agent/config"
	"github.com/mistifyio/mistify-agent/log"
	flag "github.com/spf13/pflag"
)

func main() {
	var address, logLevel, configFile string

	flag.StringVarP(&address, "address", "a", ":8080", "listen address")
	flag.StringVarP(&logLevel, "log-level", "l", "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.StringVarP(&configFile, "config-file", "c", "", "config file")
	flag.Parse()

	if err := log.SetLogLevel(logLevel); err != nil {
		log.Fatal(err)
	}

	if configFile == "" {
		// TODO: allow a config directory as well
		log.Fatal("need a config file")
	}

	conf := config.NewConfig()

	err := conf.AddConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = conf.Fixup()
	if err != nil {
		log.Fatal(err)
	}

	ctx, err := agent.NewContext(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = ctx.RunGuests()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(agent.Run(ctx, address))
}
