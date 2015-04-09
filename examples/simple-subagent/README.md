# Creating a Subagent

A subagent provides specialized functionality to be performed during certain agent actions. For example, the storage subagent may be used as a stage of the guest `create` action to create the disks. It also may act as a simple hook, triggering side effects during actions.

This simple tutorial shows how to create a subagent with a single RPC method. It will randomly return either an error or the guest from the request. The final result is available in [main.go](./main.go).

## Imports

A few packages will be needed to make the subagent

```go
package main

import (
    "fmt"
    "math/rand"
    "net/http"
    "time"

    log "github.com/Sirupsen/logrus"
    "github.com/mistifyio/mistify-agent/rpc"
    logx "github.com/mistifyio/mistify-logrus-ext"
    flag "github.com/spf13/pflag"
)
```

* `fmt` will be used for creating formatted `Error` objects
* `math/rand` provides the random number generation behind whether to return an error or success response
* `net/http` is needed for the `http.Request` type
* `Sirupsen/logrus` allows for logs to be consistently formatted and to include a map of additional, easilly parsable fields
* `mistify-agent/rpc` provides the RPC types and sub-agent helper methods
* `mistifyio/mistify-logrus-ext` makes `logrus` easier to set up and handle error logging more cleanly
* `spf13/pflag` is a drop in replacement for the stdlib `flag` package and provides POSIX/GNU-style --flags

## Main Struct

This is a struct for the service and an instance will act as the receiver for all of the RPC calls.

```go
type Simple struct {
    rand    *rand.Rand // random number generator
    percent int        // how often to return an error
}
```

An instance holds a random number generator and the percentage of responses that should fail.

## RPC Method

This subagent will have a single RPC method attached to `Simple`. The method signature is defined by [gorilla/rpc](http://www.gorillatoolkit.org/pkg/rpc). It takes a `http.Request` pointer as well as the data request and response pointers, returning any error.

```go
func (s *Simple) DoStuff(r *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
    if num := s.rand.Intn(100); num <= s.percent {
        return fmt.Errorf("returning an error as I do %d%% of the time", s.percent)
    }
    // just return the guest from the response
    *response = rpc.GuestResponse{
        Guest: request.Guest,
    }
    return nil
}
```

It first decides whether to fail or not. In the case of a failure, it creates and returns an error. The success case points the response pointer at a new `rpc.GuestResponse` and returns `nil` for the error value. Turning either of these results into the final response to the requesting client is handled automatically.

## Main

The `main()` method handles setup and running of the subagent server. It contains the following parts:

### Configuration Flags

```go
var port uint
var percent uint

flag.UintVarP(&port, "port", "p", 21356, "listen port")
flag.UintVarP(&percent, "percent", "c", 50, "Percentage to return an error")
flag.Parse()

if percent > 100 {
    percent = 100
}
```

Here, `pflag` parses the arguments for what port the server should run on and how often it should return an error (limiting the maximum to 100%). A nice feature of `pflag` is that the `-h / --help` flag and output are provided automatically.

### Logging

```go
if err := logx.DefaultSetup("info"); err != nil {
    log.WithFields(log.Fields{
        "error": err,
        "func":  "logx.DefaultSetup",
    }).Fatal("failed to set up logging")
}
```

The `logx.DefaultSetup` method takes care of setting the log level and the formatter to JSON. The `Sirupsen/logrus` package can then be included anywhere and its methods, such as `log.Info("foo")`, will use the configured behavior.

### Receiver

```go
s := Simple{
    rand:    rand.New(rand.NewSource(time.Now().Unix())),
    percent: int(percent),
}
```

A new `Simple` is instantiated with the configured error probability and a newly seeded random number generator.

### RPC Server

```go
server, err := rpc.NewServer(port)
if err != nil {
    log.WithFields(log.Fields{
        "error": err,
        "func":  "rpc.NewServer",
    }).Fatal(err)
}

if err := server.RegisterService(&s); err != nil {
    log.WithFields(log.Fields{
        "error": err,
        "func":  "rpc.Server.RegisterService",
    }).Fatal(err)
}

log.WithField("port", port).Info("starting server")

if err = server.ListenAndServe(); err != nil {
    log.WithFields(log.Fields{
        "error": err,
        "func":  "rpc.Server.ListenAndServe",
    }).Fatal(err)
}
```

A new RPC server is created and the receiver (with its methods) is registered so that requests can be routed properly. The server is then ready to start listening and responding to requests.

The path that it will respond to requests on is `/_mistify_RPC_`

## Testing

By running `go run main.go`, the subagent will be compiled and started on the default port 21356 with the default error rate of 50%. There should be the following output:

```go
$ go run main.go
{"level":"info","msg":"starting server","port":21356,"time":"2015-04-09T21:56:11Z"}
```

### Request Structure

```json
{
    "method": "Simple.DoStuff",
    "params": [
        {
            "guest": {
                "id": "123456789"
            }
        }
    ],
    "id": 0
}
```

* `method` is which receiver method to run.
* `params` is an array containing one `rpc.GuestRequest` (many fields omitted), as that is what the `Simple.DoStuff` method is set up to receive.
* `id` is the id of the mistify agent making the request, but 0 can be used for testing.

### Making Requests

Using curl, requests can be issued to the subagent as follows:

```bash
curl -s -H "Content-Type: application/json" \
    http://localhost:21356/_mistify_RPC_ \
    --data-binary '{ "method": "Simple.DoStuff", 
    "params": [ { "guest": { "id": "123456789" } } ], "id": 0 }'
```

50% of the responses should be an error:

```json
{
    "result": null,
    "error": "returning an error as I do 50% of the time",
    "id": 0
}
```

50% of the responses should be successful, returning the guest back:

```json
{
    "result": {
        "guest": {
            "id": "123456789",
            "action": ""
        }
    },
    "error": null,
    "id": 0
}
```

## Agent Integration

To use the subagent with the main agent, the subagent method needs to be added as a stage to one of the agent action pipelines in the agent config file. Using the [test agent](../test-rpc-service) with its `agent.json`:

Add the service like so:

```json
"services": {
    "test": {
        "port": 9999
    },
    "testDownload": {
        "path": "/snapshots/download",
        "port": 9999
    },
    "simple": {
        "port": 21356
    }
}
```

And then add the subagent method to an action, like `create`

```json
"create": {
    "stages": [
        {
            "method": "Simple.DoStuff",
            "service": "simple"
        },
        {
            "method": "Test.Create",
            "service": "test"
        }
    ]
},
```
_Note: All stages in an action share a request, so create subagent method signatures accordingly. See the official subagent methods for each action for more information._

Now restart the agent. With both the agent and the subagent running, whenever a guest create action is requested, it will make a call to the simple subagent (and fail 50% of the time).
