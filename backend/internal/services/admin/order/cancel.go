package order

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (s *Service) CancelOrder(ctx context.Context, tenantUUID, adminID, orderID, reason string) (*OrderSummaryDTO, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	adminID = strings.TrimSpace(adminID)
	orderID = strings.TrimSpace(orderID)
	reason = strings.TrimSpace(reason)

	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}

	var summary *OrderSummaryDTO
	err := s.OrderRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		order, err := s.OrderRepo.LockByID(ctx, tx, tenantUUID, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}

		if strings.TrimSpace(order.Status) != "pending_payment" {
			return ErrOrderNotCancellable
		}

		var items []ordermodel.OrderItem
		if err := tx.WithContext(ctx).
			Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).
			Order("created_at ASC, id ASC").
			Find(&items).Error; err != nil {
			return err
		}

		for _, it := range items {
			if it.Qty <= 0 || strings.TrimSpace(it.SKUID) == "" {
				continue
			}
			if _, err := s.InventoryRepo.UnlockInventory(ctx, tx, tenantUUID, it.SKUID, defaultWarehouseID, it.Qty); err != nil {
				return err
			}
		}

		now := time.Now().UTC()
		if err := tx.WithContext(ctx).
			Model(&ordermodel.Order{}).
			Where("tenant_uuid = ? AND id = ?", tenantUUID, orderID).
			Updates(map[string]any{
				"status":     "cancelled",
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		eventPayload, _ := json.Marshal(map[string]any{
			"requestId": requestIDFromContext(ctx),
			"reason":    reason,
		})
		event := &ordermodel.OrderEvent{
			ID:           uuid.NewString(),
			TenantUUID:   tenantUUID,
			OrderID:      orderID,
			EventType:    "order.cancelled",
			OperatorType: "admin",
			Operator:     adminID,
			Payload:      datatypes.JSON(eventPayload),
		}
		if err := s.EventRepo.CreateWithTx(ctx, tx, event); err != nil {
			return err
		}

		var shippingSnap *ShippingAddress
		if len(order.ShippingAddressSnap) > 0 {
			var snap ShippingAddress
			if err := json.Unmarshal([]byte(order.ShippingAddressSnap), &snap); err == nil {
				shippingSnap = &snap
			}
		}

		summary = &OrderSummaryDTO{
			OrderID:                 orderID,
			OrderNo:                 order.OrderNo,
			CustomerID:              order.CustomerID,
			Channel:                 order.Channel,
			CreatedByType:           order.CreatedByType,
			Status:                  "cancelled",
			Amounts:                 MoneyDTO{Currency: order.Currency, Subtotal: order.SubtotalAmount, Total: order.TotalAmount},
			ShippingAddressSnapshot: shippingSnap,
			// Keep the original created timestamp.
			CreatedAt: order.CreatedAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return summary, nil
}
