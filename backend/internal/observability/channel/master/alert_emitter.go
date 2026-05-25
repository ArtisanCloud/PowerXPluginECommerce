package master

import (
	"context"
	pxlogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"

	"github.com/sirupsen/logrus"
)

// AlertEmitter writes structured alert logs for downstream collectors.
type AlertEmitter struct {
	logger *logrus.Entry
}

// NewAlertEmitter builds emitter.
func NewAlertEmitter(logger *logrus.Entry) *AlertEmitter {
	if logger == nil {
		logger = pxlogger.WithField("component", "channel-master-alert")
	}
	return &AlertEmitter{logger: logger}
}

// Emit writes alert metadata.
func (e *AlertEmitter) Emit(ctx context.Context, event string, payload map[string]any) {
	if e == nil || e.logger == nil {
		return
	}
	fields := logrus.Fields{"event": event}
	for k, v := range payload {
		fields[k] = v
	}
	if ctx != nil {
		if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
			fields["request_id"] = reqID
		}
	}
	e.logger.WithFields(fields).Info("channel alert")
}
