package agent

import (
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mistifyio/mistify-agent/config"
)

type (
	Stage struct {
		Service  *Service
		Type     config.ActionType
		Method   string
		Args     map[string]string
		Request  interface{}
		Response interface{}
		RW       http.ResponseWriter // For streaming responses
	}

	Pipeline struct {
		ID       string
		Action   string
		Type     config.ActionType
		Stages   []*Stage
		DoneChan chan error // Signal async is done or errored, for post-hooks
	}

	Action struct {
		Name   string
		Type   config.ActionType
		Stages []*Stage
	}
)

// Run makes an individual stage request
func (stage *Stage) Run() error {
	if stage.Type == config.StreamAction {
		stage.Service.Client.DoRaw(stage.Method, stage.Request, stage.RW)
		return nil
	} else {
		return stage.Service.Client.Do(stage.Method, stage.Request, stage.Response)
	}
}

// Run executes each stage in the pipeline. It bails out as soon as an error
// is encountered
func (pipeline *Pipeline) Run() error {
	for _, stage := range pipeline.Stages {
		if err := stage.Run(); err != nil {
			if pipeline.DoneChan != nil {
				pipeline.DoneChan <- err
			}
			return err
		}
	}
	if pipeline.DoneChan != nil {
		pipeline.DoneChan <- nil
	}
	return nil
}

// GeneragePipeline creates an instance of Pipeline based on an action's
// stages and supplied request & response. It is returned so that any additional
// modifications (such as adding stage args to requests) can be made before
// running if needed.
func (action *Action) GeneratePipeline(request interface{}, response interface{}, rw http.ResponseWriter, done chan error) *Pipeline {
	pipeline := &Pipeline{
		ID:       uuid.New(),
		Action:   action.Name,
		Type:     action.Type,
		Stages:   make([]*Stage, len(action.Stages)),
		DoneChan: done,
	}
	for i, stage := range action.Stages {
		pipeline.Stages[i] = &Stage{
			Service:  stage.Service,
			Type:     action.Type,
			Method:   stage.Method,
			Args:     stage.Args,
			Request:  request,
			Response: response,
			RW:       rw,
		}
	}
	return pipeline
}
