package after_sales

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var ErrOrderSyncUnavailable = errors.New("after-sales order sync unavailable")

type OrderSyncService struct {
	deps      *app.Deps
	eventRepo *orderrepo.OrderEventRepository
}

func NewOrderSyncService(deps *app.Deps) *OrderSyncService {
	if deps == nil || deps.DB == nil {
		return &OrderSyncService{deps: deps}
	}
	return &OrderSyncService{
		deps:      deps,
		eventRepo: orderrepo.NewOrderEventRepository(deps.DB),
	}
}

func (s *OrderSyncService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.eventRepo != nil
}

func (s *OrderSyncService) SyncCaseTransitionWithTx(ctx context.Context, tx *gorm.DB, tenantUUID, action, operatorID, note string, row *AfterSalesModel.AfterSaleCase) error {
	if !s.Ready() {
		return ErrOrderSyncUnavailable
	}
	if tx == nil || row == nil {
		return nil
	}
	eventType := eventTypeByAction(action)
	if eventType == "" {
		return nil
	}
	payload := map[string]any{
		"case_id":                  strings.TrimSpace(row.ID),
		"case_no":                  strings.TrimSpace(row.CaseNo),
		"case_type":                strings.TrimSpace(row.CaseType),
		"status":                   strings.TrimSpace(row.Status),
		"reason_code":              strings.TrimSpace(row.ReasonCode),
		"requested_amount_minor":   row.RequestedAmountMinor,
		"currency":                 strings.TrimSpace(row.Currency),
		"updated_at":               time.Now().UTC().Format(time.RFC3339Nano),
		"note":                     strings.TrimSpace(note),
		"exchange_boundary_locked": strings.EqualFold(strings.TrimSpace(row.CaseType), "exchange"),
	}
	if strings.EqualFold(strings.TrimSpace(row.CaseType), "exchange") {
		payload["fulfillment_guard"] = map[string]any{
			"auto_reship":            false,
			"auto_inventory_reserve": false,
			"mode":                   "manual_only",
		}
	}
	body, _ := json.Marshal(payload)
	event := &ordermodel.OrderEvent{
		TenantUUID:   strings.TrimSpace(tenantUUID),
		OrderID:      strings.TrimSpace(row.OrderID),
		EventType:    eventType,
		OperatorType: "admin",
		Operator:     strings.TrimSpace(operatorID),
		Payload:      datatypes.JSON(body),
	}
	return s.eventRepo.CreateWithTx(ctx, tx, event)
}

func eventTypeByAction(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "accept":
		return "after_sales.accepted"
	case "review":
		return "after_sales.reviewing"
	case "approve":
		return "after_sales.approved"
	case "reject":
		return "after_sales.rejected"
	case "complete":
		return "after_sales.completed"
	case "close":
		return "after_sales.closed"
	default:
		return ""
	}
}

// Ensure table constants are referenced by this linkage service.
var _ = models.TableOrderEvents
