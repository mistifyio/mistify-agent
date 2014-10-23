package main

import (
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent"
	"github.com/mistifyio/mistify-agent/config"
	"github.com/mistifyio/mistify-agent/log"
	"os"
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
