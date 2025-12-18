package product

import (
	"context"
	"strings"
	"sync"
	"time"

	productmetrics "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/product"
	taskcenter "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/taskcenter"
	"github.com/sirupsen/logrus"
)

// Feedback represents a channel or bulk job callback that must update status dashboards.
type Feedback struct {
	TaskID   string
	Channel  string
	State    string
	Message  string
	Resource string
}

// SyncWorker consumes feedback messages and reconciles job statuses asynchronously.
type SyncWorker struct {
	jobs    *taskcenter.Store
	logger  *logrus.Entry
	metrics *productmetrics.SPUMetrics

	queue chan Feedback
	wg    sync.WaitGroup
}

// NewSyncWorker creates a worker bound to the shared taskcenter store.
func NewSyncWorker(store *taskcenter.Store, logger *logrus.Entry, metrics *productmetrics.SPUMetrics) *SyncWorker {
	if store == nil {
		store = taskcenter.DefaultStore()
	}
	if logger == nil {
		logger = logrus.New().WithField("component", "spu-sync-worker")
	}
	return &SyncWorker{
		jobs:    store,
		logger:  logger,
		metrics: metrics,
		queue:   make(chan Feedback, 128),
	}
}

// Start begins draining feedback until the provided context is canceled.
func (w *SyncWorker) Start(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case fb, ok := <-w.queue:
				if !ok {
					return
				}
				w.handleFeedback(fb)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop drains remaining messages and waits for shutdown.
func (w *SyncWorker) Stop() {
	close(w.queue)
	w.wg.Wait()
}

// Enqueue records a feedback event for asynchronous processing.
func (w *SyncWorker) Enqueue(fb Feedback) {
	select {
	case w.queue <- fb:
	default:
		if w.logger != nil {
			w.logger.WithField("task_id", fb.TaskID).Warn("feedback queue full, dropping event")
		}
	}
}

func (w *SyncWorker) handleFeedback(fb Feedback) {
	if w.jobs != nil && fb.TaskID != "" {
		status := fb.State
		if status == "" {
			status = "updated"
		}
		message := fb.Message
		if message == "" && fb.Channel != "" {
			message = fb.Channel + " -> " + status
		}
		if fb.Resource != "" {
			_, _ = w.jobs.SetStatus(fb.TaskID, status, message+" ("+fb.Resource+")")
		} else {
			_, _ = w.jobs.SetStatus(fb.TaskID, status, message)
		}
	}
	if w.logger != nil {
		fields := logrus.Fields{
			"task_id": fb.TaskID,
			"state":   fb.State,
			"channel": fb.Channel,
		}
		if fb.Resource != "" {
			fields["resource"] = fb.Resource
		}
		w.logger.WithFields(fields).Info("processed channel feedback")
	}
	if w.metrics != nil && fb.Channel != "" {
		state := strings.ToLower(fb.State)
		switch state {
		case "success", "published":
			w.metrics.RecordChannelSync(true)
		case "failed", "withheld":
			w.metrics.RecordChannelSync(false)
		}
	}
}

// StartPolling is a helper that periodically emits heartbeat events for long-running jobs.
func (w *SyncWorker) StartPolling(ctx context.Context, interval time.Duration, taskID string) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				w.Enqueue(Feedback{TaskID: taskID, State: "running", Message: "渠道同步中"})
			case <-ctx.Done():
				return
			}
		}
	}()
}
