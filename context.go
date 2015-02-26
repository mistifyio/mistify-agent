package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/config"
)

type (
	// Context is the core of the Agent.
	Context struct {
		db       *kvite.DB
		Config   *config.Config
		Actions  map[string]*Action
		Services map[string]*Service

		GuestRunners     map[string]*GuestRunner
		GuestRunnerMutex sync.Mutex
		GuestJobLogs     map[string]*GuestJobLog
	}
)

// NewContext creates a new context. In general, there should only be one.
func NewContext(cfg *config.Config) (*Context, error) {
	ctx := &Context{
		Config:   cfg,
		Actions:  make(map[string]*Action),
		Services: make(map[string]*Service),
	}

	db, err := kvite.Open(cfg.DBPath, "mistify_agent")
	if err != nil {
		return nil, err
	}
	ctx.db = db
	if err != nil {
		return nil, err
	}

	for name, service := range cfg.Services {
		ctx.Services[name], err = ctx.NewService(name, service.Port, service.Path, service.MaxPending)
		if err != nil {
			return nil, err
		}
	}

	for name, cfgAction := range cfg.Actions {

		action := &Action{
			Name:   name,
			Type:   cfgAction.Type,
			Stages: make([]*Stage, len(cfgAction.Stages)),
		}

		for i, stage := range cfgAction.Stages {
			action.Stages[i] = &Stage{
				Service: ctx.Services[stage.Service],
				Type:    action.Type,
				Method:  stage.Method,
				Args:    stage.Args,
			}
		}

		ctx.Actions[name] = action
	}

	ctx.GuestRunners = make(map[string]*GuestRunner)
	ctx.GuestJobLogs = make(map[string]*GuestJobLog)

	log.WithFields(log.Fields{
		"data": ctx,
	}).Info("created new context")

	return ctx, nil
}

func (ctx *Context) GetAction(name string) (*Action, error) {
	action, ok := ctx.Actions[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("%s: Not configured", name))
	}
	return action, nil
}

// GetGuest fetches a single guest
func (ctx *Context) GetGuest(id string) (*client.Guest, error) {
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
			return NotFound
		}

		return json.Unmarshal(data, &g)
	})
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// RunGuests creates and runs helpers for each defined guest. In general, this should only be called early in a process
// There is no locking provided.
func (ctx *Context) RunGuests() error {
	return ctx.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guests")
		if err != nil {
			return err
		}
		return b.ForEach(func(k string, v []byte) error {
			var guest client.Guest
			if err := json.Unmarshal(v, &guest); err != nil {
				// should this be fatal if it just fails on one guest??
				return err
			}
			_ = ctx.NewGuestRunner(guest.Id, 100, 5)
			_ = ctx.CreateGuestJobLog(guest.Id)
			return nil
		})
	})
}
