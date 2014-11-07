// This is a test Mistify sub-agent that implements all the actions the agent may call.
// It returns the guest received for Guest requests and fake metrics for metric requests.
// A sub-agent should generally only have one area of concern and do one thing well.  This allows
// sub-agents to be composited in various ways.
package main

import (
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	"log"
	"net/http"
	"os"
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

func main() {

	var port int
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
	s.RegisterService(&Test{})
	log.Fatal(s.ListenAndServe())
}
