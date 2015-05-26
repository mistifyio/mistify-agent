// Package client provides a simple client for the Mistify Agent.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type (

	// Config is used to configure the creation of a client
	Config struct {
		// Address is the address of the Mistify Agent
		Address string

		// Scheme is the URI scheme for the Mistify Agent
		Scheme string

		// HTTPClient is the client to use. Default will be
		// used if not provided.
		HTTPClient *http.Client
	}

	// Client for the Mistify Agent
	Client struct {
		Config Config
	}
)

// DefaultConfig returns a default configuration for the client
func DefaultConfig() *Config {
	return &Config{
		Address: "127.0.0.1:8080",
		Scheme:  "http",
		HTTPClient: &http.Client{
			Timeout: time.Duration(5 * time.Second),
		},
	}
}

// NewClient returns a new client
func NewClient(config *Config) (*Client, error) {
	defConfig := DefaultConfig()

	if len(config.Address) == 0 {
		config.Address = defConfig.Address
	}

	if len(config.Scheme) == 0 {
		config.Scheme = defConfig.Scheme
	}

	if config.HTTPClient == nil {
		config.HTTPClient = defConfig.HTTPClient
	}

	client := &Client{
		Config: *config,
	}
	return client, nil
}

func (c *Client) doRequest(method, path string, input interface{}, expectedStatus int, output interface{}) error {
	u := url.URL{
		Scheme: c.Config.Scheme,
		Host:   c.Config.Address,
		Path:   path,
	}

	// bug?? must pass nil if no body, not just an empty body??
	var req *http.Request
	var err error
	if input != nil {
		var data []byte
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, u.String(), bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
	}

	if err != nil {
		return err
	}

	resp, err := c.Config.HTTPClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("expected %d but got %d", expectedStatus, resp.StatusCode)
	}
	d := json.NewDecoder(resp.Body)

	err = d.Decode(output)

	return err
}

// ListGuests gets a list of guests
func (c *Client) ListGuests() (GuestSlice, error) {
	guests := make(GuestSlice, 0)
	if err := c.doRequest("GET", "/guests", nil, http.StatusOK, &guests); err != nil {
		return nil, err
	}

	return guests, nil
}

// GetGuest requests creation of a guest
func (c *Client) GetGuest(id string) (*Guest, error) {
	var g Guest
	if err := c.doRequest("GET", filepath.Join("/guests", id), nil, http.StatusOK, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// CreateGuest requests creation of a new guest
func (c *Client) CreateGuest(guest *Guest) (*Guest, error) {
	var g Guest
	if err := c.doRequest("POST", "/guests", guest, http.StatusAccepted, &g); err != nil {
		return nil, err
	}

	return &g, nil
}
