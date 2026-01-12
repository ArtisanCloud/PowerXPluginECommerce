package order

import (
	"context"
	"errors"
	"strings"

	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type OrderItemRepository struct {
	*repository.BaseRepository[ordermodel.OrderItem]
}

func NewOrderItemRepository(db *gorm.DB) *OrderItemRepository {
	return &OrderItemRepository{BaseRepository: repository.NewBaseRepository[ordermodel.OrderItem](db)}
}

func (r *OrderItemRepository) CreateBatchWithTx(ctx context.Context, tx *gorm.DB, items []*ordermodel.OrderItem) error {
	if r == nil || r.DB == nil {
		return errors.New("order item repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if len(items) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&items).Error
}

func (r *OrderItemRepository) ListByOrderID(ctx context.Context, tenantUUID, orderID string) ([]ordermodel.OrderItem, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order item repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var rows []ordermodel.OrderItem
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).
		Order("created_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
