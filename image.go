package agent

import (
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-agent/rpc"
)

func imageMultiQuery(ctx *Context, actionBaseName string, desiredImageType string, request *rpc.ImageRequest) ([]*rpc.Image, *HTTPError) {
	// Determine the set of actions to query based on desired image type
	imageTypes := []string{"", "container"}
	if desiredImageType != "" {
		imageTypes = []string{desiredImageType}
	}
	n := len(imageTypes)

	// Create channels to aggregate results
	resps := make(chan *rpc.ImageResponse, n)
	errors := make(chan error, n)

	// Get the action runner
	runner, err := ctx.GetAgentRunner()
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, err)
	}

	// Query in parallel
	for _, imageType := range imageTypes {
		actionName := prefixedActionName(imageType, actionBaseName)
		go imageQuery(runner, actionName, request, resps, errors)
	}

	// Wait for all to finish and aggregate results
	images := make([]*rpc.Image, 0)
	for i := 0; i < n; i++ {
		select {
		case resp := <-resps:
			images = append(images, resp.Images...)
		case err = <-errors:
			if err.Error() == "no such image" {
				err = nil
				continue
			}
			log.WithField("err", err).Info("image query error")
		}
	}
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, err)
	}

	return images, nil
}

func imageQuery(runner *GuestRunner, actionName string, request *rpc.ImageRequest, respChan chan *rpc.ImageResponse, errChan chan error) {
	response := &rpc.ImageResponse{}

	action, err := runner.Context.GetAction(actionName)
	if err != nil {
		respChan <- response
		return
	}
	pipeline := action.GeneratePipeline(request, response, nil, nil)

	if err = runner.Process(pipeline); err != nil {
		errChan <- err
		return
	}
	respChan <- response
}

// listImages returns a list of images, optionally filtered by type
func listImages(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ctx := getContext(r)
	vars := mux.Vars(r)

	request := &rpc.ImageRequest{}
	images, err := imageMultiQuery(ctx, "listImages", vars["type"], request)
	if err != nil {
		hr.JSON(err.Code, err)
		return
	}
	hr.JSON(http.StatusOK, images)
}

func getImage(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ctx := getContext(r)
	vars := mux.Vars(r)

	request := &rpc.ImageRequest{
		Id: vars["id"],
	}
	images, err := imageMultiQuery(ctx, "getImage", "", request)
	if err != nil {
		hr.JSON(err.Code, err)
		return
	}

	if len(images) < 1 {
		hr.JSONError(http.StatusNotFound, ErrNotFound)
		return
	}

	// This may happen if more than one backend have the image stored under the
	// same id, which is a problem. It will be much less likely when images are
	// all pulled from a central image server that assigns ids.
	if len(images) > 1 {
		log.WithFields(log.Fields{
			"imageID": request.Id,
			"images":  images,
		}).Error("more than one image share id")
	}

	hr.JSON(http.StatusOK, images[0])
}

func deleteImage(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ctx := getContext(r)
	vars := mux.Vars(r)

	response := &rpc.ImageResponse{}
	request := &rpc.ImageRequest{
		Id: vars["id"],
	}

	// First find the image in order to know the type and, therefore, what
	// specific action to use to delete it
	images, mqErr := imageMultiQuery(ctx, "getImage", "", request)
	if mqErr != nil {
		hr.JSON(mqErr.Code, mqErr)
		return
	}

	if len(images) < 1 {
		hr.JSONError(http.StatusNotFound, ErrNotFound)
		return
	}

	// This may happen if more than one backend have the image stored under the
	// same id, which is a problem. It will be much less likely when images are
	// all pulled from a central image server that assigns ids.
	if len(images) > 1 {
		log.WithFields(log.Fields{
			"imageID": request.Id,
			"images":  images,
		}).Error("more than one image share id")
	}

	// Go ahead with the delete
	action, err := ctx.GetAction(prefixedActionName(images[0].Type, "deleteImage"))
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}

	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)

	runner, err := ctx.GetAgentRunner()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	if err := runner.Process(pipeline); err != nil {
		// how to check for not found??
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusAccepted, struct{}{})
}

func fetchImage(w http.ResponseWriter, r *http.Request) {
	hr := HTTPResponse{w}
	ctx := getContext(r)

	request := &rpc.ImageRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		hr.JSONError(http.StatusBadRequest, err)
		return
	}
	if request.Id == "" {
		hr.JSONError(http.StatusBadRequest, errors.New("missing id"))
		return
	}

	response := &rpc.ImageResponse{}
	action, err := ctx.GetAction(prefixedActionName(request.Type, "fetchImage"))
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)

	runner, err := ctx.GetAgentRunner()
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	if err := runner.Process(pipeline); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	hr.JSON(http.StatusAccepted, response)
}
