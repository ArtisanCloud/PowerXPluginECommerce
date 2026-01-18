package order

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (s *Service) UpdateOrderShippingAddress(ctx context.Context, tenantUUID, adminID, orderID string, addr *ShippingAddress) (*OrderSummaryDTO, error) {
	if !s.Ready() {
		return nil, ErrOrderServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	adminID = strings.TrimSpace(adminID)
	orderID = strings.TrimSpace(orderID)

	if adminID == "" {
		return nil, ErrAdminRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	if !isValidShippingAddress(addr) {
		return nil, ErrInvalidShippingAddress
	}

	tx := s.deps.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	ord, err := s.OrderRepo.LockByID(ctx, tx, tenantUUID, orderID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if !isOrderEditableStatus(ord.Status) {
		tx.Rollback()
		return nil, ErrOrderNotEditable
	}

	snapJSON, err := json.Marshal(addr)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Model(&ordermodel.Order{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, orderID).
		Updates(map[string]any{
			"shipping_address_snapshot": datatypes.JSON(snapJSON),
			"updated_at":                time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &OrderSummaryDTO{
		OrderID:                 ord.ID,
		OrderNo:                 ord.OrderNo,
		CustomerID:              ord.CustomerID,
		Channel:                 ord.Channel,
		CreatedByType:           ord.CreatedByType,
		Status:                  ord.Status,
		Amounts:                 MoneyDTO{Currency: ord.Currency, Subtotal: ord.SubtotalAmount, Total: ord.TotalAmount},
		ShippingAddressSnapshot: addr,
		CreatedAt:               ord.CreatedAt,
	}, nil
}

func isOrderEditableStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "pending_payment", "draft":
		return true
	default:
		return false
	}
}
