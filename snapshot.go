package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mistifyio/mistify-agent/client"
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
	response := rpc.SnapshotResponse{}
	request := rpc.SnapshotRequest{Id: getEntityId(r)}
	err := r.Context.ImageClient.Do("ImageStore.ListSnapshots", &request, &response)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func getSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	request := rpc.SnapshotRequest{Id: getEntityId(r)}
	err := r.Context.ImageClient.Do("ImageStore.GetSnapshot", &request, &response)
	if err != nil {
		return getHttpError(r, err)
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
	request.Recursive = r.Parameter("disk") == ""
	if request.Dest == "" {
		request.Dest = fmt.Sprintf("snap-%d", time.Now().Unix())
	}
	err = r.Context.ImageClient.Do("ImageStore.CreateSnapshot", &request, &response)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func deleteSnapshot(r *HttpRequest) *HttpErrorMessage {
	response := rpc.SnapshotResponse{}
	request := rpc.SnapshotRequest{
		Id:        getEntityId(r),
		Recursive: r.Parameter("disk") == "",
	}
	err := r.Context.ImageClient.Do("ImageStore.DeleteSnapshot", &request, &response)
	if err != nil {
		return getHttpError(r, err)
	}
	return r.JSON(200, response.Snapshots)
}

func rollbackSnapshot(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		// Shutdown
		g.Action = "shutdown"
		g, err := r.Context.runSyncAction(g)
		if err != nil {
			return r.NewError(err, 500)
		}

		// Rollback
		response := rpc.SnapshotResponse{}
		var request rpc.SnapshotRequest
		err = json.NewDecoder(r.Request.Body).Decode(&request)
		if err != nil {
			return r.NewError(err, 400)
		}
		request.Id = getEntityId(r)
		err = r.Context.ImageClient.Do("ImageStore.RollbackSnapshot", &request, &response)
		if err != nil {
			return getHttpError(r, err)
		}

		// Startup
		g.Action = "shutdown"
		g, err = r.Context.runSyncAction(g)
		if err != nil {
			return r.NewError(err, 500)
		}

		return r.JSON(200, response.Snapshots)
	})
}

func downloadSnapshot(r *HttpRequest) *HttpErrorMessage {
	request := rpc.SnapshotRequest{
		Id:        getEntityId(r),
		Recursive: r.Parameter("disk") == "",
	}
	data, err := json.Marshal(request)
	if err != nil {
		return r.NewError(err, 400)
	}
	resp, err := http.Post("http://127.0.0.1:16000/snapshots/download", "application/json", bytes.NewReader(data))
	if err != nil {
		return r.NewError(err, 500)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		http.Error(r.ResponseWriter, buf.String(), resp.StatusCode)
	} else {
		r.ResponseWriter.Header().Set("Content-Type", "application/octet-stream")
		_, err = io.Copy(r.ResponseWriter, resp.Body)
		if err != nil {
			http.Error(r.ResponseWriter, err.Error(), 500)
		}
	}
	return nil
}
