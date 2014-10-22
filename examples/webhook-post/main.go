// This is a test Mistify sub-agent that implements most of the guest 'actions'.
// It does not modify the guest is anyway
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent/rpc"
	"log"
	"net/http"
	"os"
	"time"
)

type (
	// Webhook is an endpoint
	Webhook struct {
		Endpoint string
		Client   *http.Client
	}

	// Payload represents a message to be posted
	Payload struct {
		Text  string `json:"text"`
		Emoji string `json:"icon_emoji"`
		Name  string `json:"username"`
	}
)

// Post posts the action to the endpoint and returns the guest
func (w *Webhook) Post(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {

	// just in case
	if request.Guest == nil || request.Guest.Id == "" {
		return errors.New("invalid guest")
	}

	action := "unknown"
	emoji := "rainbow"
	username := "mistify-webhook-post"
	if request.Action != "" {
		action = request.Action
	}

	if request.Args != nil {
		val, ok := request.Args["emoji"]
		if ok && val != "" {
			emoji = val
		}
		val, ok = request.Args["name"]
		if ok && val != "" {
			username = val
		}
	}
	payload := Payload{
		Text:  fmt.Sprintf("%s called on %s", action, request.Guest.Id),
		Emoji: fmt.Sprintf(":%s:", emoji),
		Name:  username,
	}

	data, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	log.Printf("body: %s", string(data))
	resp, err := w.Client.Post(w.Endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp.Body.Close()

	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

func main() {

	var port int
	var h bool
	var endpoint string

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.IntVar(&port, []string{"p", "#port", "-port"}, 31245, "listen port")
	flag.StringVar(&endpoint, []string{"e", "#endpoint", "-endpoint"}, "", "webhook endpoint")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if endpoint == "" {
		log.Fatalf("endpoint is required")
	}

	s, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}

	w := &Webhook{
		Endpoint: endpoint,
		Client: &http.Client{
			Timeout: time.Second * 5,
		},
	}
	s.RegisterService(w)
	log.Fatal(s.ListenAndServe())
}
