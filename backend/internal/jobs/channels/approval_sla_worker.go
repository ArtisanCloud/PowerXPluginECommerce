package channels

import (
	"context"
	"errors"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ApprovalReminder dispatches SLA reminders to downstream systems.
type ApprovalReminder interface {
	Send(ctx context.Context, record productmodel.SPUApprovalRecord) error
}

// ApprovalSLAWorker scans pending approvals and escalates overdue entries.
type ApprovalSLAWorker struct {
	db        *gorm.DB
	notifier  ApprovalReminder
	logger    *logrus.Entry
	batchSize int
	now       func() time.Time
}

// NewApprovalSLAWorker constructs a worker with optional notifier/logger.
func NewApprovalSLAWorker(db *gorm.DB, notifier ApprovalReminder, logger *logrus.Entry) *ApprovalSLAWorker {
	if logger == nil {
		logger = logrus.New().WithField("component", "approval-sla-worker")
	}
	return &ApprovalSLAWorker{
		db:        db,
		notifier:  notifier,
		logger:    logger,
		batchSize: 50,
		now:       time.Now,
	}
}

// WithBatchSize overrides the default batch size during tests.
func (w *ApprovalSLAWorker) WithBatchSize(size int) *ApprovalSLAWorker {
	if size > 0 {
		w.batchSize = size
	}
	return w
}

// WithClock overrides the default clock (useful for deterministic tests).
func (w *ApprovalSLAWorker) WithClock(now func() time.Time) *ApprovalSLAWorker {
	if now != nil {
		w.now = now
	}
	return w
}

// Process escalates overdue approvals and emits reminders.
func (w *ApprovalSLAWorker) Process(ctx context.Context) (int, error) {
	if w == nil || w.db == nil {
		return 0, errors.New("approval SLA worker is not initialized")
	}
	now := w.now().UTC()
	var pending []productmodel.SPUApprovalRecord
	if err := w.db.WithContext(ctx).
		Where("status = ? AND sla_due_at IS NOT NULL AND sla_due_at < ?", "pending", now).
		Order("sla_due_at ASC").
		Limit(w.batchSize).
		Find(&pending).Error; err != nil {
		return 0, err
	}
	count := 0
	for _, record := range pending {
		if err := w.db.WithContext(ctx).
			Model(&productmodel.SPUApprovalRecord{}).
			Where("id = ?", record.ID).
			Updates(map[string]any{"status": "escalated", "updated_at": now}).Error; err != nil {
			return count, err
		}
		count++
		if w.notifier != nil {
			if err := w.notifier.Send(ctx, record); err != nil && w.logger != nil {
				w.logger.WithContext(ctx).
					WithError(err).
					WithField("approval_id", record.ID).
					Warn("failed to emit SLA reminder")
			}
		}
	}
	return count, nil
}
