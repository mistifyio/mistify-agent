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
	// GuestJobStatus is the string status of a job
	GuestJobStatus string

	// GuestJob holds information about a job
	GuestJob struct {
		ID        string
		Action    string
		QueuedAt  time.Time
		StartedAt time.Time
		UpdatedAt time.Time
		Status    GuestJobStatus
		Message   string
	}

	// GuestJobLog holds the most recent jobs for a guest
	GuestJobLog struct {
		GuestID     string
		Context     *Context
		Index       map[string]int
		Jobs        []*GuestJob
		ModifyMutex sync.Mutex
	}
)

const (
	// MaxLoggedJobs configures how many jobs to prune the log to
	MaxLoggedJobs int = 100
	// Queued is the queued job status
	Queued GuestJobStatus = "Queued"
	// Running is the running job status
	Running GuestJobStatus = "Running"
	// Complete is the complete job status
	Complete GuestJobStatus = "Complete"
	// Errored is the errored job status
	Errored GuestJobStatus = "Error"
)

// ReIndex rebuilds the job id index for a job log
func (guestJobLog *GuestJobLog) ReIndex() {
	index := make(map[string]int)
	for i, guestJob := range guestJobLog.Jobs {
		index[guestJob.ID] = i
	}
	guestJobLog.Index = index
}

// GetJob retrieves a job from the log based on job id
func (guestJobLog *GuestJobLog) GetJob(jobID string) (*GuestJob, error) {
	index, ok := guestJobLog.Index[jobID]
	if !ok {
		return nil, ErrNotFound
	}
	return guestJobLog.Jobs[index], nil
}

// GetLatestJobs returns the latest X jobs in the log
func (guestJobLog *GuestJobLog) GetLatestJobs(limit int) []*GuestJob {
	if limit <= 0 {
		return make([]*GuestJob, 0)
	}
	if limit > len(guestJobLog.Jobs) {
		limit = len(guestJobLog.Jobs)
	}

	return guestJobLog.Jobs[:limit]
}

// AddJob adds a job to the log
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

// UpdateJob updates a job's status and timing information
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

// Persist saves a job log
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

// CreateGuestJobLog creates a new job log for a guest
func (context *Context) CreateGuestJobLog(guestID string) error {
	// Attempt to load from database
	guestJobLog, err := context.GetGuestJobLog(guestID)
	if err == nil {
		context.GuestJobLogs[guestID] = guestJobLog
		return nil
	}
	if err != ErrNotFound {
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

// GetGuestJobLog retrieves a guest's job log
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
			return ErrNotFound
		}
		return json.Unmarshal(data, &guestJobLog.Jobs)
	})
	if err != nil {
		return nil, err
	}
	guestJobLog.ReIndex()
	return guestJobLog, nil
}

// DeleteGuestJobLog removes a guest's job log
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

func getLatestJobs(r *HTTPRequest) *HTTPErrorMessage {
	guestJobLog, ok := r.Context.GuestJobLogs[r.Guest.Id]
	if !ok {
		return r.NewError(ErrNotFound, 404)
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

func getJobStatus(r *HTTPRequest) *HTTPErrorMessage {
	guestJobLog, ok := r.Context.GuestJobLogs[r.Guest.Id]
	if !ok {
		return r.NewError(ErrNotFound, 404)
	}
	guestJob, err := guestJobLog.GetJob(r.Parameter("jobID"))
	if err != nil {
		code := 500
		if err == ErrNotFound {
			code = 404
		}
		return r.NewError(err, code)
	}
	return r.JSON(200, guestJob)
}
