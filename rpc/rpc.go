// Package rpc defines the JSON-RPC over HTTP types the Agent uses to communicate with sub-agents.  It also contains various helpers for writing sub-agents.
package rpc

const (
	// RPCPath is the URI endpoint that the Agent posts to for sub-agent communication.
	RPCPath = "/_mistify_RPC_"
)
