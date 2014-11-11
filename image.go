package agent

import (
	"encoding/json"

	"github.com/mistifyio/mistify-agent/rpc"
)

func listImages(r *HttpRequest) *HttpErrorMessage {
	response := rpc.ImageResponse{}
	handler := r.Context.ImageActions["listImages"]
	err := handler.Service.Client.Do(handler.Method, &rpc.ImageRequest{}, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Images)
}

func getImage(r *HttpRequest) *HttpErrorMessage {
	response := rpc.ImageResponse{}
	handler := r.Context.ImageActions["getImage"]
	err := handler.Service.Client.Do(handler.Method, &rpc.ImageRequest{Id: r.Parameter("id")}, &response)
	if err != nil {
		return r.NewError(err, 500)
	}

	if len(response.Images) < 1 {
		return r.NewError(NotFound, 404)
	}

	return r.JSON(200, response.Images[0])
}

func deleteImage(r *HttpRequest) *HttpErrorMessage {
	response := rpc.ImageResponse{}
	handler := r.Context.ImageActions["deleteImage"]
	err := handler.Service.Client.Do(handler.Method, &rpc.ImageRequest{Id: r.Parameter("id")}, &response)
	// how to check for not found??
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(202, struct{}{})
}

func fetchImage(r *HttpRequest) *HttpErrorMessage {
	var req rpc.ImageRequest
	err := json.NewDecoder(r.Request.Body).Decode(&req)
	if err != nil {
		return r.NewError(err, 400)
	}
	response := rpc.ImageResponse{}
	handler := r.Context.ImageActions["fetchImage"]
	err = handler.Service.Client.Do(handler.Method, &req, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(202, response)
}
