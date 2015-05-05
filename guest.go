package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"unicode"
	"unicode/utf8"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (

	// Guest is a "helper struct"
	Guest struct {
		context *Context
		*client.Guest
	}
)

// PersistGuest writes guest data to the data store
func (ctx *Context) PersistGuest(g *client.Guest) error {
	return ctx.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guests")
		if err != nil {
			return err
		}
		data, err := json.Marshal(g)
		if err != nil {
			return err
		}
		return b.Put(g.Id, data)
	})
}

// DeleteGuest removes a guest from the data store
func (ctx *Context) DeleteGuest(g *client.Guest) error {
	err := ctx.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guests")
		if err != nil {
			return err
		}
		return b.Delete(g.Id)
	})

	if err != nil {
		return err
	}
	ctx.DeleteGuestRunner(g.Id)
	return nil
}

// prefixedActionName creates the appropriate action name for the guest type
func prefixedActionName(gType, actionName string) string {
	if gType != "container" || actionName == "" {
		return actionName
	}
	r, n := utf8.DecodeRuneInString(actionName)
	return "container" + string(unicode.ToUpper(r)) + actionName[n:]
}

func listGuests(r *HTTPRequest) *HTTPErrorMessage {
	var guests []*client.Guest

	err := r.Context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guests")
		if err != nil {
			return err
		}
		return b.ForEach(func(k string, v []byte) error {
			var g client.Guest
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			// Do we want to actually verify this information or trust the pipelines??
			guests = append(guests, &g)
			return nil
		})
	})

	if err != nil {
		return r.NewError(err, 500)
	}

	return r.JSON(200, guests)
}

// TODO: A lot of the duplicated code between here and the guest action wrapper
// will go away when we fix our middlewares. The initial setup here can be a
// simple middleware called first before the guest and runner retrieval
// middlewares
// NOTE: The config for create should include stages for startup
func createGuest(r *HTTPRequest) *HTTPErrorMessage {
	g := &client.Guest{}
	if err := json.NewDecoder(r.Request.Body).Decode(g); err != nil {
		return r.NewError(err, 400)
	}
	if g.Id != "" {
		if uuid.Parse(g.Id) == nil {
			return r.NewError(fmt.Errorf("id must be uuid"), 400)
		}
	} else {
		g.Id = uuid.New()
	}

	// TODO: make sure it's actually unique
	g.State = "create"

	// TODO: general validations, like memory, disks look sane, etc

	if err := r.Context.PersistGuest(g); err != nil {
		return r.NewError(err, 500)
	}

	runner := r.Context.NewGuestRunner(g.Id, 100, 5)

	action, err := r.Context.GetAction(prefixedActionName(g.Type, "create"))
	if err != nil {
		return r.NewError(err, 404)
	}
	response := &rpc.GuestResponse{}
	pipeline := action.GeneratePipeline(nil, response, r.ResponseWriter, nil)
	// Guest requests are special in that they have Args and need
	// the action name, so fold them in to the request
	for _, stage := range pipeline.Stages {
		stage.Request = &rpc.GuestRequest{
			Guest:  g,
			Args:   stage.Args,
			Action: action.Name,
		}
	}
	r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(202, g)
}

func withGuest(r *HTTPRequest, fn func(r *HTTPRequest) *HTTPErrorMessage) *HTTPErrorMessage {
	id := r.Parameter("id")
	var g client.Guest
	err := r.Context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guests")
		if err != nil {
			return err
		}
		data, err := b.Get(id)
		if err != nil {
			return err
		}
		if data == nil {
			return ErrNotFound
		}

		return json.Unmarshal(data, &g)
	})

	if err != nil {
		code := 500
		if err == ErrNotFound {
			code = 404
		}
		return r.NewError(err, code)
	}
	r.Guest = &g
	r.GuestRunner, err = r.Context.GetGuestRunner(g.Id)
	if err != nil {
		return r.NewError(err, 500)
	}
	return fn(r)
}

