package logistics

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type WebhookService struct {
	waybillSvc *WaybillService
}

func NewWebhookService(deps *app.Deps) *WebhookService {
	return &WebhookService{waybillSvc: NewWaybillService(deps)}
}

type HandleWebhookRequest struct {
	EventID    string         `json:"event_id"`
	WaybillNo  string         `json:"waybill_no"`
	Status     string         `json:"status"`
	OccurredAt *time.Time     `json:"occurred_at,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type WebhookAck struct {
	Accepted          bool   `json:"accepted"`
	IdempotencyKey    string `json:"idempotency_key"`
	IdempotencyStatus string `json:"idempotency_status"`
}

func (s *WebhookService) Handle(ctx context.Context, tenantUUID string, req HandleWebhookRequest) (*WebhookAck, error) {
	if s == nil || s.waybillSvc == nil || s.waybillSvc.waybillRepo == nil {
		return nil, errors.New("webhook service unavailable")
	}
	if strings.TrimSpace(req.EventID) == "" || strings.TrimSpace(req.WaybillNo) == "" || strings.TrimSpace(req.Status) == "" {
		return nil, errors.New("event_id/waybill_no/status required")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	wb, err := s.waybillSvc.waybillRepo.GetByWaybillNo(ctx, req.WaybillNo)
	if err != nil {
		return nil, err
	}
	_, idem, err := s.waybillSvc.AppendTracking(ctx, tenantUUID, wb.ID, AppendTrackingRequest{
		EventID:    req.EventID,
		Status:     req.Status,
		Source:     "provider",
		OccurredAt: req.OccurredAt,
		Payload:    req.Payload,
	})
	if err != nil {
		return nil, err
	}
	return &WebhookAck{
		Accepted:          true,
		IdempotencyKey:    strings.TrimSpace(req.EventID) + ":" + strings.TrimSpace(req.WaybillNo),
		IdempotencyStatus: idem,
	}, nil
}
