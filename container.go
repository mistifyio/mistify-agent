package agent

import (
	"encoding/json"
	"net/http"

	"github.com/mistifyio/mistify-agent/rpc"
)

func listContainers(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerResponse{}
	request := &rpc.ContainerRequest{}
	action, err := r.Context.GetAction("listContainers")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}
	return r.JSON(http.StatusOK, response.Containers)
}

func getContainer(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerResponse{}
	request := &rpc.ContainerRequest{
		ID: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("getContainer")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	if len(response.Containers) < 1 {
		return r.NewError(ErrNotFound, http.StatusNotFound)
	}

	return r.JSON(http.StatusOK, response.Containers[0])
}

func createContainer(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerResponse{}
	request := &rpc.ContainerRequest{}
	if err := json.NewDecoder(r.Request.Body).Decode(request); err != nil {
		return r.NewError(err, http.StatusBadRequest)
	}

	action, err := r.Context.GetAction("createContainer")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	return r.JSON(http.StatusAccepted, response.Containers[0])
}

func startContainer(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerResponse{}
	request := &rpc.ContainerRequest{
		ID: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("startContainer")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	return r.JSON(http.StatusAccepted, response.Containers[0])
}
func stopContainer(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerResponse{}
	request := &rpc.ContainerRequest{
		ID: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("stopContainer")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	return r.JSON(http.StatusAccepted, response.Containers[0])
}
func deleteContainer(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerResponse{}
	request := &rpc.ContainerRequest{
		ID: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("deleteContainer")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	return r.JSON(http.StatusAccepted, response.Containers[0])
}
