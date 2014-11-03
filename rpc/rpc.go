// Package rpc defines the JSON-RPC over HTTP types the Agent uses to communicate with sub-agents.  It also contains various helpers for writing sub-agents.
package rpc

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
)

const (
	// RPCPath is the URI endpoint that the Agent posts to for sub-agent communication.
	RPCPath = "/_mistify_RPC_"
)

type Codec struct {
	*json.Codec
}

func NewCodec() *Codec {
	c := json.NewCodec()
	return &Codec{c}
}

func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	cr := c.Codec.NewRequest(r)

	if m, err := cr.Method(); err == nil {
		context.Set(r, RPCPath, m)
	}

	return cr
}
