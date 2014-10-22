package rpc

import (
	"bytes"
	"fmt"
	"github.com/gorilla/rpc/json"
	"net/http"
)

type (
	// Client is a simple JSON-RPC over HTTP client used by the agent.
	Client struct {
		Url string
	}
)

// NewClient create a new client.  This only communicates with 127.0.0.1
func NewClient(port uint) (*Client, error) {
	c := &Client{
		//XXX: this seems wrong
		Url: fmt.Sprintf("http://127.0.0.1:%d%s", port, RPCPath),
	}
	return c, nil
}

// Do calls an RPC method
func (c *Client) Do(method string, request interface{}, response interface{}) error {
	data, err := json.EncodeClientRequest(method, request)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.Url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.DecodeClientResponse(resp.Body, &response)
	if err != nil {
		return err
	}
	return nil
}
