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
- **GuestMetricsRequest** - RPC request - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestMetricsRequest
- **GuestMetricsResponse** - RPC response - http://godoc.org/github.com/mistifyio/mistify-agent/rpc#GuestMetricsResponse

# API calls #

|||
|---|---|
|REST | GET /metadata|
|Description | Hypervisor level metadata|
|Output|hash of key/value pairs|


|||
|---|---|
|REST | PATCH /metadata|
|Description | Hypervisor level metadata|
|Input|key/value pairs (both must be strings). Values to be removed can be set to null|
|Output|hash of key/value pairs|


|||
|---|---|
|REST | GET /guests|
|Description | Get list of guests|
|Output|Array of **Guest**|

Note: the Agent keeps data about the guests in its data store, this is the data that is returned. Sub-agents should add any metadata desired to the Guest metadata.

|||
|---|---|
|REST | POST /guests|
|Description | Create a guest|
|Input|**Guest** though certain fields should not be set, such as ID, network/disk devices/etc|
|Output|**Guest**|

Sample Guest request body:
```json
{
}
```

|||
|---|---|
|REST | GET /guests/{id}|
|Description | get a single guest|
|Output|**Guest**|

|||
|---|---|
|REST | GET /guests/{id}/metadata|
|Description | get  guest metadata|
|Output|key/value pairs|

|||
|---|---|
|REST | PATCH /guests/{id}/metadata|
|Description | get  guest metadata|
|Input|key/value pairs (both must be strings). Values to be removed can be set to null|
|Output|hash of key/value pairs|

|||
|---|---|
|REST | GET /guests/{id}/metrics/cpu|
|Description | get guest CPU metrics|
|Output|**GuestCpuMetrics**|
|RPC|This calls to the cpu metrics service. This is a single call, there is no pipeline (currently)|
|_RPC Input_|**GuestMetricsRequest**|
|_RPC Output_|**GuestMetricsResponse** with only CPU metrics|


|||
|---|---|
|REST | GET /guests/{id}/metrics/nic|
|Description | get guest NIC metrics|
|Output|**GuestNicMetrics**|
|RPC|This calls to the nic metrics service. This is a single call, there is no pipeline (currently)|
|_RPC Input_|**GuestMetricsRequest**|
|_RPC Output_|**GuestMetricsResponse** with only NIC metrics|

|||
|---|---|
|REST | GET /guests/{id}/metrics/disk|
|Description | get guest disk metrics|
|Output|**GuestDiskMetrics**|
|RPC|This calls to the disk metrics service. This is a single call, there is no pipeline (currently)|
|_RPC Input_|**GuestMetricsRequest**|
|_RPC Output_|**GuestMetricsResponse** with only disk metrics|

|||
|---|---|
|REST | POST /guests/{id}/shutdown|
|Description | perform soft shutdown of the guest. may require guest support|
|Output|**Guest**|
|RPC|This calls the _shutdown_ action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|


|||
|---|---|
|REST | POST /guests/{id}/reboot|
|Description | perform reboot of the guest.|
|Output|**Guest**|
|RPC|This calls the reboot action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|



|||
|---|---|
|REST | POST /guests/{id}/restart|
|Description | perform restart of the guest. This is a _hard_ reset.|
|Output|**Guest**|
|RPC|This calls the restart action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|



|||
|---|---|
|REST | POST /guests/{id}/poweroff|
|Description | perform poweroff of the guest.|
|Output|**Guest**|
|RPC|This calls the poweroff action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|



|||
|---|---|
|REST | POST /guests/{id}/start|
|Description | perform start of the guest if it is currently off, shutdown, or suspended|
|Output|**Guest**|
|RPC|This calls the start action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|



|||
|---|---|
|REST | POST /guests/{id}/suspend|
|Description | perform suspend of the guest.|
|Output|**Guest**|
|RPC|This calls the suspend action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|



|||
|---|---|
|REST | DELETE /guests/{id}|
|Description | perform delete of the guest. After succesfully deleting the guest, the agent will remove it from its data store.|
|Output|**Guest**|
|RPC|This calls the delete action which has both a synchronous and asynchronous pipeline|
|_RPC Input_|**GuestRequest**|
|_RPC Output_|**GuestResponse**|
