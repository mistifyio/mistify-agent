How to write a simple sub-agent
===============================

This is a simple "tutorial" for writing a simple sub-agent.  The completed code can be seen in [main.go](./main.go)


This simple sub-agent will present a single RPC method that returns an error a specified percentage of the time.  The rest of the time, it just returns the guest from the request unmodified in the response.

# Imports #

First, let's do our package and  imports:

```go
/*
This is a simple example of a sub-agent for Mistify Agent.  It does not do anything useful, but shows the API.
*/
package main

import (
	"fmt"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/mistifyio/mistify-agent/rpc"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)
```

We will be using Go's builtin [math/rand](http://golang.org/pkg/math/rand) package for random number generation.

I prefer the [docker "hack"](http://godoc.org/github.com/docker/docker/pkg/mflag) version of the stdlib [flag package](http://golang.org/pkg/flag/) as it supports the more familiar "double-dash" (`--`) arguments.

The [Mistify RPC package](http://github.com/mistifyio/mistify-agent/rpc) provides the types for RPC communication as well as some helpers for quickly developing an RPC sub-agent.

# Struct #

We have a simple struct for our service.  An instance of this struct is the receiver for all the RPC calls.

```go
type (
	Simple struct {
		rand    *rand.Rand // random number generator
		percent int            // how often to return an error
	}
)
```

# RPC methods #

We have a single RPC method. The signature is defined by http://www.gorillatoolkit.org/pkg/rpc

```go
// DoStuff does not actually do anything. It returns an error a certain percentage of the time.
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

# main #

In `main`, we must parse our arguments and setup our RPC server.

## flags ##
Here we simply use `mflag` to parse the arguments:

```go
    var port int
	var percent uint
	var h bool

	flag.BoolVar(&h, []string{"h", "#help", "-help"}, false, "display the help")
	flag.IntVar(&port, []string{"p", "#port", "-port"}, 21356, "listen port")
	flag.UintVar(&percent, []string{"c", "#percent", "-percent"}, 50, "Percentage to return an error")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if percent > 100 {
		percent = 100
	}
```

We also make sure that we have a maximum of 100%.

## Receiver##

Now we make a new `Simple` based on the user input.  We are simply seeding rand with the current time - you would probably do something more secure if needed. This will be the receiver for the RPC methods.

```go
	s := Simple{
		rand:    rand.New(rand.NewSource(time.Now().Unix())),
		percent: int(percent),
	    }
```

## RPC server ##

Now, we create the RPC server and register.


```go
   server, err := rpc.NewServer(port)
	if err != nil {
		log.Fatal(err)
	}

	server.RegisterService(&s)
	log.Fatal(server.ListenAndServe())
```

This uses the "helpers" provided by the mistify rpc package to create a server listening an the requested port and to register our receiver.  `ListenAndServe` generally does not return.

## Test ###

Now you should be able to test. For now just run `go run main.go` the process should start and listen on port `21356`.

You should be able to test this by using curl:

```
 curl -s -H "Content-Type: application/json" http://localhost:21356/_mistify_RPC_ --data-binary '{ "method": "Simple.DoStuff", "params": [ { "guest": { "id": "123456789" } } ], "id": 0 }'
 ```

Notice we are calling the `Simple.DoStuff` method which is the name of the struct with the method name.  Also, we are passing a bare minimum representation of a guest.

50% of the time you should get an error:

```json
{"result":null,"error":"returning an error as I do 50% of the time","id":0}
```

and 50% of the time get the same guest back:

```json
{"result":{"guest":{"id":"123456789","action":""}},"error":null,"id":0}
```

## Using with the Agent ##

To actually use in the agent, we need to add it to the agent config. Assuming we are using the [test/stub config](../test-rpc-service):

Add the service to the config:

```json
 "services": {
        "test": {
            "port": 9999
            },
         "simple": {
            "port": 21356
        }
    }
```
we could change the create action to:

```json
    "create": {
         "sync": [
             {
                "method": "Simple.DoStuff",
                "service": simple"
              }
           ],
            "async": [
                {
                    "method": "Test.Create",
                    "service": "test"
                }
            ]
            },
```

And restart the agent. Now, 50% of the time, creates will fail during request time.

