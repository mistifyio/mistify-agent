package rpc

import "github.com/fsouza/go-dockerclient"

type (
	// ContainerImageRequest is an image request to the Docker sub-agent
	ContainerImageRequest struct {
		ID   string      `json:"id"`   // Image ID
		Name string      `json:"name"` // Image name
		Opts interface{} `json:"opts"` // Generic Options. Will need converting
	}

	// ContainerImageResponse is an image response from the Docker sub-agent
	ContainerImageResponse struct {
		Images []*docker.Image `json:"images"` // Slice of one or more images
	}
)

// GetOpts returns the Opts property
func (ireq *ContainerImageRequest) GetOpts() interface{} {
	return ireq.Opts
}

// GetLookup returns the string to look an image up by based on field priority
func (ireq *ContainerImageRequest) GetLookup(d string) string {
	if ireq.ID != "" {
		return ireq.ID
	}
	if ireq.Name != "" {
		return ireq.Name
	}
	return d
}
