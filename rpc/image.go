package rpc

type (

	// Image represents a ZFS ZVOL snapshot used for creating VM disks
	Image struct {
		Id       string `json:"id"`       // Unique ID
		Volume   string `json:"volume"`   // Imported ZVOL
		Snapshot string `json:"snapshot"` // ZVOL Snapshot
		Size     uint64 `json:"size"`     // Size in MB
		Status   string `json:"status"`   // current status of the Image: pending, complete, etc
	}

	// Volume represents a ZFS ZVOL
	Volume struct {
		Id     string `json:"id"`     // Unique ID
		Size   uint64 `json:"size"`   // Size in MB
		Device string `json:"device"` // Device in /dev to use
	}

	// ImageRequest is an image request to the Storage sub-agent
	ImageRequest struct {
		Id     string `json:"id"`     // Image ID
		Dest   string `json:"dest"`   // Destination for clones, etc
		Source string `json:"source"` // Source for fetches. Generally a URL
	}

	// ImageRequest is an image response from the Storage sub-agent
	ImageResponse struct {
		Images []*Image `json:"images"` //Image slice for gets and lists. An empty slice is generally used for "not found"
	}

	// VolumeRequest is a volume request to the Storage sub-agent. Currently, only create and delete are used.
	VolumeRequest struct {
		Id   string `json:"id"`   // Volume ID
		Size uint64 `json:"size"` //  Size in MB
		Dest string `json:"dest"` // Destination for snapshots, etc
	}

	// VolumeResponse is a volume response from the Storage sub-agent
	VolumeResponse struct {
		Volumes []*Volume `json:"volumes"` //Volume slice for gets and lists. An empty slice is generally used for "not found"
	}
)
