package order

import (
	"context"
	"errors"
	"strings"

	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

type OrderEventRepository struct {
	*repository.BaseRepository[ordermodel.OrderEvent]
}

func NewOrderEventRepository(db *gorm.DB) *OrderEventRepository {
	return &OrderEventRepository{BaseRepository: repository.NewBaseRepository[ordermodel.OrderEvent](db)}
}

func (r *OrderEventRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, event *ordermodel.OrderEvent) error {
	if r == nil || r.DB == nil {
		return errors.New("order event repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if event == nil {
		return errors.New("event is required")
	}
	return tx.WithContext(ctx).Create(event).Error
}

func (r *OrderEventRepository) ListByOrderID(ctx context.Context, tenantUUID, orderID string, limit int) ([]ordermodel.OrderEvent, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order event repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	if limit <= 0 || limit > 200 {
		limit = 200
	}
	var rows []ordermodel.OrderEvent
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ?", tenantUUID, orderID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
