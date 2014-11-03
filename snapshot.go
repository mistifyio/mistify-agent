package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mistifyio/mistify-agent/rpc"
)

func getEntityId(r *HttpRequest) string {
	entityId := make([]string, 2, 6)
	entityId[0] = "guests/"
	entityId[1] = r.Parameter("id")
	if r.Parameter("disk") != "" {
		entityId = append(entityId, "/disk-", r.Parameter("disk")))
	}
	if r.Parameter("name") != "" {
		entityId = append(entityId, "@", r.Parameter("name"))
	}
	return strings.Join(entityId, "")
}

func listSnapshots(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	request := rpc.SnapshotRequest{Id: getEntityId(r)}
	err := r.Context.ImageClient.Do("ImageStore.ListSnapshots", &request, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Snapshots)
}

func getSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	request := rpc.SnapshotRequest{Id: getEntityId(r)}
	err := r.Context.ImageClient.Do("ImageStore.GetSnapshot", &request, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Snapshots)
}

func createSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	var request rpc.SnapshotRequest
	err := json.NewDecoder(r.Request.Body).Decode(&request)
	if err != nil {
		return r.NewError(err, 400)
	}
	request.Id = getEntityId(r)
	request.Recursive = r.Parameter("disk") != ""
	if request.Dest == "" {
		request.Dest = fmt.Sprintf("snap-%s", time.Now().Unix())
	}
	err = r.Context.ImageClient.Do("ImageStore.CreateSnapshots", &request, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Snapshots)
}

func deleteSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	request := rpc.SnapshotRequest{
		Id:        getEntityId(r),
		Recursive: r.Parameter("disk") != "",
	}
	err := r.Context.ImageClient.Do("ImageStore.DeleteSnapshot", &request, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Snapshots)
}

func rollbackSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	var request rpc.SnapshotRequest
	err := json.NewDecoder(r.Request.Body).Decode(&request)
	if err != nil {
		return r.NewError(err, 400)
	}
	request.Id = getEntityId(r)
	err = r.Context.ImageClient.Do("ImageStore.RollbackSnapshot", &request, &response)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(200, response.Snapshots)
}

/*
func DownloadSnapshot(r *HttpRequest) *HttpErrorMessage {
	return r.Context.ImageClient.Do("ImageStore.DownloadSnapshot", r.ResponseWriter, r.Request)
}
*/
