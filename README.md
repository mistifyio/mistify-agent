# agent

[![agent](https://godoc.org/github.com/mistifyio/mistify-agent?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent)

Mistify Agent is the core agent for managing a hypervisor. It runs local on the
hypervisor and exposes an HTTP API for managing virtual machines.


### General Architecture

Mistify Agent provides a REST-ish API for manipulating a hypervisor and defines
a general set of actions. It is primarilly an interface to the sub-agents, which
do most of the work. Sub-agents are assumed to be running on the same host as
the agent and communication is done via JSON-RPC over HTTP.

The data structures for the REST API are defined in:
http://godoc.org/github.com/mistifyio/mistify-agent/client

The data structures for the RPC API are defined in:
http://godoc.org/github.com/mistifyio/mistify-agent/rpc

Sub-agents do not have to be written in Go, but the agent does provide helpers
for easily creating them. All of the official sub-agents are written in Go.

There are two types of sub-agents:

* images/storage - Mistify will provide an opinionated version that uses ZFS
ZVOLs for guest disks.

* guest "actions" - create/delete/reboot/etc.

Only one storage sub-agent is used, but multiple guest sub-agents may be used.


### Actions

While there is a defined set of actions, the work performed is configurable.
Each action has a pipeline, a series of one or more steps that need to be
performed to complete the action, configurable in the config file. All steps
must succeed, in order, for an action to be considered successful.

There are three action types:

* Info - Information retrieval actions, such as getting a list of guests, called
synchronously at request time. A JSON result is returned to the requesting
client when the pipeline is complete.

* Async - Modification actions, such as rebooting a guest, called
asynchronously. One action per guest is performed at a time, while the rest are
queued. A response containing the job id in the header X-Guest-Job-ID is
returned after queueing the action, which can be used to check the status at a
later time.

* Stream - Data retrieval, such as downloading a zfs snapshot, called
synchronously at request time. Rather than a JSON response, data is streamed
back in chunks.

Valid actions are defined in:
http://godoc.org/github.com/mistifyio/mistify-agent/config

### HTTP API Endpoints

    /debug/pprof
    	* GET

    /debug/pprof/{profileName}
    	* GET

    /debug/pprof/cmdline
    	* GET

    /debug/pprof/symbol
    	* GET

    /debug/pprof/symbol
    	* GET

    /metadata
    	* GET   - Retrieve the hypervisor's metadata
    	* PATCH - Modify the hypervisor's metadata

    /images
    	* GET  - Retrieve a list of disk images
    	* POST - Fetch a disk image

    /images/{imageID}
    	* GET    - Retrieve information about a disk image
    	* DELETE - Delete a disk image

    /container_images
    	* GET  - Retrieve a list of container images
    	* POST - Fetch a container image

    /container_images/{imageID}
    	* GET    -  Retrieve information about a container image
    	* DELETE - Delete a container image

    /guests
    	* GET  - Retrieve a list of guests
    	* POST - Create a new guest

    /guests/{guestID}
    	* GET - Retrieve information about a guest

    /guests/{guestID}/jobs
    	* GET - Retrieve a list of recent action jobs for the guest

    /guests/{guestID}/jobs/{jobID}
    	* GET - Retrieve information about a specific action job

    /guests/{guestID}/metadata
    	* GET   - Retrieve a guest's metadata
    	* PATCH - Modify the guest's metadata

    /guests/{guestID}/metrics/cpu
    	* GET - Retrieve guest CPU metrics

    /guests/{guestID}/metrics/disk
    	* GET - Retrieve guest disk metrics

    /guests/{guestID}/metrics/nic
    	* GET - Retrieve guest NIC metrics

    /guests/{guestID}/{actionName}
    	Actions: shutdown, reboot, restart, poweroff, start, suspend, delete
    	* GET - Perform the specified action for the guest

    /guests/{guestID}/snapshots
    /guests/{guestID}/disks/{diskID}/snapshots
    	* GET  - Retrieve a list of snapshots
    	* POST - Create a new snapshot

    /guests/{guestID}/snapshots/{snapshotName}
    /guests/{guestID}/disks/{diskID}/snapshots/{snapshotName}
    	* GET    - Retrieve information about a snapshot
    	* DELETE - Delete a snapshot

    /guests/{guestID}/snapshots/{snapshotName}/rollback
    /guests/{guestID}/disks/{diskID}/snapshots/{snapshotName}/rollback
    	* POST - Roll back to the snapshot

    /guests/{guestID}/snapshots/{snapshotName}/download
    /guests/{guestID}/disks/{diskID}/snapshots/{snapshotName}/download
    	* GET - Download the snapshot


### Contributing

See the guidelines:
https://github.com/mistifyio/mistify-agent/blob/master/CONTRIBUTING.md

## Usage

```go
const (
	// MaxLoggedJobs configures how many jobs to prune the log to
	MaxLoggedJobs int = 1000
	// Queued is the queued job status
	Queued JobStatus = "Queued"
	// Running is the running job status
	Running JobStatus = "Running"
	// Complete is the complete job status
	Complete JobStatus = "Complete"
	// Errored is the errored job status
	Errored JobStatus = "Error"
)
```

```go
var (
	// ErrNotFound is the error for a resouce not found
	ErrNotFound = errors.New("not found")
)
```

#### func  AttachProfiler

```go
func AttachProfiler(router *mux.Router)
```
AttachProfiler enables debug profiling exposed on http api endpoints

#### func  LogRunnerError

```go
func LogRunnerError(guestID string, runnerName string, pipelineID string, logLine string)
```
LogRunnerError writes error logs

#### func  LogRunnerInfo

```go
func LogRunnerInfo(guestID string, runnerName string, pipelineID string, logLine string)
```
LogRunnerInfo writes informational logs

#### func  Run

```go
func Run(ctx *Context, address string) error
```
Run prepares and runs the http server

#### type Action

```go
type Action struct {
	Name   string
	Type   config.ActionType
	Stages []*Stage
}
```

Action is a full set of stage templates required to complete an action

#### func (*Action) GeneratePipeline

```go
func (action *Action) GeneratePipeline(request interface{}, response interface{}, rw http.ResponseWriter, done chan error) *Pipeline
```
GeneratePipeline creates an instance of Pipeline based on an action's stages and
supplied request & response. It is returned so that any additional modifications
(such as adding stage args to requests) can be made before running if needed.

#### type Chain

```go
type Chain struct {
	alice.Chain
}
```

Chain is a middleware chain

#### func (*Chain) GuestActionWrapper

```go
func (c *Chain) GuestActionWrapper(actionName string) http.HandlerFunc
```
GuestActionWrapper wraps an HTTP request with a Guest action to avoid duplicated
code

#### func (*Chain) GuestRunnerWrapper

```go
func (c *Chain) GuestRunnerWrapper(fn func(*HTTPRequest) *HTTPErrorMessage) http.HandlerFunc
```
GuestRunnerWrapper is a middleware that retrieves the runner for a guest

#### func (*Chain) RequestWrapper

```go
func (c *Chain) RequestWrapper(fn func(*HTTPRequest) *HTTPErrorMessage) http.HandlerFunc
```
RequestWrapper turns a basic http request into an HTTPRequest

#### type Context

```go
type Context struct {
	Config   *config.Config
	Actions  map[string]*Action
	Services map[string]*Service

	GuestRunners     map[string]*GuestRunner
	GuestRunnerMutex sync.Mutex
	JobLog           *JobLog
}
```

Context is the core of the Agent.

#### func  NewContext

```go
func NewContext(cfg *config.Config) (*Context, error)
```
NewContext creates a new context. In general, there should only be one.

#### func (*Context) CreateJobLog

```go
func (context *Context) CreateJobLog() error
```
CreateJobLog creates a new job log

#### func (*Context) DeleteGuest

```go
func (ctx *Context) DeleteGuest(g *client.Guest) error
```
DeleteGuest removes a guest from the data store

#### func (*Context) DeleteGuestRunner

```go
func (context *Context) DeleteGuestRunner(guestID string)
```
DeleteGuestRunner deletes a GuestRunner

#### func (*Context) GetAction

```go
func (ctx *Context) GetAction(name string) (*Action, error)
```
GetAction looks up an action by name

#### func (*Context) GetAgentRunner

```go
func (context *Context) GetAgentRunner() (*GuestRunner, error)
```
GetAgentRunner retrieves the main agent runner

#### func (*Context) GetGuest

```go
func (ctx *Context) GetGuest(id string) (*client.Guest, error)
```
GetGuest fetches a single guest

#### func (*Context) GetGuestRunner

```go
func (context *Context) GetGuestRunner(guestID string) (*GuestRunner, error)
```
GetGuestRunner retrieves a GuestRunner

#### func (*Context) GetJobLog

```go
func (context *Context) GetJobLog() (*JobLog, error)
```
GetJobLog retrieves a job log

#### func (*Context) NewGuestRunner

```go
func (context *Context) NewGuestRunner(guestID string, maxInfo uint, maxStream uint) *GuestRunner
```
NewGuestRunner creates a new GuestRunner

#### func (*Context) NewService

```go
func (ctx *Context) NewService(name string, port uint, path string, maxConcurrent uint) (*Service, error)
```
NewService creates a new Service

#### func (*Context) PersistGuest

```go
func (ctx *Context) PersistGuest(g *client.Guest) error
```
PersistGuest writes guest data to the data store

#### func (*Context) RunGuests

```go
func (ctx *Context) RunGuests() error
```
RunGuests creates and runs helpers for each defined guest. In general, this
should only be called early in a process There is no locking provided.

#### type Guest

```go
type Guest struct {
	*client.Guest
}
```

Guest is a "helper struct"

#### type GuestRunner

```go
type GuestRunner struct {
	Context  *Context
	GuestID  string
	Info     *SyncThrottle
	Stream   *SyncThrottle
	Async    *PipelineQueue
	QuitChan chan struct{}
}
```

GuestRunner manages actions being performed for a guest

#### func (*GuestRunner) Process

```go
func (gr *GuestRunner) Process(pipeline *Pipeline) error
```
Process directs actions into sync or async handling depending on the type

#### func (*GuestRunner) Quit

```go
func (gr *GuestRunner) Quit()
```
Quit shuts down a GuestRunner

#### type HTTPErrorMessage

```go
type HTTPErrorMessage struct {
	Message string   `json:"message"`
	Code    int      `json:"code"`
	Stack   []string `json:"stack"`
}
```

HTTPErrorMessage is an enhanced error struct for http error responses

#### type HTTPRequest

```go
type HTTPRequest struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Context        *Context

	Guest       *client.Guest
	GuestRunner *GuestRunner
}
```

HTTPRequest is a container for an http request and its context

#### func (*HTTPRequest) JSON

```go
func (r *HTTPRequest) JSON(code int, obj interface{}) *HTTPErrorMessage
```
JSON sends an http response with a json body

#### func (*HTTPRequest) NewError

```go
func (r *HTTPRequest) NewError(err error, code int) *HTTPErrorMessage
```
NewError creates an HTTPErrorMessage

#### func (*HTTPRequest) Parameter

```go
func (r *HTTPRequest) Parameter(key string) string
```
Parameter retrieves request parameter

#### func (*HTTPRequest) SetHeader

```go
func (r *HTTPRequest) SetHeader(key, val string)
```
SetHeader sets an http response header

#### type Job

```go
type Job struct {
	ID        string
	GuestID   string
	Action    string
	QueuedAt  time.Time
	StartedAt time.Time
	UpdatedAt time.Time
	Status    JobStatus
	Message   string
}
```

Job holds information about a job

#### type JobLog

```go
type JobLog struct {
	GuestID     string
	Context     *Context
	Index       map[string]int
	GuestIndex  map[string][]int
	Jobs        []*Job
	ModifyMutex sync.Mutex
}
```

JobLog holds the most recent jobs for a guest

#### func (*JobLog) AddJob

```go
func (jobLog *JobLog) AddJob(jobID, guestID, action string) error
```
AddJob adds a job to the log

#### func (*JobLog) GetJob

```go
func (jobLog *JobLog) GetJob(jobID string) (*Job, error)
```
GetJob retrieves a job from the log based on job id

#### func (*JobLog) GetLatestGuestJobs

```go
func (jobLog *JobLog) GetLatestGuestJobs(guestID string, limit int) []*Job
```
GetLatestGuestJobs returns the latest X jobs in the log for a guest

#### func (*JobLog) GetLatestJobs

```go
func (jobLog *JobLog) GetLatestJobs(limit int) []*Job
```
GetLatestJobs returns the latest X jobs in the log

#### func (*JobLog) Persist

```go
func (jobLog *JobLog) Persist() error
```
Persist saves a job log

#### func (*JobLog) UpdateJob

```go
func (jobLog *JobLog) UpdateJob(jobID string, action string, status JobStatus, message string) error
```
UpdateJob updates a job's status and timing information

#### type JobStatus

```go
type JobStatus string
```

JobStatus is the string status of a job

#### type Pipeline

```go
type Pipeline struct {
	ID       string
	Action   string
	Type     config.ActionType
	Stages   []*Stage
	DoneChan chan error // Signal async is done or errored, for post-hooks
}
```

Pipeline is a full set of stage instances required to complete an action

#### func (*Pipeline) Run

```go
func (pipeline *Pipeline) Run() error
```
Run executes each stage in the pipeline. It bails out as soon as an error is
encountered

#### type PipelineQueue

```go
type PipelineQueue struct {
	GuestID      string
	Name         string
	Context      *Context
	PipelineChan chan *Pipeline
	QuitChan     chan struct{}
}
```

PipelineQueue holds asyncronous action pipelines

#### func  NewPipelineQueue

```go
func NewPipelineQueue(name string, guestID string, context *Context) *PipelineQueue
```
NewPipelineQueue creates a new PipelineQueue

#### func (*PipelineQueue) Enqueue

```go
func (pq *PipelineQueue) Enqueue(pipeline *Pipeline)
```
Enqueue queues an async action

#### func (*PipelineQueue) Process

```go
func (pq *PipelineQueue) Process()
```
Process monitors the queue and kicks off async actions

#### func (*PipelineQueue) Quit

```go
func (pq *PipelineQueue) Quit()
```
Quit signals the pipeline queue to stop processing after the current action

#### type Service

```go
type Service struct {
	Client *rpc.Client
	Name   string
}
```

Service is an RPC service

#### type Stage

```go
type Stage struct {
	Service  *Service
	Type     config.ActionType
	Method   string
	Args     map[string]string
	Request  interface{}
	Response interface{}
	RW       http.ResponseWriter // For streaming responses
}
```

Stage is a single step an action must take

#### func (*Stage) Run

```go
func (stage *Stage) Run() error
```
Run makes an individual stage request

#### type SyncThrottle

```go
type SyncThrottle struct {
	GuestID        string
	Name           string
	ConcurrentChan chan struct{}
	QuitChan       chan struct{}
}
```

SyncThrottle throttles synchronous actions

#### func  NewSyncThrottle

```go
func NewSyncThrottle(name string, guestID string, maxConcurrency uint) *SyncThrottle
```
NewSyncThrottle creates a new SyncThrottle

#### func (*SyncThrottle) Process

```go
func (st *SyncThrottle) Process(pipeline *Pipeline) error
```
Process runs an action

#### func (*SyncThrottle) Release

```go
func (st *SyncThrottle) Release()
```
Release signals that the action is done

#### func (*SyncThrottle) Reserve

```go
func (st *SyncThrottle) Reserve()
```
Reserve blocks until an action is allowed to run based on throttling

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
