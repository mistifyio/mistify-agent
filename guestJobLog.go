package agent

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/mistifyio/kvite"
)

type (
	GuestJobStatus string

	GuestJob struct {
		ID        string
		Action    string
		QueuedAt  time.Time
		StartedAt time.Time
		UpdatedAt time.Time
		Status    GuestJobStatus
		Message   string
	}

	GuestJobLog struct {
		GuestID     string
		Context     *Context
		Index       map[string]int
		Jobs        []*GuestJob
		ModifyMutex sync.Mutex
	}
)

const (
	MaxLoggedJobs int            = 100
	Queued        GuestJobStatus = "Queued"
	Running       GuestJobStatus = "Running"
	Complete      GuestJobStatus = "Complete"
	Errored       GuestJobStatus = "Error"
)

func (guestJobLog *GuestJobLog) ReIndex() {
	index := make(map[string]int)
	for i, guestJob := range guestJobLog.Jobs {
		index[guestJob.ID] = i
	}
	guestJobLog.Index = index
}

func (guestJobLog *GuestJobLog) GetJob(jobID string) (*GuestJob, error) {
	index, ok := guestJobLog.Index[jobID]
	if !ok {
		return nil, NotFound
	}
	return guestJobLog.Jobs[index], nil
}

func (guestJob *GuestJobLog) GetLatestJobs(limit int) []*GuestJob {
	if limit <= 0 {
		return make([]*GuestJob, 0)
	}
	if limit > len(guestJob.Jobs) {
		limit = len(guestJob.Jobs)
	}

	return guestJob.Jobs[:limit]
}

func (guestJobLog *GuestJobLog) AddJob(jobID string, action string) error {
	guestJobLog.ModifyMutex.Lock()
	defer guestJobLog.ModifyMutex.Unlock()

	guestJob := &GuestJob{
		ID:        jobID,
		Action:    action,
		QueuedAt:  time.Now(),
		UpdatedAt: time.Now(),
		Status:    Queued,
	}

	// Prepend
	newJobs := make([]*GuestJob, 1, MaxLoggedJobs+1)
	newJobs[0] = guestJob
	newJobs = append(newJobs, guestJobLog.Jobs...)

	// Maintain the max length
	if len(newJobs) == MaxLoggedJobs {
		newJobs = newJobs[:MaxLoggedJobs-1]
	}

	guestJobLog.Jobs = newJobs
	guestJobLog.ReIndex()

	return guestJobLog.Persist()
}

func (guestJobLog *GuestJobLog) UpdateJob(jobID string, action string, status GuestJobStatus, message string) error {
	guestJobLog.ModifyMutex.Lock()
	defer guestJobLog.ModifyMutex.Unlock()

	guestJob, err := guestJobLog.GetJob(jobID)
	if err != nil {
		return err
	}
	guestJob.Status = status
	guestJob.UpdatedAt = time.Now()
	if (guestJob.StartedAt == time.Time{} && status == Running) {
		guestJob.StartedAt = time.Now()
	}
	guestJob.Message = message
	return guestJobLog.Persist()
}

func (guestJobLog *GuestJobLog) Persist() error {
	return guestJobLog.Context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guest_jobs")
		if err != nil {
			return err
		}
		data, err := json.Marshal(guestJobLog.Jobs)
		if err != nil {
			return err
		}
		return b.Put(guestJobLog.GuestID, data)
	})
}

func (context *Context) CreateGuestJobLog(guestID string) error {
	// Attempt to load from database
	guestJobLog, err := context.GetGuestJobLog(guestID)
	if err == nil {
		context.GuestJobLogs[guestID] = guestJobLog
		return nil
	}
	if err != NotFound {
		return err
	}
	// Create a new one
	guestJobLog = &GuestJobLog{
		GuestID: guestID,
		Context: context,
		Index:   make(map[string]int),
		Jobs:    make([]*GuestJob, 0, MaxLoggedJobs+1),
	}
	context.GuestJobLogs[guestID] = guestJobLog
	// Don't bother persisting a new empty job log
	return nil
}

func (context *Context) GetGuestJobLog(guestID string) (*GuestJobLog, error) {
	guestJobLog := &GuestJobLog{
		GuestID: guestID,
		Context: context,
	}
	err := context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guest_jobs")
		if err != nil {
			return err
		}

		data, err := b.Get(guestID)
		if err != nil {
			return err
		}
		if data == nil {
			return NotFound
		}
		return json.Unmarshal(data, &guestJobLog.Jobs)
	})
	if err != nil {
		return nil, err
	}
	guestJobLog.ReIndex()
	return guestJobLog, nil
}

func (context *Context) DeleteGuestJobLog(guestID string) error {
	return context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guest_jobs")
		if err != nil {
			return err
		}
		err = b.Delete(guestID)
		if err != nil {
			return err
		}
		delete(context.GuestJobLogs, guestID)
		return nil
	})
}

func getLatestJobs(r *HttpRequest) *HttpErrorMessage {
	guestJobLog, ok := r.Context.GuestJobLogs[r.Guest.Id]
	if !ok {
		return r.NewError(NotFound, 404)
	}
	limitParam := r.Parameter("limit")
	limit := MaxLoggedJobs
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			return r.NewError(err, 400)
		}
	}
	fmt.Println("limit", limit)
	return r.JSON(200, guestJobLog.GetLatestJobs(limit))
}

func getJobStatus(r *HttpRequest) *HttpErrorMessage {
	guestJobLog, ok := r.Context.GuestJobLogs[r.Guest.Id]
	if !ok {
		return r.NewError(NotFound, 404)
	}
	guestJob, err := guestJobLog.GetJob(r.Parameter("jobID"))
	if err != nil {
		code := 500
		if err == NotFound {
			code = 404
		}
		return r.NewError(err, code)
	}
	return r.JSON(200, guestJob)
}
