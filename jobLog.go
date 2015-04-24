package agent

import (
	"encoding/json"
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
		Index       map[string]int
		GuestIndex  map[string][]int
		Jobs        []*Job
		ModifyMutex sync.Mutex
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

// reindex rebuilds the job id index for a job log
func (jobLog *JobLog) reindex() {
	jobLog.Index = make(map[string]int)
	jobLog.GuestIndex = make(map[string][]int)
	for i, job := range jobLog.Jobs {
		jobLog.addIndex(job, i)
	}
}

func (jobLog *JobLog) addIndex(job *Job, position int) {
	jobLog.Index[job.ID] = position
	if job.GuestID != "" {
		jobLog.GuestIndex[job.GuestID] = append(jobLog.GuestIndex[job.GuestID], position)
	}
}

// GetJob retrieves a job from the log based on job id
func (jobLog *JobLog) GetJob(jobID string) (*Job, error) {
	index, ok := jobLog.Index[jobID]
	if !ok {
		return nil, ErrNotFound
	}
	return jobLog.Jobs[index], nil
}

// GetLatestJobs returns the latest X jobs in the log
func (jobLog *JobLog) GetLatestJobs(limit int) []*Job {
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

	return jobLog.Persist()
}

// UpdateJob updates a job's status and timing information
func (jobLog *JobLog) UpdateJob(jobID string, action string, status JobStatus, message string) error {
	jobLog.ModifyMutex.Lock()
	defer jobLog.ModifyMutex.Unlock()

	job, err := jobLog.GetJob(jobID)
	if err != nil {
		return err
	}
	job.Status = status
	job.UpdatedAt = time.Now()
	if (job.StartedAt == time.Time{} && status == Running) {
		job.StartedAt = time.Now()
	}
	job.Message = message
	return jobLog.Persist()
}

// Persist saves a job log
func (jobLog *JobLog) Persist() error {
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

func (jobLog *JobLog) prune() error {
	jobLog.ModifyMutex.Lock()
	defer jobLog.ModifyMutex.Unlock()

	n := len(jobLog.Jobs)
	if n <= MaxLoggedJobs {
		return nil
	}

	jobLog.Jobs = jobLog.Jobs[n-MaxLoggedJobs:]
	jobLog.reindex()
	return jobLog.Persist()
}

// CreateJobLog creates a new job log
func (context *Context) CreateJobLog() error {
	// Attempt to load from database
	jobLog, err := context.GetJobLog()
	if err == nil {
		context.JobLog = jobLog
		return nil
	}
	if err != ErrNotFound {
		return err
	}
	// Create a new one
	context.JobLog = &JobLog{
		Context:    context,
		Index:      make(map[string]int),
		GuestIndex: make(map[string][]int),
		Jobs:       make([]*Job, 0, MaxLoggedJobs+1),
	}

	go func() {
		for {
			context.JobLog.prune()
			time.Sleep(60 * time.Second)
		}
	}()

	return nil
}

// GetJobLog retrieves a job log
func (context *Context) GetJobLog() (*JobLog, error) {
	jobLog := &JobLog{
		Context: context,
	}
	err := context.db.Transaction(func(tx *kvite.Tx) error {
		b, err := tx.Bucket("joblog")
		if err != nil {
			return err
		}

		data, err := b.Get("jobs")
		if err != nil {
			return err
		}
		if data == nil {
			return ErrNotFound
		}
		return json.Unmarshal(data, &jobLog.Jobs)
	})
	if err != nil {
		return nil, err
	}
	jobLog.reindex()
	return jobLog, nil
}

func getLatestGuestJobs(r *HTTPRequest) *HTTPErrorMessage {
	jobLog := r.Context.JobLog
	limitParam := r.Parameter("limit")
	limit := MaxLoggedJobs
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			return r.NewError(err, 400)
		}
	}
	return r.JSON(200, jobLog.GetLatestGuestJobs(r.Guest.Id, limit))
}

func getLatestJobs(r *HTTPRequest) *HTTPErrorMessage {
	jobLog := r.Context.JobLog
	limitParam := r.Parameter("limit")
	limit := MaxLoggedJobs
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			return r.NewError(err, 400)
		}
	}
	return r.JSON(200, jobLog.GetLatestJobs(limit))
}
func getJobStatus(r *HTTPRequest) *HTTPErrorMessage {
	jobLog := r.Context.JobLog
	job, err := jobLog.GetJob(r.Parameter("jobID"))
	if err != nil {
		code := 500
		if err == ErrNotFound {
			code = 404
		}
		return r.NewError(err, code)
	}
	return r.JSON(200, job)
}
