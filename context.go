package agent

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/mistifyio/kvite"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/config"
	"github.com/mistifyio/mistify-agent/log"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
	// Context is the core of the Agent.
	Context struct {
		db          *kvite.DB
		ImageClient *rpc.Client
		Config      *config.Config
		Actions     map[string]*Action
		Services    map[string]*Service

		Runners     map[string]*GuestRunner
		RunnerMutex sync.Mutex

		Metrics map[string]*Stage
		// TODO: have a metrics method/channel
	}

	// GuestRunner represents a task runner for a single guest
	GuestRunner struct {
		Context *Context
		ID      string
		nudge   chan struct{} //this "nudges" the runner to run if it's in select
		exit    chan struct{} //tells the runner to exit
		wait    time.Duration //time to wait between runs
	}
)

// NewContext creates a new context. In general, there should only be one.
func NewContext(cfg *config.Config) (*Context, error) {
	ctx := &Context{
		Config:   cfg,
		Actions:  make(map[string]*Action),
		Services: make(map[string]*Service),
		Metrics:  make(map[string]*Stage),
	}

	db, err := kvite.Open(cfg.DBPath, "mistify_agent")
	if err != nil {
		return nil, err
	}
	ctx.db = db
	ctx.ImageClient, err = rpc.NewClient(16000)
	if err != nil {
		return nil, err
	}

	for name, service := range cfg.Services {
		ctx.Services[name], err = ctx.NewService(name, service.Port, service.MaxPending)
		if err != nil {
			return nil, err
		}
	}

	for name, cfgAction := range cfg.Actions {

		action := &Action{
			Name: name,
			ctx:  ctx,
		}

		stages := make([]*Stage, 0, len(cfgAction.Sync))
		for _, stage := range cfgAction.Sync {
			//log.Info(stage.Service)
			//log.Info("%+v\n", ctx.Services[stage.Service])
			stages = append(stages, &Stage{
				Action:  name,
				Service: ctx.Services[stage.Service],
				Method:  stage.Method,
				Args:    stage.Args,
			})
		}

		action.Sync, err = ctx.NewPipeline(stages, name)
		if err != nil {
			return nil, err
		}

		stages = make([]*Stage, 0, len(cfgAction.Async))
		for _, stage := range cfgAction.Async {
			stages = append(stages, &Stage{
				Action:  name,
				Service: ctx.Services[stage.Service],
				Method:  stage.Method,
				Args:    stage.Args,
			})
		}
		action.Async, err = ctx.NewPipeline(stages, name)
		log.Info("async stages: %s %+v", name, stages)
		log.Info("async pipeline: %s %+v", name, action.Async)

		if err != nil {
			return nil, err
		}

		ctx.Actions[name] = action
	}

	for name, cfgMetric := range cfg.Metrics {
		stage := &Stage{
			Action:  name,
			Service: ctx.Services[cfgMetric.Service],
			Method:  cfgMetric.Method,
			Args:    cfgMetric.Args,
		}

		ctx.Metrics[name] = stage
	}

	ctx.Runners = make(map[string]*GuestRunner)

	data, err := json.MarshalIndent(ctx, "   ", " ")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("%s\n", data)

	return ctx, nil
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

// RunAction executes the pipeline associated with the current guest action
func (runner *GuestRunner) RunAction() {
	log.Error("GuestRunner: %s start", runner.ID)

	guest, err := runner.Context.GetGuest(runner.ID)
	if err != nil {
		log.Error("GuestRunner: %s => %s", runner.ID, err)
		return
	}

	action := guest.Action
	log.Info("%s: running %s", runner.ID, action)

	guest, err = runner.Context.runAsyncAction(guest)
	if err != nil {
		// this error should get persisted to the guest somehow?
		log.Error("GuestRunner: %s, %s", runner.ID, err)
		return
	}
	//create and delete are special
	// should we allow the stages to change action???
	switch action {
	case "create":
		guest.Action = "run"
		if err := runner.Context.PersistGuest(guest); err != nil {
			// this error should get persisted to the guest somehow?
			log.Error("GuestRunner: %s", runner.ID, err)
		}
		runner.Nudge()
	case "delete":
		if err := runner.Context.DeleteGuest(guest); err != nil {
			log.Error("GuestRunner: %s", runner.ID, err)
		}
		runner.Exit()
	}
}

// Run executes the guest loop
func (runner *GuestRunner) Run() {
LOOP:
	for {
		log.Info("run loop %s", runner.ID)

		// TODO: be configurable
		select {

		case <-runner.exit:
			log.Info("run loop time to die %s", runner.ID)
			break LOOP

			// TODO: have a way for stages/pipelines to set this and us set it back??
		case <-time.After(runner.wait):
			runner.RunAction()

		case <-runner.nudge:
			runner.RunAction()
		}
	}
}

// CreateGuestRunner creates a new runner for a guest
func (ctx *Context) CreateGuestRunner(guest *client.Guest) (*GuestRunner, error) {
	ctx.RunnerMutex.Lock()
	defer ctx.RunnerMutex.Unlock()

	runner := &GuestRunner{
		Context: ctx,
		ID:      guest.Id,
		// what should the buffering be here?
		nudge: make(chan struct{}, 4),
		exit:  make(chan struct{}, 1),
		// configurable?
		wait: 5 * time.Second,
	}

	ctx.Runners[guest.Id] = runner
	log.Info("starting %s, %s", guest.Id, guest.Action)
	go runner.Run()

	return runner, nil
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
			_, err := ctx.CreateGuestRunner(&guest)
			if err != nil {
				return err
			}
			return nil
		})
	})
}

func (runner *GuestRunner) Exit() {
	runner.exit <- struct{}{}
}

func (runner *GuestRunner) Nudge() {
	runner.nudge <- struct{}{}
}

func (ctx *Context) NudgeGuest(id string) {
	ctx.RunnerMutex.Lock()
	defer ctx.RunnerMutex.Unlock()

	runner := ctx.Runners[id]
	if runner != nil {
		runner.Nudge()
	}
}
