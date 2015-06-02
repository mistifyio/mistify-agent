package agent

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mistifyio/mistify-agent/rpc"
)

func getMetrics(w http.ResponseWriter, r *http.Request, mtype string) {
	hr := &HTTPResponse{w}
	ctx := GetContext(r)
	guest := GetRequestGuest(r)
	runner := GetRequestRunner(r)

	action, err := ctx.GetAction(fmt.Sprintf("%sMetrics", mtype))
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}

	// Metric requests are special in that they have Args that can vary by stage
	// Create a unique request for each stage with the args
	// TODO: Fix to use Pre/PostStage functions
	response := &rpc.GuestMetricsResponse{}
	pipeline := action.GeneratePipeline(nil, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	for _, stage := range pipeline.Stages {
		stage.Request = &rpc.GuestMetricsRequest{
			Guest: guest,
			Args:  stage.Args,
			Type:  mtype,
		}
	}
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	switch mtype {
	case "cpu":
		hr.JSON(http.StatusOK, response.CPU)
		return
	case "nic":
		hr.JSON(http.StatusOK, response.Nic)
		return
	case "disk":
		hr.JSON(http.StatusOK, response.Disk)
		return
	}
	hr.JSONError(http.StatusInternalServerError, errors.New("Unknown metric"))
}

func getCPUMetrics(w http.ResponseWriter, r *http.Request) {
	getMetrics(w, r, "cpu")
}

func getNicMetrics(w http.ResponseWriter, r *http.Request) {
	getMetrics(w, r, "nic")
}

func getDiskMetrics(w http.ResponseWriter, r *http.Request) {
	getMetrics(w, r, "disk")
}
