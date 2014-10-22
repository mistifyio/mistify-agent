package agent

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"fmt"
	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
	"net/http"
)

type (
	Guest struct {
		context *Context
		*client.Guest
	}
)

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

func listGuests(r *HttpRequest) *HttpErrorMessage {
	guests := make([]*client.Guest, 0)

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
		return nil
	})

	if err != nil {
		return r.NewError(err, 500)
	}

	return r.JSON(200, guests)
}

func (ctx *Context) runSyncAction(guest *client.Guest) (*client.Guest, error) {
	action, ok := ctx.Actions[guest.Action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", guest.Action)
	}
	return action.Sync.Run(guest)
}

// XXX: all the sync/async initialization and running seems redundant
func (ctx *Context) runAsyncAction(guest *client.Guest) (*client.Guest, error) {
	action, ok := ctx.Actions[guest.Action]
	if !ok {
		return nil, fmt.Errorf("unknown action: %s", guest.Action)
	}
	return action.Async.Run(guest)
}

func createGuest(r *HttpRequest) *HttpErrorMessage {
	var g client.Guest
	err := json.NewDecoder(r.Request.Body).Decode(&g)
	if err != nil {
		return r.NewError(err, 400)
	}
	if g.Id != "" {
		return r.NewError(fmt.Errorf("id must not be set"), 400)
	}
	g.Id = uuid.New()

	// TODO: make sure it's actually unique
	g.State = "create"
	g.Action = "create"

	// TODO: general validations, like memory, disks look sane, etc

	guest, err := r.Context.runSyncAction(&g)
	if err != nil {
		return r.NewError(err, 500)
	}

	if err := r.Context.PersistGuest(guest); err != nil {
		return r.NewError(err, 500)
	}

	_, err = r.Context.CreateGuestRunner(guest)
	if err != nil {
		return r.NewError(err, 500)
	}
	return r.JSON(202, guest)
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

func getGuest(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		return r.JSON(200, g)
	})
}

func deleteGuest(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		g.Action = "delete"
		err := r.Context.PersistGuest(g)
		if err != nil {
			return r.NewError(err, 500)
		}
		return r.JSON(202, g)
	})
}

func getGuestMetadata(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		return r.JSON(200, g.Metadata)
	})
}

func setGuestMetadata(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
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

func (ctx *Context) switchToRunning(guest *client.Guest) error {
	guest.Action = "run"
	return ctx.PersistGuest(guest)
}

func (c *Chain) GuestActionWrapper(action string) http.HandlerFunc {
	return c.RequestWrapper(func(r *HttpRequest) *HttpErrorMessage {
		return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
			g.Action = action
			guest, err := r.Context.runSyncAction(g)
			if err != nil {
				return r.NewError(err, 500)
			}
			return r.JSON(202, guest)
		})
	})
}

func getGuestCpuMetrics(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		metrics, err := r.Context.getCpuMetrics(g)
		if err != nil {
			return r.NewError(err, 500)
		}
		return r.JSON(200, metrics)
	})
}

func getGuestNicMetrics(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		metrics, err := r.Context.getNicMetrics(g)
		if err != nil {
			return r.NewError(err, 500)
		}
		return r.JSON(200, metrics)
	})
}

func getGuestDiskMetrics(r *HttpRequest) *HttpErrorMessage {
	return withGuest(r, func(g *client.Guest) *HttpErrorMessage {
		metrics, err := r.Context.getDiskMetrics(g)
		if err != nil {
			return r.NewError(err, 500)
		}
		return r.JSON(200, metrics)
	})
}
