package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-agent/rpc"
)

func getHTTPErrorCode(err error) int {
	if err.Error() == ErrNotFound.Error() {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func getEntityID(vars map[string]string) string {
	entityID := make([]string, 2, 6)
	entityID[0] = "guests/"
	entityID[1] = vars["id"]
	if vars["disk"] != "" {
		entityID = append(entityID, "/disk-", vars["disk"])
	}
	if vars["name"] != "" {
		entityID = append(entityID, "@", vars["name"])
	}
	return strings.Join(entityID, "")
}

func listSnapshots(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := getContext(r)
	runner := getRequestRunner(r)
	vars := mux.Vars(r)

	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{Id: getEntityID(vars)}
	action, err := ctx.GetAction("listSnapshots")
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(getHTTPErrorCode(err), err)
		return
	}
	hr.JSON(http.StatusOK, response.Snapshots)
}

func getSnapshot(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := getContext(r)
	runner := getRequestRunner(r)
	vars := mux.Vars(r)
	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{Id: getEntityID(vars)}
	action, err := ctx.GetAction("getSnapshot")
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(getHTTPErrorCode(err), err)
		return
	}
	hr.JSON(http.StatusOK, response.Snapshots)
}

func createSnapshot(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := getContext(r)
	runner := getRequestRunner(r)
	vars := mux.Vars(r)

	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		hr.JSONError(http.StatusBadRequest, err)
		return
	}
	request.Id = getEntityID(vars)
	// If no disk is specified, recursively snapshot all of the guest's disks.
	request.Recursive = vars["disk"] == ""
	if request.Dest == "" {
		request.Dest = fmt.Sprintf("snap-%d", time.Now().Unix())
	}
	action, err := ctx.GetAction("createSnapshot")
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(getHTTPErrorCode(err), err)
		return
	}
	hr.JSON(http.StatusOK, response.Snapshots)
}

func deleteSnapshot(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := getContext(r)
	runner := getRequestRunner(r)
	vars := mux.Vars(r)

	response := &rpc.SnapshotResponse{}
	// If no disk is specified, recursively delete all of the guest's disks'
	// snapshots.
	request := &rpc.SnapshotRequest{
		Id:        getEntityID(vars),
		Recursive: vars["disk"] == "",
	}
	action, err := ctx.GetAction("deleteSnapshot")
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(getHTTPErrorCode(err), err)
		return
	}
	hr.JSON(http.StatusOK, response.Snapshots)
}

func rollbackSnapshot(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := getContext(r)
	runner := getRequestRunner(r)
	vars := mux.Vars(r)

	response := &rpc.SnapshotResponse{}
	request := &rpc.SnapshotRequest{}
	err := json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	request.Id = getEntityID(vars)
	action, err := ctx.GetAction("rollbackSnapshot")
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(getHTTPErrorCode(err), err)
		return
	}

	hr.JSON(http.StatusOK, response.Snapshots)
}

func downloadSnapshot(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := getContext(r)
	runner := getRequestRunner(r)
	vars := mux.Vars(r)

	response := &rpc.SnapshotResponse{}
	// If no disk is specified, recursively download all of the guest's disks'
	// snapshots.
	request := &rpc.SnapshotRequest{
		Id:        getEntityID(vars),
		Recursive: vars["disk"] == "",
	}
	action, err := ctx.GetAction("downloadSnapshot")
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	// Streaming handles sending its own error responses
	_ = runner.Process(pipeline)

	return
}
