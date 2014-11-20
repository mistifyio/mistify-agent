package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.google.com/p/go-uuid/uuid"
	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
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
	ctx.RunnerMutex.Lock()
	defer ctx.RunnerMutex.Unlock()
	delete(ctx.Runners, g.Id)
	return nil
}

func listGuests(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext(r)
	guests := make([]*client.Guest, 0)

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
		NewError(err, 500).Serve(w)
		return
	}

	JSON(w, 200, guests)
}

func (ctx *Context) runSyncAction(guest *client.Guest) (*client.Guest, error) {
	action, ok := ctx.Actions[guest.Action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", guest.Action)
	}

	g, err := action.Sync.Run(guest)
	if err != nil {
		return nil, err
	}
	ctx.NudgeGuest(g.Id)
	return g, nil
}

// XXX: all the sync/async initialization and running seems redundant
func (ctx *Context) runAsyncAction(guest *client.Guest) (*client.Guest, error) {
	action, ok := ctx.Actions[guest.Action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", guest.Action)
	}
	return action.Async.Run(guest)

}

func createGuest(w http.ResponseWriter, r *http.Request) {
	ctx := GetContext(r)
	var g client.Guest
	err := json.NewDecoder(r.Body).Decode(&g)
	if err != nil {
		NewError(err, 400).Serve(w)
		return
	}
	if g.Id != "" {
		NewError(errors.New("id must not be set"), 400).Serve(w)
		return
	}
	g.Id = uuid.New()

	// TODO: make sure it's actually unique
	g.State = "create"
	g.Action = "create"

	guest, err := ctx.runSyncAction(&g)
	if err != nil {
		NewError(err, 500).Serve(w)
		return
	}

	if err := ctx.PersistGuest(guest); err != nil {
		NewError(err, 500).Serve(w)
		return
	}

	runner, err := ctx.CreateGuestRunner(guest)
	if err != nil {
		NewError(err, 500).Serve(w)
		return
	}

	runner.Nudge()

	JSON(w, 202, g)
}

func withGuest(r *HttpRequest, fn func(g *client.Guest) *HttpErrorMessage) *HttpErrorMessage {
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
			return NotFound
		}

		return json.Unmarshal(data, &g)
	})

	if err != nil {
		code := 500
		if err == NotFound {
			code = 404
		}
		return r.NewError(err, code)
	}
	return fn(&g)
}

func getGuest(w http.ResponseWriter, r *http.Request) {
	// we always assume that this is called after GuestHandler, it handles the error case
	if g := GetGuest(r); g != nil {
		JSON(w, 200, g)
	}
}

func deleteGuest(w http.ResponseWriter, r *http.Request) {
	// will panic if not called after GuestHandler
	g := GetGuest(r)
	ctx := GetContext(r)
	g.Action = "delete"
	err := ctx.PersistGuest(g)
	if err != nil {
		NewError(err, 500).Serve(w)
	}
	JSON(w, 202, g)
}

func getGuestMetadata(w http.ResponseWriter, r *http.Request) {
	if g := GetGuest(r); g != nil {
		JSON(w, 200, g.Metadata)
	}
}

func setGuestMetadata(w http.ResponseWriter, r *http.Request) {
	// will panic if not called after GuestHandler
	g := GetGuest(r)
	var metadata map[string]string
	err := json.NewDecoder(r.Body).Decode(&metadata)
	if err != nil {
		NewError(err, 400).Serve(w)
		return
	}

	for key, value := range metadata {
		if value == "" {
			delete(g.Metadata, key)
		} else {
			g.Metadata[key] = value
		}
	}

	//assume we get called after ContextHandler. If not, this will panic
	ctx := GetContext(r)

	err = ctx.PersistGuest(g)
	if err != nil {
		NewError(err, 400).Serve(w)
		return
	}
	JSON(w, 200, g.Metadata)

}

func (ctx *Context) switchToRunning(guest *client.Guest) error {
	guest.Action = "run"
	return ctx.PersistGuest(guest)
}

func actionWrapper(action string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g := GetGuest(r)
		ctx := GetContext(r)
		g.Action = action
		guest, err := ctx.runSyncAction(g)
		if err != nil {
			NewError(err, 500).Serve(w)
			return
		}
		JSON(w, 200, guest)
	})
}

func getGuestCPUMetrics(w http.ResponseWriter, r *http.Request) {
	g := GetGuest(r)
	ctx := GetContext(r)
	metrics, err := ctx.getCpuMetrics(g)
	if err != nil {
		NewError(err, 500).Serve(w)
		return
	}
	JSON(w, 200, metrics)

}

func getGuestNicMetrics(w http.ResponseWriter, r *http.Request) {
	g := GetGuest(r)
	ctx := GetContext(r)
	metrics, err := ctx.getNicMetrics(g)
	if err != nil {
		NewError(err, 500).Serve(w)
		return
	}
	JSON(w, 200, metrics)

}

func getGuestDiskMetrics(w http.ResponseWriter, r *http.Request) {
	g := GetGuest(r)
	ctx := GetContext(r)
	metrics, err := ctx.getDiskMetrics(g)
	if err != nil {
		NewError(err, 500).Serve(w)
		return
	}
	JSON(w, 200, metrics)

}
