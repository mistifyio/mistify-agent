package main

import (
	//	"encoding/json"
	"github.com/mistifyio/mistify-agent"
	"github.com/mistifyio/mistify-agent/config"
	"github.com/mistifyio/mistify-agent/log"
)

func main() {
	conf := config.NewConfig()

	err := conf.AddConfig("agent.json")
	if err != nil {
		log.Fatal(err)
	}

	err = conf.Fixup()
	if err != nil {
		log.Fatal(err)
	}
	/*
		data, err := json.MarshalIndent(conf, "   ", " ")
		if err != nil {
			log.Fatal(err)
		}

		log.Info("%s\n", data)
	*/
	ctx, err := agent.NewContext(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = ctx.RunGuests()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(agent.Run(ctx, ":8080"))
}
