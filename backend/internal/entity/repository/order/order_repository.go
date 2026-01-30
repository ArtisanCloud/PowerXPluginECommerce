package order

import (
	"context"
	"errors"
	"strings"

	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	*repository.BaseRepository[ordermodel.Order]
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{BaseRepository: repository.NewBaseRepository[ordermodel.Order](db)}
}

func (r *OrderRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, order *ordermodel.Order) error {
	if r == nil || r.DB == nil {
		return errors.New("order repository is not initialized")
	}
	if tx == nil {
		return errors.New("transaction is required")
	}
	if order == nil {
		return errors.New("order is required")
	}
	return tx.WithContext(ctx).Create(order).Error
}

func (r *OrderRepository) LockByID(ctx context.Context, tx *gorm.DB, tenantUUID, orderID string) (*ordermodel.Order, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order repository is not initialized")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var row ordermodel.Order
	query := tx.WithContext(ctx)
	if query.Dialector != nil && query.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Where("tenant_uuid = ? AND id = ?", tenantUUID, orderID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, tenantUUID, orderID string) (*ordermodel.Order, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	var row ordermodel.Order
	if err := r.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, orderID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

type OrderListFilter struct {
	OrderNo   string
	CustomerID string
	Status     string
}

type OrderListResult struct {
	Items []ordermodel.Order
	Total int64
}

func (r *OrderRepository) List(ctx context.Context, tenantUUID string, filter OrderListFilter, page, pageSize int) (*OrderListResult, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("order repository is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, repository.ErrTenantUuidRequired
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	query := r.DB.WithContext(ctx).Model(&ordermodel.Order{}).Where("tenant_uuid = ?", tenantUUID)
	if cid := strings.TrimSpace(filter.CustomerID); cid != "" {
		query = query.Where("customer_id = ?", cid)
	}
	if orderNo := strings.TrimSpace(filter.OrderNo); orderNo != "" {
		query = query.Where("order_no ILIKE ?", "%"+orderNo+"%")
	}
	if st := strings.TrimSpace(filter.Status); st != "" {
		query = query.Where("status = ?", st)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var rows []ordermodel.Order
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, err
	}
	return &OrderListResult{Items: rows, Total: total}, nil
}
