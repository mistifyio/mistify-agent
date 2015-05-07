package agent

import (
	"encoding/json"
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/rpc"
)

func imageMultiQuery(r *HTTPRequest, actionBaseName string, request *rpc.ImageRequest) ([]*rpc.Image, *HTTPErrorMessage) {
	// Determine the set of actions to query based on desired image type
	desiredImageType := r.Parameter("type")
	imageTypes := []string{"", "container"}
	if desiredImageType != "" {
		imageTypes = []string{desiredImageType}
	}
	n := len(imageTypes)

	// Create channels to aggregate results
	resps := make(chan *rpc.ImageResponse, n)
	errors := make(chan error, n)

	// Get the action runner
	runner, err := r.Context.GetAgentRunner()
	if err != nil {
		return nil, r.NewError(err, 500)
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
		return nil, r.NewError(err, 500)
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
	return
}

// listImages returns a list of images, optionally filtered by type
func listImages(r *HTTPRequest) *HTTPErrorMessage {
	request := &rpc.ImageRequest{}
	images, err := imageMultiQuery(r, "listImages", request)
	if err != nil {
		return err
	}
	return r.JSON(200, images)
}

func getImage(r *HTTPRequest) *HTTPErrorMessage {
	request := &rpc.ImageRequest{
		Id: r.Parameter("id"),
	}
	images, err := imageMultiQuery(r, "getImage", request)
	if err != nil {
		return err
	}

	if len(images) < 1 {
		return r.NewError(ErrNotFound, 404)
	}
	if len(images) > 1 {
		log.WithFields(log.Fields{
			"imageID": request.Id,
			"images":  images,
		}).Error("more than one image share id")
	}

	return r.JSON(200, images[0])
}

func deleteImage(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.ImageResponse{}
	request := &rpc.ImageRequest{
		Id: r.Parameter("id"),
	}

	// First find the image in order to know the type and, therefore, what
	// specific action to use to delete it
	images, mqErr := imageMultiQuery(r, "getImage", request)
	if mqErr != nil {
		return mqErr
	}

	if len(images) < 1 {
		return r.NewError(ErrNotFound, 404)
	}

	if len(images) > 1 {
		log.WithFields(log.Fields{
			"imageID": request.Id,
			"images":  images,
		}).Error("more than one image share id")
	}

	// Go ahead with the delete
	action, err := r.Context.GetAction(prefixedActionName(images[0].Type, "deleteImage"))
	if err != nil {
		return r.NewError(err, 404)
	}

	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)

	runner, err := r.Context.GetAgentRunner()
	if err != nil {
		return r.NewError(err, 500)
	}

	if err := runner.Process(pipeline); err != nil {
		// how to check for not found??
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
	if request.Source == "" {
		return r.NewError(errors.New("missing source"), 400)
	}

	response := &rpc.ImageResponse{}
	action, err := r.Context.GetAction(prefixedActionName(request.Type, "fetchImage"))
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)

	runner, err := r.Context.GetAgentRunner()
	if err != nil {
		return r.NewError(err, 500)
	}

	if err := runner.Process(pipeline); err != nil {
		return r.NewError(err, 500)
	}

	return r.JSON(202, response)
}
