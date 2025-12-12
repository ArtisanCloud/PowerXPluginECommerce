package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PublishTask reflects a queued asynchronous publish request.
type PublishTask struct {
	TaskID     string   `json:"taskId"`
	TenantUUID string   `json:"tenantUuid"`
	SPUID      string   `json:"spuId"`
	VersionID  string   `json:"versionId"`
	Channels   []string `json:"channels"`
	Status     string   `json:"status"`
}

// Publisher declares behavior required by the SPU service to enqueue channel tasks.
type Publisher interface {
	EnqueuePublish(ctx context.Context, tenantUUID, spuID, versionID string, channels []string) (*PublishTask, error)
}

// AsyncPublisher is a lightweight implementation that logs the intent.
type AsyncPublisher struct {
	logger *logrus.Entry
}

// NewAsyncPublisher builds an async publisher backed by structured logging.
func NewAsyncPublisher(logger *logrus.Entry) *AsyncPublisher {
	if logger == nil {
		logger = logrus.New().WithField("component", "channels-publisher")
	}
	return &AsyncPublisher{logger: logger}
}

// EnqueuePublish records a channel publish task. Real implementations can swap this out for queues.
func (p *AsyncPublisher) EnqueuePublish(ctx context.Context, tenantUUID, spuID, versionID string, channels []string) (*PublishTask, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("channels are required for publish")
	}
	filtered := make([]string, 0, len(channels))
	for _, ch := range channels {
		c := strings.TrimSpace(ch)
		if c == "" {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("channels are required for publish")
	}
	task := &PublishTask{
		TaskID:     uuid.NewString(),
		TenantUUID: tenantUUID,
		SPUID:      spuID,
		VersionID:  versionID,
		Channels:   filtered,
		Status:     "queued",
	}
	if p.logger != nil {
		p.logger.WithContext(ctx).WithFields(logrus.Fields{
			"tenant_uuid": tenantUUID,
			"spu_id":      spuID,
			"version_id":  versionID,
			"channels":    filtered,
			"task_id":     task.TaskID,
		}).Info("enqueued channel publish task")
	}
	return task, nil
}