func getGuest(r *HTTPRequest) *HTTPErrorMessage {
	return withGuest(r, func(r *HTTPRequest) *HTTPErrorMessage {
		g := r.Guest
		return r.JSON(200, g)
	})
}

func deleteGuest(r *HTTPRequest) *HTTPErrorMessage {
	return withGuest(r, func(r *HTTPRequest) *HTTPErrorMessage {
		g := r.Guest
		// TODO: Make sure to use the DoneChan here
		err := r.Context.PersistGuest(g)
		if err != nil {
			return r.NewError(err, 500)
		}
		return r.JSON(202, g)
	})
}

func getGuestMetadata(r *HTTPRequest) *HTTPErrorMessage {
	return withGuest(r, func(r *HTTPRequest) *HTTPErrorMessage {
		g := r.Guest
		return r.JSON(200, g.Metadata)
	})
}

func setGuestMetadata(r *HTTPRequest) *HTTPErrorMessage {
	return withGuest(r, func(r *HTTPRequest) *HTTPErrorMessage {
		g := r.Guest
		var metadata map[string]string
		err := json.NewDecoder(r.Request.Body).Decode(&metadata)
		if err != nil {
			return r.NewError(err, 400)
		}

		for key, value := range metadata {
			if value == "" {
				delete(g.Metadata, key)
			} else {
				g.Metadata[key] = value
			}
		}

		err = r.Context.PersistGuest(g)
		if err != nil {
			return r.NewError(err, 500)
		}
		return r.JSON(200, g.Metadata)
	})
}

// TODO: These wrappers are ugly nesting. Try to find a cleaner, more modular
// way to do it

// GuestRunnerWrapper is a middleware that retrieves the runner for a guest
func (c *Chain) GuestRunnerWrapper(fn func(*HTTPRequest) *HTTPErrorMessage) http.HandlerFunc {
	return c.RequestWrapper(func(r *HTTPRequest) *HTTPErrorMessage {
		return withGuest(r, func(r *HTTPRequest) *HTTPErrorMessage {
			g := r.Guest
			runner, err := r.Context.GetGuestRunner(g.Id)
			if err != nil {
				return r.NewError(err, 500)
			}

			r.GuestRunner = runner
			return fn(r)
		})
	})
}

// GuestActionWrapper wraps an HTTP request with a Guest action to avoid duplicated code
func (c *Chain) GuestActionWrapper(actionName string) http.HandlerFunc {
	return c.GuestRunnerWrapper(func(r *HTTPRequest) *HTTPErrorMessage {
		g := r.Guest
		runner := r.GuestRunner

		actionName := prefixedActionName(g.Type, actionName)
		action, err := r.Context.GetAction(actionName)
		if err != nil {
			return r.NewError(err, 404)
		}

		response := &rpc.GuestResponse{}
		doneChan := make(chan error)
		pipeline := action.GeneratePipeline(nil, response, r.ResponseWriter, doneChan)
		// Guest requests are special in that they have Args and need
		// the action name, so fold them in to the request
		for _, stage := range pipeline.Stages {
			stage.Request = &rpc.GuestRequest{
				Guest:  g,
				Args:   stage.Args,
				Action: action.Name,
			}
		}
		r.ResponseWriter.Header().Set("X-Guest-Job-ID", pipeline.ID)
		err = runner.Process(pipeline)
		if err != nil {
			return r.NewError(err, 500)
		}
		// Extra processing after the pipeline finishes
		go func() {
			err := <-doneChan
			if err != nil {
				return
			}
			if actionName == prefixedActionName(g.Type, "delete") {
				if err := r.Context.DeleteGuest(g); err != nil {
					log.WithFields(log.Fields{
						"guest": g.Id,
						"error": err,
						"func":  "agent.Context.DeleteGuest",
					}).Error("Delete Error:", err)
				}
				return
			}
			if err := r.Context.PersistGuest(g); err != nil {
				return
			}
		}()
		return r.JSON(202, g)
	})
}
