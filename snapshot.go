package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mistifyio/mistify-agent/rpc"
)

func getHttpError(r *HttpRequest, err error) *HttpErrorMessage {
	if err.Error() == NotFound.Error() {
		return r.JSON(404, err)
	}
	return r.NewError(err, 500)
}

func getEntityId(r *HttpRequest) string {
	entityId := make([]string, 2, 6)
	entityId[0] = "guests/"
	entityId[1] = r.Parameter("id")
	if r.Parameter("disk") != "" {
		entityId = append(entityId, "/disk-", r.Parameter("disk"))
	}
	if r.Parameter("name") != "" {
		entityId = append(entityId, "@", r.Parameter("name"))
	}
	return strings.Join(entityId, "")
}

func listSnapshots(r *HttpRequest) *HttpErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{Id: getEntityId(r)}
	action, err := r.Context.GetAction("listSnapshots")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func getSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{Id: getEntityId(r)}
	action, err := r.Context.GetAction("getSnapshot")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func createSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{}
	err := json.NewDecoder(r.Request.Body).Decode(request)
	if err != nil {
		return r.NewError(err, 400)
	}
	request.Id = getEntityId(r)
	request.Recursive = r.Parameter("disk") == ""
	if request.Dest == "" {
		request.Dest = fmt.Sprintf("snap-%d", time.Now().Unix())
	}
	action, err := r.Context.GetAction("createSnapshot")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func deleteSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{
		Id:        getEntityId(r),
		Recursive: r.Parameter("disk") == "",
	}
	action, err := r.Context.GetAction("deleteSnapshot")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func rollbackSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{}
	err := json.NewDecoder(r.Request.Body).Decode(request)
	if err != nil {
		return r.NewError(err, 400)
	}
	request.Id = getEntityId(r)
	action, err := r.Context.GetAction("rollbackSnapshot")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	err = r.GuestRunner.Process(pipeline)
	if err != nil {
		return getHttpError(r, err)
	}

	return r.JSON(200, response.Snapshots)
}

func downloadSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{
		Id:        getEntityId(r),
		Recursive: r.Parameter("disk") == "",
	}
	action, err := r.Context.GetAction("downloadSnapshot")
	if err != nil {
		return r.NewError(err, 404)
	}
	pipeline := action.GeneratePipeline(request, response, r.ResponseWriter, nil)
	// Streaming handles sending its own error responses
	r.GuestRunner.Process(pipeline)

	return nil
}
