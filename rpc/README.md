# rpc

[![rpc](https://godoc.org/github.com/mistifyio/mistify-agent/rpc?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent/rpc)

Package rpc defines the JSON-RPC over HTTP types the Agent uses to communicate
with sub-agents. It also contains various helpers for writing sub-agents.

## Usage

```go
const (
	// RPCPath is the URI endpoint that the Agent posts to for sub-agent communication.
	RPCPath = "/_mistify_RPC_"
)
```

#### type Client

```go
type Client struct {
	URL string
}
```

Client is a simple JSON-RPC over HTTP client used by the agent.

#### func  NewClient

```go
func NewClient(port uint, path string) (*Client, error)
```
NewClient create a new client. This only communicates with 127.0.0.1

#### func (*Client) Do

```go
func (c *Client) Do(method string, request interface{}, response interface{}) error
```
Do calls an RPC method

#### func (*Client) DoRaw

```go
func (c *Client) DoRaw(request interface{}, rw http.ResponseWriter)
```
DoRaw calls a service and proxies the response

#### type Codec

```go
type Codec struct {
	*json.Codec
}
```

Codec is a wrapper for the json.Codec

#### func  NewCodec

```go
func NewCodec() *Codec
```
NewCodec creates a new Codec

#### func (*Codec) NewRequest

```go
func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest
```
NewRequest creates a new request from the codec

#### type ContainerImageRequest

```go
type ContainerImageRequest struct {
	ID   string      `json:"id"`   // Image ID
	Name string      `json:"name"` // Image name
	Opts interface{} `json:"opts"` // Generic Options. Will need converting
}
```

ContainerImageRequest is an image request to the Docker sub-agent

#### func (*ContainerImageRequest) GetLookup

```go
func (ireq *ContainerImageRequest) GetLookup(d string) string
```
GetLookup returns the string to look an image up by based on field priority

#### func (*ContainerImageRequest) GetOpts

```go
func (ireq *ContainerImageRequest) GetOpts() interface{}
```
GetOpts returns the Opts property

#### type ContainerImageResponse

```go
type ContainerImageResponse struct {
	Images []*docker.Image `json:"images"` // Slice of one or more images
}
```

ContainerImageResponse is an image response from the Docker sub-agent

#### type ContainerRequest

```go
type ContainerRequest struct {
	ID   string      `json:"id"`   // Container ID
	Opts interface{} `json:"opts"` // Generic Options. Will need converting
}
```

ContainerRequest is a container request to the Docker sub-agent

#### func (*ContainerRequest) GetOpts

```go
func (creq *ContainerRequest) GetOpts() interface{}
```
GetOpts returns the Opts property

#### type ContainerResponse

```go
type ContainerResponse struct {
	Containers []*docker.Container `json:"containers"` // Slice of one or more containers
}
```

ContainerResponse is a container response from the Docker sub-agent

#### type GuestMetricsRequest

```go
type GuestMetricsRequest struct {
	Guest *client.Guest     `json:"guest"`          // Guest
	Type  string            `json:"type"`           // type of metric desired
	Args  map[string]string `json:"args,omitempty"` // Opaque, optional arguments
}
```

GuestMetricsRequest is a request for guest metrics

#### type GuestMetricsResponse

```go
type GuestMetricsResponse struct {
	Guest *client.Guest                       `json:"guest"`          // Guest - in general this should not be modified
	Type  string                              `json:"type"`           // Type of metrics returned
	Disk  map[string]*client.GuestDiskMetrics `json:"disk,omitempty"` // Disk metrics
	Nic   map[string]*client.GuestNicMetrics  `json:"nic,omitempty"`  // Network Interface metrics
	CPU   []*client.GuestCpuMetrics           `json:"cpu,omitempty"`  // CPU metrics
}
```

GuestMetricsResponse is a response of guest metrics

#### type GuestRequest

```go
type GuestRequest struct {
	Guest  *client.Guest     `json:"guest"`          // Guest
	Action string            `json:"action"`         // Action
	Args   map[string]string `json:"args,omitempty"` // Opaque, optional arguments
}
```

GuestRequest is a request to a sub-agent

#### type GuestResponse

```go
type GuestResponse struct {
	Guest   *client.Guest `json:"guest"`             // Guest, possibly modified
	Message string        `json:"message,omitempty"` // Any informational message
	Retry   int           `json:"retry,omitempty"`   // instruct the agent to retry after this many second. Not yet implemented
}
```

GuestResponse is a response from a sub-agent

#### type Image

```go
type Image struct {
	Id       string `json:"id"`       // Unique ID
	Volume   string `json:"volume"`   // Imported ZVOL
	Snapshot string `json:"snapshot"` // ZVOL Snapshot
	Size     uint64 `json:"size"`     // Size in MB
	Status   string `json:"status"`   // current status of the Image: pending, complete, etc
}
```

Image represents a ZFS ZVOL snapshot used for creating VM disks

#### type ImageRequest

```go
type ImageRequest struct {
	Id     string `json:"id"`     // Image ID
	Dest   string `json:"dest"`   // Destination for clones, etc
	Source string `json:"source"` // Source for fetches. Generally a URL
}
```

ImageRequest is an image request to the Storage sub-agent

#### type ImageResponse

```go
type ImageResponse struct {
	Images []*Image `json:"images"` //Image slice for gets and lists. An empty slice is generally used for "not found"
}
```

ImageResponse is an image response from the Storage sub-agent

#### type Server

```go
type Server struct {
	RPCServer  *rpc.Server
	HTTPServer *http.Server
	Router     *mux.Router
	Chain      alice.Chain
}
```

Server is a basic JSON-RPC over HTTP server.

#### func  NewServer

```go
func NewServer(port uint) (*Server, error)
```
NewServer creates an JSON-RPC HTTP server bound to 127.0.0.1. This answers RPC
requests on the Mistify RPC Path. This server logs to STDOUT and also presents
pprof on /debug/pprof/

#### func (*Server) Handle

```go
func (s *Server) Handle(pattern string, handler http.Handler)
```
Handle is a helper for registering a handler for a given path.

#### func (*Server) HandleFunc

```go
func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
```
HandleFunc is a helper for registering a handler function for a given path

#### func (*Server) ListenAndServe

```go
func (s *Server) ListenAndServe() error
```
ListenAndServe is a helper for starting the HTTP service. This generally does
not return.

#### func (*Server) RegisterService

```go
func (s *Server) RegisterService(receiver interface{}) error
```
RegisterService is a helper for registering a new RPC service

#### type Snapshot

```go
type Snapshot struct {
	Id   string `json:"id"`   // Unique ID
	Size uint64 `json:"size"` // Size in MB
}
```

Snapshot represents a ZFS Snapshot

#### type SnapshotRequest

```go
type SnapshotRequest struct {
	Id                string `json:"id"`                // Volume ID
	Dest              string `json:"dest"`              // Destination for clones, creates, etc
	Recursive         bool   `json:"recursive"`         // Recursively create snapshots for all descendent file systems
	DestroyMoreRecent bool   `json:"destroyMoreRecent"` // Destroy more recent snapshots when rolling back
}
```

SnapshotRequest is a snapshot request for the Storage sub-agent

#### type SnapshotResponse

```go
type SnapshotResponse struct {
	Snapshots []*Snapshot `json:"snapshots"` // Snapshot slice for gets and lists. An empty slice is generally used for "not found"
}
```

SnapshotResponse is a snapshot response for the Storage sub-agent

#### type Volume

```go
type Volume struct {
	Id     string `json:"id"`     // Unique ID
	Size   uint64 `json:"size"`   // Size in MB
	Device string `json:"device"` // Device in /dev to use
}
```

Volume represents a ZFS ZVOL

#### type VolumeRequest

```go
type VolumeRequest struct {
	Id   string `json:"id"`   // Volume ID
	Size uint64 `json:"size"` //  Size in MB
}
```

VolumeRequest is a volume request to the Storage sub-agent. Currently, only
create and delete are used.

#### type VolumeResponse

```go
type VolumeResponse struct {
	Volumes []*Volume `json:"volumes"` //Volume slice for gets and lists. An empty slice is generally used for "not found"
}
```

VolumeResponse is a volume response from the Storage sub-agent

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
