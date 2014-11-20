package rpc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/rpc/json"
)

type (
	// Client is a simple JSON-RPC over HTTP client used by the agent.
	Client struct {
		Url string
	}
)

// NewClient create a new client.  This only communicates with 127.0.0.1
func NewClient(port uint, path string) (*Client, error) {
	if path == "" {
		path = RPCPath
	}
	c := &Client{
		//XXX: this seems wrong
		Url: fmt.Sprintf("http://127.0.0.1:%d%s", port, path),
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

// DoRaw calls a service and proxies the response
func (c *Client) DoRaw(method string, request interface{}, rw http.ResponseWriter) {
	data, err := json.EncodeClientRequest(method, request)
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	resp, err := http.Post(c.Url, "application/json", bytes.NewReader(data))
	if err != nil {
		http.Error(rw, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		http.Error(rw, buf.String(), resp.StatusCode)
		return
	} else {
		rw.Header().Set("Content-Type", "application/octet-stream")
		_, err = io.Copy(rw, resp.Body)
		if err != nil {
			http.Error(rw, err.Error(), 500)
		}
		return
	}
}
