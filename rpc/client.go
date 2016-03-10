package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	rpcJSON "github.com/gorilla/rpc/json"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

type (
	// Client is a simple JSON-RPC over HTTP client used by the agent.
	Client struct {
		URL string
	}
)

// NewClient create a new client.  This only communicates with 127.0.0.1
func NewClient(port uint, path string) (*Client, error) {
	if path == "" {
		path = RPCPath
	}
	c := &Client{
		//XXX: this seems wrong
		URL: fmt.Sprintf("http://127.0.0.1:%d%s", port, path),
	}
	return c, nil
}

// Do calls an RPC method
func (c *Client) Do(method string, request interface{}, response interface{}) error {
	data, err := rpcJSON.EncodeClientRequest(method, request)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.URL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer logx.LogReturnedErr(resp.Body.Close, nil, "failed to close response body")

	if resp.StatusCode >= http.StatusBadRequest {
		var buf bytes.Buffer
		if _, err = buf.ReadFrom(resp.Body); err != nil {
			return err
		}
		return errors.New(buf.String())
	}

	err = rpcJSON.DecodeClientResponse(resp.Body, &response)
	if err != nil {
		return err
	}
	return nil
}

// DoRaw calls a service and proxies the response
func (c *Client) DoRaw(request interface{}, rw http.ResponseWriter) {
	data, err := json.Marshal(request)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := http.Post(c.URL, "application/json", bytes.NewReader(data))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	defer logx.LogReturnedErr(resp.Body.Close, nil, "failed to close response body")

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(resp.Body); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(rw, buf.String(), resp.StatusCode)
		return
	}
	rw.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(rw, resp.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	return
}
