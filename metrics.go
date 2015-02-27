package agent

import (
	"errors"
	"fmt"

	"github.com/mistifyio/mistify-agent/rpc"
)

func getMetrics(r *HTTPRequest, mtype string) *HTTPErrorMessage {
	action, err := r.Context.GetAction(fmt.Sprintf("%sMetrics", mtype))
	if err != nil {
		return r.NewError(err, 404)
	}

	// Metric requests are special in that they have Args that can vary by stage
	// Create a unique request for each stage with the args
	response := &rpc.GuestMetricsResponse{}
	pipeline := action.GeneratePipeline(nil, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	for _, stage := range pipeline.Stages {
		stage.Request = &rpc.GuestMetricsRequest{
			Guest: r.Guest,
			Args:  stage.Args,
			Type:  mtype,
		}
	}
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return r.NewError(err, 500)
	}
	switch mtype {
	case "cpu":
		return r.JSON(200, response.CPU)
	case "nic":
		return r.JSON(200, response.Nic)
	case "disk":
		return r.JSON(200, response.Disk)
	}
	return r.NewError(errors.New("Unknown metric"), 500)
}

func getCPUMetrics(r *HTTPRequest) *HTTPErrorMessage {
	return getMetrics(r, "cpu")
}

func getNicMetrics(r *HTTPRequest) *HTTPErrorMessage {
	return getMetrics(r, "nic")
}

func getDiskMetrics(r *HTTPRequest) *HTTPErrorMessage {
	return getMetrics(r, "disk")
}
