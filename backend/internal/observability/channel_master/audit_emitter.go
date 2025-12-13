package channel_master

import (
	"context"

	"github.com/sirupsen/logrus"
)

// AuditEmitter writes structured audit logs for channel operations.
type AuditEmitter struct {
	logger *logrus.Entry
}

// NewAuditEmitter instantiates an emitter with the provided logger.
func NewAuditEmitter(logger *logrus.Entry) *AuditEmitter {
	if logger == nil {
		logger = logrus.New().WithField("component", "channel-master-audit")
	}
	return &AuditEmitter{logger: logger}
}

// EmitChannelAudit satisfies the service AuditEmitter interface.
func (e *AuditEmitter) EmitChannelAudit(ctx context.Context, action string, payload map[string]any) error {
	if e == nil || e.logger == nil {
		return nil
	}
	fields := logrus.Fields{"action": action}
	for k, v := range payload {
		fields[k] = v
	}
	if ctx != nil {
		if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
			fields["request_id"] = reqID
		}
	}
	e.logger.WithFields(fields).Info("channel audit event")
	return nil
}
