package agent

import (
	"encoding/json"
	"net/http"

	"github.com/mistifyio/mistify-agent/rpc"
)

func listContainerImages(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerImageResponse{}
	request := &rpc.ContainerImageRequest{}
	action, err := r.Context.GetAction("containerListImages")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}
	return r.JSON(http.StatusOK, response.Images)
}

func getContainerImage(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerImageResponse{}
	request := &rpc.ContainerImageRequest{
		ID: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("containerGetImage")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	if len(response.Images) < 1 {
		return r.NewError(ErrNotFound, http.StatusNotFound)
	}

	return r.JSON(http.StatusOK, response.Images[0])
}

func deleteContainerImage(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerImageResponse{}
	request := &rpc.ContainerImageRequest{
		ID: r.Parameter("id"),
	}

	action, err := r.Context.GetAction("containerDeleteImage")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	if len(response.Images) < 1 {
		return r.NewError(ErrNotFound, http.StatusNotFound)
	}

	return r.JSON(http.StatusAccepted, response.Images[0])
}

func pullContainerImage(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ContainerImageResponse{}
	request := &rpc.ContainerImageRequest{}
	if err := json.NewDecoder(r.Request.Body).Decode(request); err != nil {
		return r.NewError(err, http.StatusBadRequest)
	}

	action, err := r.Context.GetAction("containerFetchImage")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	if err := pipeline.Run(); err != nil {
		return r.NewError(err, http.StatusInternalServerError)
	}

	if len(response.Images) < 1 {
		return r.NewError(ErrNotFound, http.StatusNotFound)
	}

	return r.JSON(http.StatusAccepted, response.Images[0])
}
