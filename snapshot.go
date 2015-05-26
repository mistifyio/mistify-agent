package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mistifyio/mistify-agent/rpc"
)

func getHTTPError(r *HTTPRequest, err error) *HTTPErrorMessage {
	if err.Error() == ErrNotFound.Error() {
		return r.JSON(http.StatusNotFound, err)
	}
	return r.NewError(err, http.StatusInternalServerError)
}

func getEntityID(r *HTTPRequest) string {
	entityID := make([]string, 2, 6)
	entityID[0] = "guests/"
	entityID[1] = r.Parameter("id")
	if r.Parameter("disk") != "" {
		entityID = append(entityID, "/disk-", r.Parameter("disk"))
	}
	if r.Parameter("name") != "" {
		entityID = append(entityID, "@", r.Parameter("name"))
	}
	return strings.Join(entityID, "")
}

func listSnapshots(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{Id: getEntityID(r)}
	action, err := r.Context.GetAction("listSnapshots")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHTTPError(r, err)
	}
	return r.JSON(http.StatusOK, response.Snapshots)
}

func getSnapshot(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{Id: getEntityID(r)}
	action, err := r.Context.GetAction("getSnapshot")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHTTPError(r, err)
	}
	return r.JSON(http.StatusOK, response.Snapshots)
}

func createSnapshot(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{}
	err := json.NewDecoder(r.Request.Body).Decode(request)
	if err != nil {
		return r.NewError(err, http.StatusBadRequest)
	}
	request.Id = getEntityID(r)
	request.Recursive = r.Parameter("disk") == ""
	if request.Dest == "" {
		request.Dest = fmt.Sprintf("snap-%d", time.Now().Unix())
	}
	action, err := r.Context.GetAction("createSnapshot")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHTTPError(r, err)
	}
	return r.JSON(http.StatusOK, response.Snapshots)
}

func deleteSnapshot(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{
		Id:        getEntityID(r),
		Recursive: r.Parameter("disk") == "",
	}
	action, err := r.Context.GetAction("deleteSnapshot")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHTTPError(r, err)
	}
	return r.JSON(http.StatusOK, response.Snapshots)
}

func rollbackSnapshot(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{}
	err := json.NewDecoder(r.Request.Body).Decode(request)
	if err != nil {
		return r.NewError(err, http.StatusBadRequest)
	}
	request.Id = getEntityID(r)
	action, err := r.Context.GetAction("rollbackSnapshot")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHTTPError(r, err)
	}

	return r.JSON(http.StatusOK, response.Snapshots)
}

func downloadSnapshot(r *HTTPRequest) *HTTPErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{
		Id:        getEntityID(r),
		Recursive: r.Parameter("disk") == "",
	}
	action, err := r.Context.GetAction("downloadSnapshot")
	if err != nil {
		return r.NewError(err, http.StatusNotFound)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	// Streaming handles sending its own error responses
	_ = r.GuestRunner.Process(pipeline)

	return nil
}
