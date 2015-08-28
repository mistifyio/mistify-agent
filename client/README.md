# client

[![client](https://godoc.org/github.com/mistifyio/mistify-agent/client?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent/client)

Package client provides a simple client for the Mistify Agent.

## Usage

#### type Client

```go
type Client struct {
	Config Config
}
```

Client for the Mistify Agent

#### func  NewClient

```go
func NewClient(config *Config) (*Client, error)
```
NewClient returns a new client

#### func (*Client) CreateGuest

```go
func (c *Client) CreateGuest(guest *Guest) (*Guest, error)
```
CreateGuest requests creation of a new guest

#### func (*Client) GetGuest

```go
func (c *Client) GetGuest(id string) (*Guest, error)
```
GetGuest requests creation of a guest

#### func (*Client) ListGuests

```go
func (c *Client) ListGuests() (GuestSlice, error)
```
ListGuests gets a list of guests

#### type Config

```go
type Config struct {
	// Address is the address of the Mistify Agent
	Address string

	// Scheme is the URI scheme for the Mistify Agent
	Scheme string

	// HTTPClient is the client to use. Default will be
	// used if not provided.
	HTTPClient *http.Client
}
```

Config is used to configure the creation of a client

#### func  DefaultConfig

```go
func DefaultConfig() *Config
```
DefaultConfig returns a default configuration for the client

#### type Disk

```go
type Disk struct {
	Bus    string `json:"bus"`    // the type of disk device to emulate. "ide", "scsi", "sata", virtio"
	Device string `json:"device"` // target device inside the guest, ie "vda", "sda", "hda", etc
	Size   uint64 `json:"size"`   // size in MB.  On create, this is not used for image based disks.
	Volume string `json:"volume"` // zfs zvol
	Image  string `json:"image"`  // which image to clone.  If this is not set, then a blank zvol is created
	Source string `json:"source"` // the device name: /dev/zvol/...
}
```

Disk is a guest storage disk

#### type Guest

```go
type Guest struct {
	Id       string            `json:"id"`
	Type     string            `json:"type,omitempty"`
	Image    string            `json:"image,omitempty"`
	Nics     []Nic             `json:"nics,omitempty"`
	Disks    []Disk            `json:"disks,omitempty"`
	State    string            `json:"state,omitempty"`  //current State
	Memory   uint              `json:"memory,omitempty"` // Memory in MB
	Cpu      uint              `json:"cpu,omitempty"`    // number of Virtual CPU's
	VNC      int               `json:"vnc,omitempty"`    // VNC port
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

Guest is a guest virtual machine +gen * slice:"Where,Each,SortBy" set

#### type GuestCPUMetrics

```go
type GuestCPUMetrics struct {
	CpuTime  float64 `json:"cpu_time"`
	VcpuTime float64 `json:"vcpu_time"`
}
```

GuestCPUMetrics is a set of metrics on a guest's cpu

#### type GuestDiskMetrics

```go
type GuestDiskMetrics struct {
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
```

GuestDiskMetrics is a set of metrics on a guest's storage disk

#### type GuestNicMetrics

```go
type GuestNicMetrics struct {
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
```

GuestNicMetrics is a set of metrics on a guests's nic

#### type GuestSet

```go
type GuestSet map[*Guest]struct{}
```

GuestSet is the primary type that represents a set

#### func  NewGuestSet

```go
func NewGuestSet(a ...*Guest) GuestSet
```
NewGuestSet creates and returns a reference to an empty set.

#### func (GuestSet) Add

```go
func (set GuestSet) Add(i *Guest) bool
```
Add adds an item to the current set if it doesn't already exist in the set.

#### func (GuestSet) Cardinality

```go
func (set GuestSet) Cardinality() int
```
Cardinality returns how many items are currently in the set.

#### func (*GuestSet) Clear

```go
func (set *GuestSet) Clear()
```
Clear clears the entire set to be the empty set.

#### func (GuestSet) Clone

```go
func (set GuestSet) Clone() GuestSet
```
Clone returns a clone of the set. Does NOT clone the underlying elements.

#### func (GuestSet) Contains

```go
func (set GuestSet) Contains(i *Guest) bool
```
Contains determines if a given item is already in the set.

#### func (GuestSet) ContainsAll

```go
func (set GuestSet) ContainsAll(i ...*Guest) bool
```
ContainsAll determines if the given items are all in the set

#### func (GuestSet) Difference

```go
func (set GuestSet) Difference(other GuestSet) GuestSet
```
Difference returns a new set with items in the current set but not in the other
set

#### func (GuestSet) Equal

```go
func (set GuestSet) Equal(other GuestSet) bool
```
Equal determines if two sets are equal to each other. If they both are the same
size and have the same items they are considered equal. Order of items is not
relevent for sets to be equal.

#### func (GuestSet) Intersect

```go
func (set GuestSet) Intersect(other GuestSet) GuestSet
```
Intersect returns a new set with items that exist only in both sets.

#### func (GuestSet) IsSubset

```go
func (set GuestSet) IsSubset(other GuestSet) bool
```
IsSubset determines if every item in the other set is in this set.

#### func (GuestSet) IsSuperset

```go
func (set GuestSet) IsSuperset(other GuestSet) bool
```
IsSuperset determines if every item of this set is in the other set.

#### func (GuestSet) Iter

```go
func (set GuestSet) Iter() <-chan *Guest
```
Iter returns a channel of type *Guest that you can range over.

#### func (GuestSet) Remove

```go
func (set GuestSet) Remove(i *Guest)
```
Remove allows the removal of a single item in the set.

#### func (GuestSet) SymmetricDifference

```go
func (set GuestSet) SymmetricDifference(other GuestSet) GuestSet
```
SymmetricDifference returns a new set with items in the current set or the other
set but not in both.

#### func (GuestSet) ToSlice

```go
func (set GuestSet) ToSlice() []*Guest
```
ToSlice returns the elements of the current set as a slice

#### func (GuestSet) Union

```go
func (set GuestSet) Union(other GuestSet) GuestSet
```
Union returns a new set with all items in both sets.

#### type GuestSlice

```go
type GuestSlice []*Guest
```

GuestSlice is a slice of type *Guest. Use it where you would use []*Guest.

#### func (GuestSlice) Each

```go
func (rcv GuestSlice) Each(fn func(*Guest))
```
Each iterates over GuestSlice and executes the passed func against each element.
See: http://clipperhouse.github.io/gen/#Each

#### func (GuestSlice) SortBy

```go
func (rcv GuestSlice) SortBy(less func(*Guest, *Guest) bool) GuestSlice
```
SortBy returns a new ordered GuestSlice, determined by a func defining ‘less’.
See: http://clipperhouse.github.io/gen/#SortBy

#### func (GuestSlice) Where

```go
func (rcv GuestSlice) Where(fn func(*Guest) bool) (result GuestSlice)
```
Where returns a new GuestSlice whose elements return true for func. See:
http://clipperhouse.github.io/gen/#Where

#### type Nic

```go
type Nic struct {
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
```

Nic is a guest network interface controller

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
