package agent

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/config"
)

type (
	GuestRunner struct {
		Context  *Context
		GuestID  string
		Info     *SyncThrottle
		Stream   *SyncThrottle
		Async    *PipelineQueue
		QuitChan chan struct{}
	}

	SyncThrottle struct {
		GuestID        string
		Name           string
		ConcurrentChan chan struct{}
		QuitChan       chan struct{}
	}

	PipelineQueue struct {
		GuestID      string
		Name         string
		Context      *Context
		PipelineChan chan *Pipeline
		QuitChan     chan struct{}
	}
)

func (context *Context) NewGuestRunner(guestID string, maxInfo uint, maxStream uint) *GuestRunner {
	// Prevent others from modifying at the same time
	context.GuestRunnerMutex.Lock()
	defer context.GuestRunnerMutex.Unlock()

	// Check if one already exists
	runner, ok := context.GuestRunners[guestID]
	if ok {
		return runner
	}

	// Create a new runner
	runner = &GuestRunner{
		Context: context,
		GuestID: guestID,
		Info:    NewSyncThrottle("info", guestID, maxInfo),
		Stream:  NewSyncThrottle("stream", guestID, maxStream),
		Async:   NewPipelineQueue("async", guestID, context),
	}

	runner.Async.Process()

	context.GuestRunners[guestID] = runner
	LogRunnerInfo(guestID, "", "", "Created")

	return runner
}

func (context *Context) DeleteGuestRunner(guestID string) {
	// Prevent others from modifying at the same time
	context.GuestRunnerMutex.Lock()
	defer context.GuestRunnerMutex.Unlock()

	guestRunner, ok := context.GuestRunners[guestID]
	if ok {
		guestRunner.Quit()
	}

	delete(context.GuestRunners, guestID)

	LogRunnerInfo(guestID, "", "", "Deleted")
}

func (context *Context) GetGuestRunner(guestID string) (*GuestRunner, error) {
	runner, ok := context.GuestRunners[guestID]
	if !ok {
		return nil, errors.New("guest runner not found")
	}
	return runner, nil
}

func (gr *GuestRunner) Quit() {
	LogRunnerInfo(gr.GuestID, "", "", "Quiting")
	gr.Async.Quit()
}

func (gr *GuestRunner) Process(pipeline *Pipeline) error {
	var err error
	switch pipeline.Type {
	case config.InfoAction:
		err = gr.Info.Process(pipeline)
	case config.StreamAction:
		err = gr.Stream.Process(pipeline)
	case config.AsyncAction:
		LogRunnerInfo(gr.GuestID, "async", "", "Queued")
		gr.Async.Enqueue(pipeline)
	}
	return err
}

func NewSyncThrottle(name string, guestID string, maxConcurrency uint) *SyncThrottle {
	st := &SyncThrottle{
		Name:           name,
		GuestID:        guestID,
		ConcurrentChan: make(chan struct{}, maxConcurrency),
	}
	for i := uint(0); i < maxConcurrency; i++ {
		st.ConcurrentChan <- struct{}{}
	}

	return st
}

func (st *SyncThrottle) Process(pipeline *Pipeline) error {
	st.Reserve()
	defer st.Release()

	return pipeline.Run()
}

func (st *SyncThrottle) Reserve() {
	<-st.ConcurrentChan
	return
}

func (st *SyncThrottle) Release() {
	st.ConcurrentChan <- struct{}{}
	return
}

func NewPipelineQueue(name string, guestID string, context *Context) *PipelineQueue {
	max := 100
	pq := &PipelineQueue{
		Name:         name,
		GuestID:      guestID,
		PipelineChan: make(chan *Pipeline, max),
		QuitChan:     make(chan struct{}),
		Context:      context,
	}
	return pq
}

func (pq *PipelineQueue) Enqueue(pipeline *Pipeline) {
	if err := pq.Context.GuestJobLogs[pq.GuestID].AddJob(pipeline.ID, pipeline.Action); err != nil {
		LogRunnerError(pq.GuestID, pq.Name, pipeline.ID, err.Error())
	}
	pq.PipelineChan <- pipeline
	return
}

func (pq *PipelineQueue) Process() {
	go func() {
		for {
			select {
			case <-pq.QuitChan:
				LogRunnerInfo(pq.GuestID, pq.Name, "", "Quitting")
				return
			case pipeline := <-pq.PipelineChan:
				if err := pq.Context.GuestJobLogs[pq.GuestID].UpdateJob(pipeline.ID, pipeline.Action, Running, ""); err != nil {
					LogRunnerError(pq.GuestID, pq.Name, pipeline.ID, err.Error())
				}
				if err := pipeline.Run(); err != nil {
					if err := pq.Context.GuestJobLogs[pq.GuestID].UpdateJob(pipeline.ID, pipeline.Action, Errored, err.Error()); err != nil {
						LogRunnerError(pq.GuestID, pq.Name, pipeline.ID, err.Error())
					}
					LogRunnerError(pq.GuestID, pq.Name, pipeline.ID, err.Error())
				} else {
					if err := pq.Context.GuestJobLogs[pq.GuestID].UpdateJob(pipeline.ID, pipeline.Action, Complete, ""); err != nil {
						LogRunnerError(pq.GuestID, pq.Name, pipeline.ID, err.Error())
					}
					LogRunnerInfo(pq.GuestID, pq.Name, pipeline.ID, "Success")
				}
			}
		}
	}()
}

func (pq *PipelineQueue) Quit() {
	go func() {
		pq.QuitChan <- struct{}{}
	}()
}

func LogRunnerInfo(guestID string, runnerName string, pipelineID string, logLine string) {
	log.WithFields(log.Fields{
		"guest":    guestID,
		"runner":   runnerName,
		"pipeline": pipelineID,
	}).Info(logLine)
}

func LogRunnerError(guestID string, runnerName string, pipelineID string, logLine string) {
	log.WithFields(log.Fields{
		"guest":    guestID,
		"runner":   runnerName,
		"pipeline": pipelineID,
	}).Error(logLine)
}
