package rpc

import "github.com/fsouza/go-dockerclient"

type (
	// ContainerRequest is a container request to the Docker sub-agent
	ContainerRequest struct {
		ID   string      `json:"id"`   // Container ID
		Opts interface{} `json:"opts"` // Generic Options. Will need converting
	}

	// ContainerResponse is a container response from the Docker sub-agent
	ContainerResponse struct {
		Containers []*docker.Container `json:"containers"` // Slice of one or more containers
	}
)

// GetOpts returns the Opts property
func (creq *ContainerRequest) GetOpts() interface{} {
	return creq.Opts
}
