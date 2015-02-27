package agent

import (
	"encoding/json"

	"github.com/mistifyio/mistify-agent/rpc"
)

func listImages(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ImageResponse{}
	request := &rpc.ImageRequest{}
	action, err := r.Context.GetAction("listImages")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = pipeline.Run()
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Images)
}

func getImage(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ImageResponse{}
	request := &rpc.ImageRequest{
		Id: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("getImage")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = pipeline.Run()
	if err != nil {
		return r.NewError(err, 500)
	}

	if len(response.Images) < 1 {
		return r.NewError(ErrNotFound, 404)
	}

	return r.JSON(200, response.Images[0])
}

func deleteImage(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ImageResponse{}
	request := &rpc.ImageRequest{
		Id: r.Parameter("id"),
	}
	action, err := r.Context.GetAction("deleteImage")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = pipeline.Run()
	// how to check for not found??
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(202, struct{}{})
}

func fetchImage(r *HTTPRequest) *HTTPErrorMessage {
	request := &rpc.ImageRequest{}
	err := json.NewDecoder(r.Request.Body).Decode(request)
	if err != nil {
		return r.NewError(err, 400)
	}
	response := &rpc.ImageResponse{}
	action, err := r.Context.GetAction("fetchImage")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = pipeline.Run()
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(202, response)
}
