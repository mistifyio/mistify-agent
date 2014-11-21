package agent

import (
	"errors"

	"github.com/mistifyio/mistify-agent/config"
	"github.com/mistifyio/mistify-agent/log"
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
		Async:   NewPipelineQueue("async", guestID),
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
	switch pipeline.Type {
	case config.InfoAction:
		if err := gr.Info.Process(pipeline); err != nil {
			return err
		}
	case config.StreamAction:
		if err := gr.Stream.Process(pipeline); err != nil {
			return err
		}
	case config.AsyncAction:
		LogRunnerInfo(gr.GuestID, "async", "", "Queued")
		gr.Async.Enqueue(pipeline)
		return nil
	}
	return nil
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

func NewPipelineQueue(name string, guestID string) *PipelineQueue {
	max := 100
	pq := &PipelineQueue{
		Name:         name,
		GuestID:      guestID,
		PipelineChan: make(chan *Pipeline, max),
		QuitChan:     make(chan struct{}),
	}
	return pq
}

func (pq *PipelineQueue) Enqueue(pipeline *Pipeline) {
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
				err := pipeline.Run()
				if err != nil {
					LogRunnerError(pq.GuestID, pq.Name, pipeline.ID, err.Error())
				} else {
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
	switch {
	case pipelineID != "":
		log.Info("[Info][Guest %s][Runner %s][Pipeline %s] %s", guestID, runnerName, pipelineID, logLine)
	case runnerName != "":
		log.Info("[Info][Guest %s][Runner %s] %s", guestID, runnerName, logLine)
	default:
		log.Info("[Info][Guest %s] %s", guestID, logLine)
	}
}
func LogRunnerError(guestID string, runnerName string, pipelineID string, logLine string) {
	switch {
	case pipelineID != "":
		log.Error("[Error][Guest %s][Runner %s][Pipeline %s] %s", guestID, runnerName, pipelineID, logLine)
	case runnerName != "":
		log.Error("[Error][Guest %s][Runner %s] %s", guestID, runnerName, logLine)
	default:
		log.Error("[Error][Guest %s] %s", guestID, logLine)
	}
}
