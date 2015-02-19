/*
This is a simple example of a sub-agent for Mistify Agent.  It does not do anything useful, but shows the API.
*/
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
	Simple struct {
		rand    *rand.Rand // random number generator
		percent int        // how often to return an error
	}
)

// DoStuff does not actually do anything. It returns an error a certain percentage of the time.
func (s *Simple) DoStuff(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	if num := s.rand.Intn(100); num <= s.percent {
		return fmt.Errorf("returning an error as I do %d%% of the time", s.percent)
	}
	// just return the guest from the response
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

func main() {

	var port uint
	var percent uint
	var h bool

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.UintVar(&port, []string{"p", "#port", "-port"}, 21356, "listen port")
	flag.UintVar(&percent, []string{"c", "#percent", "-percent"}, 50, "Percentage to return an error")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if percent > 100 {
		percent = 100
	}

	s := Simple{
		rand:    rand.New(rand.NewSource(time.Now().Unix())),
		percent: int(percent),
	}

	server, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.NewServer",
		}).Fatal(err)
	}

	server.RegisterService(&s)
	if err = server.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.ListenAndServe",
		}).Fatal(err)
	}
}
