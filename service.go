package agent

import "github.com/mistifyio/mistify-agent/rpc"

type (
	// Service is an RPC service
	Service struct {
		ctx    *Context
		Client *rpc.Client
		Name   string
	}
)

// NewService creates a new Service
func (ctx *Context) NewService(name string, port uint, path string, maxConcurrent uint) (*Service, error) {
	c, err := rpc.NewClient(port, path)
	if err != nil {
		return nil, err
	}
	s := &Service{
		ctx:    ctx,
		Client: c,
		Name:   name,
	}

	return s, err
}
