package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/rpc"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/spf13/pflag"
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

	ct := "application/json"
	log.WithFields(log.Fields{
		"guest":        request.Guest.Id,
		"content_type": ct,
		"endpoint":     w.Endpoint,
		"body":         string(data),
	}).Info("posting request")

	resp, err := w.Client.Post(w.Endpoint, ct, bytes.NewReader(data))
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

	var port uint
	var endpoint string

	flag.UintVarP(&port, "port", "p", 31245, "listen port")
	flag.StringVarP(&endpoint, "endpoint", "e", "", "webhook endpoint")
	flag.Parse()

	if err := logx.DefaultSetup("info"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
		}).Fatal("failed to set up logging")
	}

	if endpoint == "" {
		log.Fatal("endpoint is required")
	}

	s, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.NewServer",
		}).Fatal(err)
	}

	w := &Webhook{
		Endpoint: endpoint,
		Client: &http.Client{
			Timeout: time.Second * 5,
		},
	}
	if err := s.RegisterService(w); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.RegisterService",
		}).Fatal(err)
	}
	if err = s.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.ListenAndServe",
		}).Fatal(err)
	}
}
