// This is a test Mistify sub-agent that implements all the actions the agent may call.
// It returns the guest received for Guest requests and fake metrics for metric requests.
// A sub-agent should generally only have one area of concern and do one thing well.  This allows
// sub-agents to be composited in various ways.
package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/spf13/pflag"
)

type (
	// Test is the basic struct for the test service
	Test struct {
	}
)

// Restart hard restarts the Guest. This is like unplugging and plugging in the power.
func (t *Test) Restart(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// Poweroff is like removing the power.
func (t *Test) Poweroff(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// Delete removes a VM. A sub-agent may need to clean up any auxillery configuration associated with it.
func (t *Test) Delete(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// Create a VM.  A sub-agent may call out to another service to configure a switch for example.
func (t *Test) Create(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// Run is called periodically for a running VM.  A sub-agent may need to maintain some state about a VM.
func (t *Test) Run(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// Reboot issues a soft-reboot
func (t *Test) Reboot(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// Shutdown is for a soft ACPI shutdown
func (t *Test) Shutdown(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	*response = rpc.GuestResponse{
		Guest: request.Guest,
	}
	return nil
}

// CPUMetrics retrieves CPU metrics.  Currently, only one sub-agent service is called for this.
func (t *Test) CPUMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	*response = rpc.GuestMetricsResponse{
		Guest: request.Guest,
		Type:  "cpu",
		CPU: []*client.GuestCPUMetrics{
			{},
		},
	}
	return nil
}

// DiskMetrics retrieves Disk metrics.  Currently, only one sub-agent service is called for this.
func (t *Test) DiskMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	*response = rpc.GuestMetricsResponse{
		Guest: request.Guest,
		Type:  "disk",
		Disk: map[string]*client.GuestDiskMetrics{
			"vda": {
				Disk: "vda",
			},
		},
	}
	return nil
}

// NicMetrics retrieves Network metrics.  Currently, only one sub-agent service is called for this.
func (t *Test) NicMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	*response = rpc.GuestMetricsResponse{
		Guest: request.Guest,
		Type:  "disk",
		Nic: map[string]*client.GuestNicMetrics{
			"net0": {
				Name: "net0",
			},
		},
	}
	return nil
}

// ListImages lists disk images
func (t *Test) ListImages(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			{
				ID:       "289120e4-ed12-431d-8d18-48f8a94bb05a",
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

// GetImage retrieves a disk image
func (t *Test) GetImage(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			{
				ID:       "289120e4-ed12-431d-8d18-48f8a94bb05a",
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

// DeleteImage deletes a disk image
func (t *Test) DeleteImage(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			{
				ID:       "289120e4-ed12-431d-8d18-48f8a94bb05a",
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

// RequestImage requests the fetching of a new images
func (t *Test) RequestImage(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			{
				ID:       "289120e4-ed12-431d-8d18-48f8a94bb05a",
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

// ListSnapshots retrieves a list of snapshots
func (t *Test) ListSnapshots(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			{
				ID:   fmt.Sprintf("%s@%s", filepath.Join("mistify", request.ID), "bar"),
				Size: 1024,
			},
		},
	}
	return nil
}

// GetSnapshot gets a single snapshot
func (t *Test) GetSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			{
				ID:   filepath.Join("mistify", request.ID),
				Size: 1024,
			},
		},
	}
	return nil
}

// CreateSnapshot creates a new snapshot
func (t *Test) CreateSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			{
				ID:   fmt.Sprintf("%s@%s", filepath.Join("mistify", request.ID), request.Dest),
				Size: 1024,
			},
		},
	}
	return nil
}

// DeleteSnapshot deletes a snapshot
func (t *Test) DeleteSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			{
				ID:   filepath.Join("mistify", request.ID),
				Size: 1024,
			},
		},
	}
	return nil
}

// RollbackSnapshot rolls the filesystem back to a snapshot
func (t *Test) RollbackSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			{
				ID:   filepath.Join("mistify", request.ID),
				Size: 1024,
			},
		},
	}
	return nil
}

// DownloadSnapshot downloads a snapshot via streaming
func (t *Test) DownloadSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprint(w, "foobar")
	return
}

// CreateContainer creates a container
func (t *Test) CreateContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	response.Guest = &client.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: "asdfasdfpoih",
	}
	return nil
}

// StartContainer starts a container
func (t *Test) StartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	response.Guest = &client.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: "asdfasdfpoih",
	}
	return nil
}

// RebootContainer reboots a container
func (t *Test) RebootContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	response.Guest = &client.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: "asdfasdfpoih",
	}
	return nil
}

// RestartContainer reboots a container
func (t *Test) RestartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	response.Guest = &client.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: "asdfasdfpoih",
	}
	return nil
}

// StopContainer stops a container
func (t *Test) StopContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	response.Guest = &client.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: "asdfasdfpoih",
	}
	return nil
}

// DeleteContainer deletes a container
func (t *Test) DeleteContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	response.Guest = &client.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: "asdfasdfpoih",
	}
	return nil
}

func main() {

	var port uint

	flag.UintVarP(&port, "port", "p", 9999, "listen port")
	flag.Parse()

	if err := logx.DefaultSetup("info"); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
		}).Fatal("failed to set up logging")
	}

	s, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.NewServer",
		}).Fatal(err)
	}

	test := &Test{}
	if err = s.RegisterService(test); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.RegisterService",
		}).Fatal(err)
	}
	s.HandleFunc("/snapshots/download", test.DownloadSnapshot)
	if err = s.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "rpc.Server.ListenAndServe",
		}).Fatal(err)
	}
}
