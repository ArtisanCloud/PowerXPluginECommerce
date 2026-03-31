package subscription_reconciliation

import (
	"context"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/logger"
)

// Publisher emits structured reconciliation domain events.
type Publisher struct{}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) Emit(ctx context.Context, tenantUUID, event string, fields map[string]any) {
	if p == nil || event == "" {
		return
	}
	entry := logger.ServiceLogger("subscription_reconciliation").WithContext(ctx).WithField("event", event)
	if tenantUUID != "" {
		entry = entry.WithField("tenant_uuid", tenantUUID)
	}
	for k, v := range fields {
		if k == "" || v == nil {
			continue
		}
		entry = entry.WithField(k, v)
	}
	entry.Info("subscription_reconciliation_event")
}
