package taskcenter

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents a lightweight task progress snapshot returned to clients.
type JobStatus struct {
	TaskID      string         `json:"taskId"`
	Type        string         `json:"type,omitempty"`
	Status      string         `json:"status"`
	Message     string         `json:"message,omitempty"`
	DownloadURL string         `json:"downloadUrl,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	CompletedAt *time.Time     `json:"completedAt,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Store keeps job statuses in-memory for quick lookups.
type Store struct {
	mu   sync.RWMutex
	jobs map[string]*JobStatus
}

var defaultStore = NewStore()

// NewStore returns an empty job store.
func NewStore() *Store {
	return &Store{jobs: make(map[string]*JobStatus)}
}

// DefaultStore exposes the process-wide job store singleton.
func DefaultStore() *Store {
	return defaultStore
}

// Create registers a new job with queued status and optional metadata.
func (s *Store) Create(jobType string, metadata map[string]any) *JobStatus {
	if s == nil {
		return nil
	}
	now := time.Now().UTC()
	job := &JobStatus{
		TaskID:    uuid.NewString(),
		Type:      jobType,
		Status:    "queued",
		CreatedAt: now,
		Metadata:  cloneMetadata(metadata),
	}
	s.mu.Lock()
	s.jobs[job.TaskID] = job
	s.mu.Unlock()
	return cloneJob(job)
}

// Update mutates a job status in-place using the provided mutator.
func (s *Store) Update(taskID string, mutate func(*JobStatus)) (*JobStatus, error) {
	if s == nil {
		return nil, errors.New("job store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[taskID]
	if !ok {
		return nil, errors.New("job not found")
	}
	if mutate != nil {
		mutate(job)
	}
	return cloneJob(job), nil
}

// SetStatus updates job status and optional message.
func (s *Store) SetStatus(taskID, status, message string) (*JobStatus, error) {
	return s.Update(taskID, func(job *JobStatus) {
		job.Status = status
		if message != "" {
			job.Message = message
		}
	})
}

// Complete marks the job with a terminal status and timestamp.
func (s *Store) Complete(taskID, status, message, downloadURL string) (*JobStatus, error) {
	now := time.Now().UTC()
	return s.Update(taskID, func(job *JobStatus) {
		job.Status = status
		job.Message = message
		job.DownloadURL = downloadURL
		job.CompletedAt = &now
	})
}

// Fail sets the job to failed and records the error message.
func (s *Store) Fail(taskID string, err error) (*JobStatus, error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return s.Complete(taskID, "failed", msg, "")
}

// Success marks the job successful with an optional description and download link.
func (s *Store) Success(taskID, message, downloadURL string) (*JobStatus, error) {
	return s.Complete(taskID, "success", message, downloadURL)
}

// Get returns a copy of the job status if present.
func (s *Store) Get(taskID string) (*JobStatus, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.RLock()
	job, ok := s.jobs[taskID]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneJob(job), true
}

func cloneMetadata(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneJob(job *JobStatus) *JobStatus {
	if job == nil {
		return nil
	}
	var completed *time.Time
	if job.CompletedAt != nil {
		ts := *job.CompletedAt
		completed = &ts
	}
	return &JobStatus{
		TaskID:      job.TaskID,
		Type:        job.Type,
		Status:      job.Status,
		Message:     job.Message,
		DownloadURL: job.DownloadURL,
		CreatedAt:   job.CreatedAt,
		CompletedAt: completed,
		Metadata:    cloneMetadata(job.Metadata),
	}
}
