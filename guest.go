package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"unicode"
	"unicode/utf8"

	"code.google.com/p/go-uuid/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
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

const requestGuestKey = "requestGuest"

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

func listGuests(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := GetContext(r)
	var guests []*client.Guest

	err := ctx.db.Transaction(func(tx *kvite.Tx) error {
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
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	hr.JSON(http.StatusOK, guests)
}

// TODO: A lot of the duplicated code between here and the guest action wrapper
// will go away when we fix our middlewares. The initial setup here can be a
// simple middleware called first before the guest and runner retrieval
// middlewares
// NOTE: The config for create should include stages for startup
func createGuest(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := GetContext(r)

	g := &client.Guest{}
	if err := json.NewDecoder(r.Body).Decode(g); err != nil {
		hr.JSONError(http.StatusBadRequest, err)
		return
	}
	if g.Id != "" {
		if uuid.Parse(g.Id) == nil {
			hr.JSONError(http.StatusBadRequest, fmt.Errorf("id must be uuid"))
			return
		}
	} else {
		g.Id = uuid.New()
	}

	// TODO: make sure it's actually unique
	g.State = "create"

	// TODO: general validations, like memory, disks look sane, etc

	action, err := ctx.GetAction(prefixedActionName(g.Type, "create"))
	if err != nil {
		hr.JSONError(http.StatusNotFound, err)
		return
	}

	if err := ctx.PersistGuest(g); err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}

	runner := ctx.NewGuestRunner(g.Id, 100, 5)

	response := &rpc.GuestResponse{}
	request := &rpc.GuestRequest{
		Guest:  g,
		Action: action.Name,
	}
	pipeline := action.GeneratePipeline(request, response, hr, nil)
	// PreStageFunc copies the stage args into the request
	pipeline.PreStageFunc = func(p *Pipeline, s *Stage) error {
		request.Args = s.Args
		return nil
	}
	// PostStageFunc saves the guest and uses it for the next request
	pipeline.PostStageFunc = func(p *Pipeline, s *Stage) error {
		request.Guest = response.Guest
		return ctx.PersistGuest(response.Guest)
	}

	hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
	err = runner.Process(pipeline)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusAccepted, g)
}

// GetGuestMiddleware retrieves guest information into the request context
func GetGuestMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hr := &HTTPResponse{w}
		ctx := GetContext(r)
		vars := mux.Vars(r)

		id := vars["id"]
		var g client.Guest
		err := ctx.db.Transaction(func(tx *kvite.Tx) error {
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
			code := http.StatusInternalServerError
			if err == ErrNotFound {
				code = http.StatusNotFound
			}
			hr.JSONError(code, err)
			return
		}

		context.Set(r, requestGuestKey, &g)
		h.ServeHTTP(w, r)
	})
}

func getGuest(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	hr.JSON(http.StatusOK, GetRequestGuest(r))
}

func deleteGuest(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := GetContext(r)
	g := GetRequestGuest(r)

	err := ctx.PersistGuest(g)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusAccepted, g)
}

func getGuestMetadata(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	g := GetRequestGuest(r)
	hr.JSON(http.StatusOK, g.Metadata)
}

func setGuestMetadata(w http.ResponseWriter, r *http.Request) {
	hr := &HTTPResponse{w}
	ctx := GetContext(r)
	g := GetRequestGuest(r)

	var metadata map[string]string
	err := json.NewDecoder(r.Body).Decode(&metadata)
	if err != nil {
		hr.JSONError(http.StatusBadRequest, err)
		return
	}

	for key, value := range metadata {
		if value == "" {
			delete(g.Metadata, key)
		} else {
			g.Metadata[key] = value
		}
	}

	err = ctx.PersistGuest(g)
	if err != nil {
		hr.JSONError(http.StatusInternalServerError, err)
		return
	}
	hr.JSON(http.StatusOK, g.Metadata)
}

// guestRunnerMiddleware gets and places the runner into the request context
func guestRunnerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := GetContext(r)
		guest := GetRequestGuest(r)
		runner, err := ctx.GetGuestRunner(guest.Id)
		if err != nil {
			hr := &HTTPResponse{w}
			hr.JSONError(http.StatusInternalServerError, err)
			return
		}

		context.Set(r, requestRunnerKey, runner)
		h.ServeHTTP(w, r)
	})
}

// GenerateGuestAction creates a handler function for a particular guest action
func GenerateGuestAction(actionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hr := &HTTPResponse{w}
		ctx := GetContext(r)
		g := GetRequestGuest(r)
		runner := GetRequestRunner(r)

		actionName := prefixedActionName(g.Type, actionName)
		action, err := ctx.GetAction(actionName)
		if err != nil {
			hr.JSONError(http.StatusNotFound, err)
			return
		}

		response := &rpc.GuestResponse{}
		request := &rpc.GuestRequest{
			Guest:  g,
			Action: action.Name,
		}
		doneChan := make(chan error)
		pipeline := action.GeneratePipeline(request, response, hr, doneChan)
		// PreStageFunc copies the stage args into the request
		pipeline.PreStageFunc = func(p *Pipeline, s *Stage) error {
			request.Args = s.Args
			return nil
		}
		// PostStageFunc saves the guest and uses it for the next request
		pipeline.PostStageFunc = func(p *Pipeline, s *Stage) error {
			request.Guest = response.Guest
			return ctx.PersistGuest(response.Guest)
		}

		hr.Header().Set("X-Guest-Job-ID", pipeline.ID)
		err = runner.Process(pipeline)
		if err != nil {
			hr.JSONError(http.StatusInternalServerError, err)
			return
		}
		// Extra processing after the pipeline finishes
		go func() {
			if <-doneChan != nil {
				return
			}
			if actionName == prefixedActionName(g.Type, "delete") {
				if err := ctx.DeleteGuest(g); err != nil {
					log.WithFields(log.Fields{
						"guest": g.Id,
						"error": err,
						"func":  "agent.Context.DeleteGuest",
					}).Error("Delete Error:", err)
				}
				return
			}
		}()
		hr.JSON(http.StatusAccepted, g)
	}
}

// GetRequestGuest retrieves the guest from the request context
func GetRequestGuest(r *http.Request) *client.Guest {
	if value := context.Get(r, requestGuestKey); value != nil {
		return value.(*client.Guest)
	}
	return nil
}
