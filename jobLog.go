package agent

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mistifyio/kvite"
)

type (
	// JobStatus is the string status of a job
	JobStatus string

	// Job holds information about a job
	Job struct {
		ID        string
		GuestID   string
		Action    string
		QueuedAt  time.Time
		StartedAt time.Time
		UpdatedAt time.Time
		Status    JobStatus
		Message   string
	}

	// JobLog holds the most recent jobs for a guest
	JobLog struct {
		GuestID     string
		Context     *Context
		ModifyMutex sync.RWMutex
		Index       map[string]int
		GuestIndex  map[string][]int
		Jobs        []*Job
	}
)

const (
	// MaxLoggedJobs configures how many jobs to prune the log to
	MaxLoggedJobs int = 1000
	// Queued is the queued job status
	Queued JobStatus = "Queued"
	// Running is the running job status
	Running JobStatus = "Running"
	// Complete is the complete job status
	Complete JobStatus = "Complete"
	// Errored is the errored job status
	Errored JobStatus = "Error"
)

// reindex rebuilds the job id index for a job log. Must be called with
// ModifyMutex held.
func (jobLog *JobLog) reindex() {
	jobLog.Index = make(map[string]int)
	jobLog.GuestIndex = make(map[string][]int)
	for i, job := range jobLog.Jobs {
		jobLog.addIndex(job, i)
	}
}

// addIndex indexes one new job with a known position. Avoids rebuilding the
// entire index for every new job. Must be called with ModifyMutex held.
func (jobLog *JobLog) addIndex(job *Job, position int) {
	jobLog.Index[job.ID] = position
	if job.GuestID != "" {
		jobLog.GuestIndex[job.GuestID] = append(jobLog.GuestIndex[job.GuestID], position)
	}
}

// getJob retrieves a job from the log based on job id. Must be called with
// ModifyMutex held.
func (jobLog *JobLog) getJob(jobID string) (*Job, error) {
	index, ok := jobLog.Index[jobID]
	if !ok {
		return nil, ErrNotFound
	}
	return jobLog.Jobs[index], nil
}

// GetJob retrieves a job from the log based on job id
func (jobLog *JobLog) GetJob(jobID string) (*Job, error) {
	jobLog.ModifyMutex.RLock()
	defer jobLog.ModifyMutex.RUnlock()

	return jobLog.getJob(jobID)
}

// GetLatestJobs returns the latest X jobs in the log
func (jobLog *JobLog) GetLatestJobs(limit int) []*Job {
	jobLog.ModifyMutex.RLock()
	defer jobLog.ModifyMutex.RUnlock()

	if limit <= 0 {
		return make([]*Job, 0)
	}
	if limit > len(jobLog.Jobs) {
		limit = len(jobLog.Jobs)
	}

	// Get the jobs in reverse order, resulting in newest job first
	jobsAsc := jobLog.Jobs[len(jobLog.Jobs)-limit:]
	jobs := make([]*Job, len(jobsAsc))
	for i, job := range jobsAsc {
		jobs[len(jobsAsc)-1-i] = job
	}

	return jobs
}

// GetLatestGuestJobs returns the latest X jobs in the log for a guest
func (jobLog *JobLog) GetLatestGuestJobs(guestID string, limit int) []*Job {
	jobLog.ModifyMutex.RLock()
	defer jobLog.ModifyMutex.RUnlock()

	if limit <= 0 {
		return make([]*Job, 0)
	}

	gi := jobLog.GuestIndex[guestID]
	if limit > len(gi) {
		limit = len(gi)
	}
	jobs := make([]*Job, limit)
	positions := gi[len(gi)-limit:]

	// Create job set in reverse order, resulting in newest job first
	for i := 0; i < len(positions); i++ {
		position := positions[len(positions)-1-i]
		jobs[i] = jobLog.Jobs[position]
	}

	return jobs
}

// AddJob adds a job to the log
func (jobLog *JobLog) AddJob(jobID, guestID, action string) error {
	jobLog.ModifyMutex.Lock()
	defer jobLog.ModifyMutex.Unlock()

	job := &Job{
		ID:        jobID,
		GuestID:   guestID,
		Action:    action,
		QueuedAt:  time.Now(),
		UpdatedAt: time.Now(),
		Status:    Queued,
	}

	// Add and index
	jobLog.Jobs = append(jobLog.Jobs, job)
	jobLog.addIndex(job, len(jobLog.Jobs)-1)

	return jobLog.persist()
}

// UpdateJob updates a job's status and timing information
func (jobLog *JobLog) UpdateJob(jobID string, action string, status JobStatus, message string) error {
	jobLog.ModifyMutex.Lock()
	defer jobLog.ModifyMutex.Unlock()

	job, err := jobLog.getJob(jobID)
	if err != nil {
		return err
	}
	job.Status = status
	job.UpdatedAt = time.Now()
	if (job.StartedAt == time.Time{} && status == Running) {
		job.StartedAt = time.Now()
	}
	job.Message = message
	return jobLog.persist()
}

// persist saves a job log. Must be called with ModifyMutex held.
func (jobLog *JobLog) persist() error {
	return jobLog.Context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("guest_jobs")
		if err != nil {
			return err
		}
		data, err := json.Marshal(jobLog.Jobs)
		if err != nil {
			return err
		}
		return b.Put(jobLog.GuestID, data)
	})
}

// Persist saves a job log
func (jobLog *JobLog) Persist() error {
	jobLog.ModifyMutex.Lock()
	defer jobLog.ModifyMutex.Unlock()

	return jobLog.persist()
}

// prune trims the job log to the max length
func (jobLog *JobLog) prune() error {
	jobLog.ModifyMutex.Lock()
	defer jobLog.ModifyMutex.Unlock()

	n := len(jobLog.Jobs)
	if n <= MaxLoggedJobs {
		return nil
	}

	jobLog.Jobs = jobLog.Jobs[n-MaxLoggedJobs:]
	jobLog.reindex()
	return jobLog.persist()
}

func getLatestGuestJobs(r *HTTPRequest) *HTTPErrorMessage {
	jobLog := r.Context.JobLog
	limitParam := r.Parameter("limit")
	limit := MaxLoggedJobs
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			return r.NewError(err, http.StatusBadRequest)
		}
	}
	return r.JSON(http.StatusOK, jobLog.GetLatestGuestJobs(r.Guest.Id, limit))
}

func getLatestJobs(r *HTTPRequest) *HTTPErrorMessage {
	jobLog := r.Context.JobLog
	limitParam := r.Parameter("limit")
	limit := MaxLoggedJobs
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			return r.NewError(err, http.StatusBadRequest)
		}
	}
	return r.JSON(http.StatusOK, jobLog.GetLatestJobs(limit))
}

func getJobStatus(r *HTTPRequest) *HTTPErrorMessage {
	jobLog := r.Context.JobLog
	job, err := jobLog.GetJob(r.Parameter("jobID"))
	if err != nil {
		code := http.StatusInternalServerError
		if err == ErrNotFound {
			code = http.StatusNotFound
		}
		return r.NewError(err, code)
	}
	return r.JSON(http.StatusOK, job)
}
