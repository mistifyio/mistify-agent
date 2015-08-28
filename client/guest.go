package client

type (
	// Guest is a guest virtual machine
	// +gen * slice:"Where,Each,SortBy" set
	Guest struct {
		ID       string            `json:"id"`
		Type     string            `json:"type,omitempty"`
		Image    string            `json:"image,omitempty"`
		Nics     []Nic             `json:"nics,omitempty"`
		Disks    []Disk            `json:"disks,omitempty"`
		State    string            `json:"state,omitempty"`  //current State
		Memory   uint              `json:"memory,omitempty"` // Memory in MB
		CPU      uint              `json:"cpu,omitempty"`    // number of Virtual CPU's
		VNC      int               `json:"vnc,omitempty"`    // VNC port
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	// Nic is a guest network interface controller
	Nic struct {
		Name    string `json:"name,omitempty"`
		Network string `json:"network"`
		Model   string `json:"model"`
		Mac     string `json:"mac,omitempty"`
		Address string `json:"address"`
		Netmask string `json:"netmask"`
		Gateway string `json:"gateway"`
		Device  string `json:"device,omitempty"`
		VLANs   []int  `json:"vlans"`
	}

	// Disk is a guest storage disk
	Disk struct {
		Bus    string `json:"bus"`    // the type of disk device to emulate. "ide", "scsi", "sata", virtio"
		Device string `json:"device"` // target device inside the guest, ie "vda", "sda", "hda", etc
		Size   uint64 `json:"size"`   // size in MB.  On create, this is not used for image based disks.
		Volume string `json:"volume"` // zfs zvol
		Image  string `json:"image"`  // which image to clone.  If this is not set, then a blank zvol is created
		Source string `json:"source"` // the device name: /dev/zvol/...
	}

	// GuestDiskMetrics is a set of metrics on a guest's storage disk
	GuestDiskMetrics struct {
		Disk       string  `json:"disk"`
		ReadOps    int64   `json:"read_ops"`
		ReadBytes  int64   `json:"read_bytes"`
		ReadTime   float64 `json:"read_time"`
		WriteOps   int64   `json:"write_ops"`
		WriteBytes int64   `json:"write_bytes"`
		WriteTime  float64 `json:"write_time"`
		FlushOps   int64   `json:"flush_ops"`
		FlushTime  float64 `json:"flush_time"`
	}

	// GuestCPUMetrics is a set of metrics on a guest's cpu
	GuestCPUMetrics struct {
		CPUTime  float64 `json:"cpu_time"`
		VCPUTime float64 `json:"vcpu_time"`
	}

	// GuestNicMetrics is a set of metrics on a guests's nic
	GuestNicMetrics struct {
		Name      string `json:"name"`
		RxBytes   int64  `json:"rx_bytes"`
		RxPackets int64  `json:"rx_packets"`
		RxErrs    int64  `json:"rx_errors"`
		RxDrop    int64  `json:"rx_drops"`
		TxBytes   int64  `json:"tx_bytes"`
		TxPackets int64  `json:"tx_packets"`
		TxErrs    int64  `json:"tx_errors"`
		TxDrop    int64  `json:"tx_drops"`
	}
)
