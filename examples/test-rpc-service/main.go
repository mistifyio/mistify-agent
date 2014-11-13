// This is a test Mistify sub-agent that implements all the actions the agent may call.
// It returns the guest received for Guest requests and fake metrics for metric requests.
// A sub-agent should generally only have one area of concern and do one thing well.  This allows
// sub-agents to be composited in various ways.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
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

// Retrieve CPU metrics.  Currently, only one sub-agent service is called for this.
func (t *Test) CpuMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	*response = rpc.GuestMetricsResponse{
		Guest: request.Guest,
		Type:  "cpu",
		CPU: []*client.GuestCpuMetrics{
			&client.GuestCpuMetrics{},
		},
	}
	return nil
}

// Retrieve Disk metrics.  Currently, only one sub-agent service is called for this.
func (t *Test) DiskMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	*response = rpc.GuestMetricsResponse{
		Guest: request.Guest,
		Type:  "disk",
		Disk: map[string]*client.GuestDiskMetrics{
			"vda": &client.GuestDiskMetrics{
				Disk: "vda",
			},
		},
	}
	return nil
}

// Retrieve Network metrics.  Currently, only one sub-agent service is called for this.
func (t *Test) NicMetrics(r *http.Request, request *rpc.GuestMetricsRequest, response *rpc.GuestMetricsResponse) error {
	*response = rpc.GuestMetricsResponse{
		Guest: request.Guest,
		Type:  "disk",
		Nic: map[string]*client.GuestNicMetrics{
			"net0": &client.GuestNicMetrics{
				Name: "net0",
			},
		},
	}
	return nil
}

func (t *Test) ListImages(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			&rpc.Image{
				Id:       filepath.Join("mistify", "images", "foo"),
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

func (t *Test) GetImage(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			&rpc.Image{
				Id:       filepath.Join("mistify", "images", request.Id),
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

func (t *Test) DeleteImage(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			&rpc.Image{
				Id:       filepath.Join("mistify", "images", request.Id),
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

func (t *Test) RequestImage(r *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	*response = rpc.ImageResponse{
		Images: []*rpc.Image{
			&rpc.Image{
				Id:       filepath.Join("mistify", "images", "foo"),
				Volume:   "",
				Snapshot: "",
				Size:     1024,
				Status:   "complete",
			},
		},
	}
	return nil
}

func (t *Test) ListSnapshots(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			&rpc.Snapshot{
				Id:   fmt.Sprintf("%s@%s", filepath.Join("mistify", request.Id), "bar"),
				Size: 1024,
			},
		},
	}
	return nil
}

func (t *Test) GetSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			&rpc.Snapshot{
				Id:   filepath.Join("mistify", request.Id),
				Size: 1024,
			},
		},
	}
	return nil
}

func (t *Test) CreateSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			&rpc.Snapshot{
				Id:   fmt.Sprintf("%s@%s", filepath.Join("mistify", request.Id), request.Dest),
				Size: 1024,
			},
		},
	}
	return nil
}

func (t *Test) DeleteSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			&rpc.Snapshot{
				Id:   filepath.Join("mistify", request.Id),
				Size: 1024,
			},
		},
	}
	return nil
}

func (t *Test) RollbackSnapshot(r *http.Request, request *rpc.SnapshotRequest, response *rpc.SnapshotResponse) error {
	*response = rpc.SnapshotResponse{
		Snapshots: []*rpc.Snapshot{
			&rpc.Snapshot{
				Id:   filepath.Join("mistify", request.Id),
				Size: 1024,
			},
		},
	}
	return nil
}

func (t *Test) DownloadSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprint(w, "foobar")
	return
}

func main() {

	var port uint
	var h bool

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.UintVar(&port, []string{"p", "#port", "-port"}, 9999, "listen port")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}
	s, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}
	test := &Test{}
	s.RegisterService(test)
	s.HandleFunc("/snapshots/download", test.DownloadSnapshot)
	log.Fatal(s.ListenAndServe())
}
