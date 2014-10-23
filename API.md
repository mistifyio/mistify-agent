This describes the API as currently implemented.

Many calls to the REST API endpoint will result in calls to sub-agents.  Some of these calls may result in both a synchronous and asynchronous pipeline to be executed.

# Data Structures #
Rather than recreate the data structures here, we will reference them as declared in the agent code documentation.  All request/response bodies are JSON.

- **Guest** - a vm/container/zone - http://godoc.org/github.com/mistifyio/mistify-agent/client#Guest
- **GuestCpuMetrics** - CPU metrics - http://godoc.org/github.com/mistifyio/mistify-agent/client#GuestCpuMetrics
- **GuestDiskMetrics** -  Disk metrics - http://godoc.org/github.com/mistifyio/mistify-agent/client#GuestDiskMetrics
- **GuestNicMetrics** - Network Interface metrics - http://godoc.org/github.com/mistifyio/mistify-agent/client#GuestNicMetrics
- **GuestRequest** - RPC request - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestRequest
- **GuestResponse** - RPC response - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestResponse
- **Image** - guest disk image - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#Image
- **ImageRequest** - RPC request - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#ImageRequest
- **ImageResponse** - RPC response - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#ImageResponse
- **Volume** - guest disk volume - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#Volume
- **VolumeRequest** - RPC request - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#VolumeRequest
- **VolumeResponse** - RPC response - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#VolumeResponse

# API calls #

-|-
-|-
 REST | GET /metadata
Description | Hypervisor level metadata

