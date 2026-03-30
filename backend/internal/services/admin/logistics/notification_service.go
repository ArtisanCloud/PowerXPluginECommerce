package logistics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NotificationService struct {
	tplRepo     *LogisticsRepo.NotificationTemplateRepository
	recordRepo  *LogisticsRepo.NotificationRecordRepository
	waybillRepo *LogisticsRepo.WaybillRepository
}

func NewNotificationService(deps *app.Deps) *NotificationService {
	if deps == nil || deps.DB == nil {
		return &NotificationService{}
	}
	return &NotificationService{
		tplRepo:     LogisticsRepo.NewNotificationTemplateRepository(deps.DB),
		recordRepo:  LogisticsRepo.NewNotificationRecordRepository(deps.DB),
		waybillRepo: LogisticsRepo.NewWaybillRepository(deps.DB),
	}
}

type UpsertNotificationTemplateRequest struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Event    string         `json:"event"`
	Channel  string         `json:"channel,omitempty"`
	Title    string         `json:"title,omitempty"`
	Body     string         `json:"body"`
	Enabled  *bool          `json:"enabled,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type SendNotificationRequest struct {
	WaybillID      string         `json:"waybill_id"`
	Event          string         `json:"event"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
}

type RetryNotificationRequest struct {
	RecordID string `json:"record_id"`
}

type NotificationSendResult struct {
	Record            *LogisticsModel.NotificationRecord `json:"record"`
	IdempotencyStatus string                             `json:"idempotency_status"`
}

func (s *NotificationService) ListTemplates(ctx context.Context, tenantUUID string) ([]LogisticsModel.NotificationTemplate, error) {
	if s == nil || s.tplRepo == nil {
		return nil, errors.New("notification service unavailable")
	}
	return s.tplRepo.List(withTenantContext(ctx, tenantUUID))
}

