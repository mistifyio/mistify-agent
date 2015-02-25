/*
This is a simple example of a sub-agent for Mistify Agent.  It does not do anything useful, but shows the API.
*/
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/mistifyio/mistify-agent/rpc"
	flag "github.com/spf13/pflag"
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

	flag.UintVarP(&port, "port", "p", 21356, "listen port")
	flag.UintVarP(&percent, "percent", "c", 50, "Percentage to return an error")
	flag.Parse()

	if percent > 100 {
		percent = 100
	}

	s := Simple{
		rand:    rand.New(rand.NewSource(time.Now().Unix())),
		percent: int(percent),
	}

	server, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}

	server.RegisterService(&s)
	log.Fatal(server.ListenAndServe())
}
