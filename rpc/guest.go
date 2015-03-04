package rpc

import (
	"github.com/mistifyio/mistify-agent/client"
)

type (
	// GuestRequest is a request to a sub-agent
	GuestRequest struct {
		Guest  *client.Guest     `json:"guest"`          // Guest
		Action string            `json:"action"`         // Action
		Args   map[string]string `json:"args,omitempty"` // Opaque, optional arguments
	}

	// GuestResponse is a response from a sub-agent
	GuestResponse struct {
		Guest   *client.Guest `json:"guest"`             // Guest, possibly modified
		Message string        `json:"message,omitempty"` // Any informational message
		Retry   int           `json:"retry,omitempty"`   // instruct the agent to retry after this many second. Not yet implemented
	}

	// GuestMetricsRequest is a request for guest metrics
	GuestMetricsRequest struct {
		Guest *client.Guest     `json:"guest"`          // Guest
		Type  string            `json:"type"`           // type of metric desired
		Args  map[string]string `json:"args,omitempty"` // Opaque, optional arguments
	}

	// GuestMetricsResponse is a response of guest metrics
	GuestMetricsResponse struct {
		Guest *client.Guest                       `json:"guest"`          // Guest - in general this should not be modified
		Type  string                              `json:"type"`           // Type of metrics returned
		Disk  map[string]*client.GuestDiskMetrics `json:"disk,omitempty"` // Disk metrics
		Nic   map[string]*client.GuestNicMetrics  `json:"nic,omitempty"`  // Network Interface metrics
		CPU   []*client.GuestCpuMetrics           `json:"cpu,omitempty"`  // CPU metrics
	}
)