func (s *NotificationService) UpsertTemplate(ctx context.Context, tenantUUID string, req UpsertNotificationTemplateRequest) (*LogisticsModel.NotificationTemplate, error) {
	if s == nil || s.tplRepo == nil {
		return nil, errors.New("notification service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	name := strings.TrimSpace(req.Name)
	event := normalizeNotificationEvent(req.Event)
	body := strings.TrimSpace(req.Body)
	if name == "" || event == "" || body == "" {
		return nil, errors.New("name/event/body required")
	}
	channel := normalizeNotificationChannel(req.Channel)
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	metaBytes, _ := jsonBytes(req.Metadata, []byte("{}"))
	if strings.TrimSpace(req.ID) != "" {
		row, err := s.tplRepo.GetByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		row.Name = name
		row.Event = event
		row.Channel = channel
		row.Title = strings.TrimSpace(req.Title)
		row.Body = body
		row.Enabled = enabled
		row.Metadata = datatypes.JSON(metaBytes)
		if err := s.tplRepo.Save(ctx, row); err != nil {
			return nil, err
		}
		return row, nil
	}
	row := &LogisticsModel.NotificationTemplate{
		ID:       utils.NewUUID(),
		Name:     name,
		Event:    event,
		Channel:  channel,
		Title:    strings.TrimSpace(req.Title),
		Body:     body,
		Enabled:  enabled,
		Metadata: datatypes.JSON(metaBytes),
	}
	if err := s.tplRepo.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *NotificationService) ListRecords(ctx context.Context, tenantUUID, status string) ([]LogisticsModel.NotificationRecord, error) {
	if s == nil || s.recordRepo == nil {
		return nil, errors.New("notification service unavailable")
	}
	return s.recordRepo.List(withTenantContext(ctx, tenantUUID), strings.TrimSpace(status))
}

func (s *NotificationService) Send(ctx context.Context, tenantUUID string, req SendNotificationRequest) (*NotificationSendResult, error) {
	if s == nil || s.tplRepo == nil || s.recordRepo == nil || s.waybillRepo == nil {
		return nil, errors.New("notification service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	event := normalizeNotificationEvent(req.Event)
	waybillID := strings.TrimSpace(req.WaybillID)
	if event == "" || waybillID == "" {
		return nil, errors.New("waybill_id/event required")
	}
	waybill, err := s.waybillRepo.GetByID(ctx, waybillID)
	if err != nil {
		return nil, err
	}
	tpl, err := s.tplRepo.FindEnabledByEvent(ctx, event)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("notification template not found")
		}
		return nil, err
	}
	idem := strings.TrimSpace(req.IdempotencyKey)
	if idem == "" {
		idem = buildNotificationIdemKey(event, waybill.ID)
	}
	existed, err := s.recordRepo.GetByIdempotencyKey(ctx, idem)
	if err == nil && existed != nil {
		return &NotificationSendResult{Record: existed, IdempotencyStatus: "replayed"}, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	renderData := buildNotificationData(waybill, req.Payload)
	title := renderNotificationTemplate(tpl.Title, renderData)
	body := renderNotificationTemplate(tpl.Body, renderData)
	payloadBytes, _ := jsonBytes(renderData, []byte("{}"))

	record := &LogisticsModel.NotificationRecord{
		ID:             utils.NewUUID(),
		TemplateID:     tpl.ID,
		WaybillID:      waybill.ID,
		Event:          event,
		Channel:        tpl.Channel,
		Status:         "pending",
		AttemptCount:   1,
		MaxAttempts:    3,
		IdempotencyKey: idem,
		LastError:      "",
		RenderedTitle:  title,
		RenderedBody:   body,
		Payload:        datatypes.JSON(payloadBytes),
	}
	if shouldNotificationFail(renderData) {
		record.Status = "failed"
		record.LastError = "simulated delivery failure"
	} else {
		record.Status = "sent"
		now := time.Now().UTC()
		record.SentAt = &now
	}
	if err := s.recordRepo.Create(ctx, record); err != nil {
		if LogisticsRepo.IsDuplicateConstraintError(err) {
			existed, getErr := s.recordRepo.GetByIdempotencyKey(ctx, idem)
			if getErr == nil && existed != nil {
				return &NotificationSendResult{Record: existed, IdempotencyStatus: "replayed"}, nil
			}
		}
		return nil, err
	}
	return &NotificationSendResult{Record: record, IdempotencyStatus: "created"}, nil
}

func (s *NotificationService) Retry(ctx context.Context, tenantUUID string, req RetryNotificationRequest) (*LogisticsModel.NotificationRecord, error) {
	if s == nil || s.recordRepo == nil {
		return nil, errors.New("notification service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	record, err := s.recordRepo.GetByID(ctx, req.RecordID)
	if err != nil {
		return nil, err
	}
	if record.Status != "failed" {
		return nil, errors.New("only failed record can retry")
	}
	record.AttemptCount++
	if record.AttemptCount > record.MaxAttempts {
		record.LastError = "max attempts exceeded"
		if err := s.recordRepo.Save(ctx, record); err != nil {
			return nil, err
		}
		return record, nil
	}
	record.Status = "sent"
	record.LastError = ""
	now := time.Now().UTC()
	record.SentAt = &now
	if err := s.recordRepo.Save(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

func normalizeNotificationEvent(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "shipped":
		return "shipped"
	case "out_for_delivery":
		return "out_for_delivery"
	case "signed":
		return "signed"
	case "exception":
		return "exception"
	default:
		return ""
	}
}

func normalizeNotificationChannel(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "email":
		return "email"
	case "sms":
		return "sms"
	case "webhook":
		return "webhook"
	default:
		return "sms"
	}
}

func buildNotificationIdemKey(event, waybillID string) string {
	return fmt.Sprintf("%s#%s", strings.TrimSpace(event), strings.TrimSpace(waybillID))
}

func buildNotificationData(waybill *LogisticsModel.Waybill, payload map[string]any) map[string]any {
	out := map[string]any{
		"waybill_id": waybill.ID,
		"waybill_no": waybill.WaybillNo,
		"order_id":   waybill.OrderID,
		"status":     waybill.Status,
	}
	for k, v := range payload {
		out[strings.TrimSpace(k)] = v
	}
	return out
}

func renderNotificationTemplate(tpl string, data map[string]any) string {
	out := strings.TrimSpace(tpl)
	if out == "" {
		return out
	}
	for k, v := range data {
		out = strings.ReplaceAll(out, "{{"+k+"}}", fmt.Sprintf("%v", v))
	}
	return out
}

func shouldNotificationFail(data map[string]any) bool {
	raw, ok := data["simulate_fail"]
	if !ok || raw == nil {
		return false
	}
	val, ok := raw.(bool)
	return ok && val
}
