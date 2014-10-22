package agent

import (
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/log"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
	// RPC Service
	Service struct {
		ctx      *Context
		Client   *rpc.Client
		Name     string
		requests chan struct{} //functions as a concurrent count limiter
	}

	// Stage of the pipeline
	Stage struct {
		Service *Service
		Action  string
		Method  string
		Args    map[string]string
	}

	// RPC Pipeline
	Pipeline struct {
		Name   string
		Stages []*Stage
		ctx    *Context
	}

	Action struct {
		Sync  *Pipeline
		Async *Pipeline
		ctx   *Context
		Name  string
	}
)

func (ctx *Context) NewService(name string, port uint, maxConcurrent uint) (*Service, error) {
	c, err := rpc.NewClient(port)
	if err != nil {
		return nil, err
	}
	s := &Service{
		ctx:      ctx,
		Client:   c,
		Name:     name,
		requests: make(chan struct{}, maxConcurrent),
	}

	empty := struct{}{}
	for i := uint(0); i < maxConcurrent; i++ {
		s.requests <- empty
	}
	return s, err
}

func (ctx *Context) NewPipeline(stages []*Stage, name string) (*Pipeline, error) {
	p := &Pipeline{
		Name:   name,
		Stages: stages,
		ctx:    ctx,
	}
	return p, nil
}

func (stage *Stage) Run(g *client.Guest) (*client.Guest, error) {

	log.Info("%+v\n", stage)

	service := stage.Service
	// should wrap in timeout?
	empty := <-service.requests
	defer func() {
		service.requests <- empty
	}()

	req := rpc.GuestRequest{
		Guest:  g,
		Args:   stage.Args,
		Action: stage.Action,
	}

	var resp rpc.GuestResponse

	err := service.Client.Do(stage.Method, &req, &resp)

	if err != nil {
		return nil, err
	}

	// TODO: backoff handling, included hinted backoff

	return resp.Guest, nil
}

// XXX: should we have some sort of results return type?
func (p *Pipeline) Run(guest *client.Guest) (*client.Guest, error) {
	log.Info("%+v\n", p)
	for _, s := range p.Stages {
		var err error
		// TODO: capture timings
		guest, err = s.Run(guest)
		if err != nil {
			return nil, err
		}
		err = p.ctx.PersistGuest(guest)
		if err != nil {
			return nil, err
		}
	}
	return guest, nil
}
