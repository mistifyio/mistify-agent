package rpc

type (

	// Image, volume, and snapshot should probably move to client package

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

	// Volume represents a ZFS Snapshot
	Snapshot struct {
		Id   string `json:"id"`   // Unique ID
		Size uint64 `json:"size"` // Size in MB
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
	}

	// VolumeResponse is a volume response from the Storage sub-agent
	VolumeResponse struct {
		Volumes []*Volume `json:"volumes"` //Volume slice for gets and lists. An empty slice is generally used for "not found"
	}

	// SnapshotRequest is a snapshot request for the Storage sub-agent
	SnapshotRequest struct {
		Id   string `json:"id"`   // Volume ID
		Dest string `json:"dest"` // Destination for clones, creates, etc
	}

	SnapshotResponse struct {
		Snapshots []*Snapshot `json:"snapshots"` // Snapshot slice for gets and lists. An empty slice is generally used for "not found"
	}
)
