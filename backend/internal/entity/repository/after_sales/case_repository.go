package after_sales

import (
	"context"
	"strings"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"gorm.io/gorm"
)

// CaseRepository persists after-sale cases.
type CaseRepository struct {
	*Repository[AfterSalesModel.AfterSaleCase]
}

func NewCaseRepository(db *gorm.DB) *CaseRepository {
	return &CaseRepository{Repository: NewRepository[AfterSalesModel.AfterSaleCase](db)}
}

func (r *CaseRepository) Create(ctx context.Context, row *AfterSalesModel.AfterSaleCase) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.TenantUUID) == "" {
		row.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(row).Error
}

func (r *CaseRepository) GetByIDAndCustomer(ctx context.Context, id, customerID string) (*AfterSalesModel.AfterSaleCase, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row AfterSalesModel.AfterSaleCase
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ? AND customer_id = ?", tenantUUID, strings.TrimSpace(id), strings.TrimSpace(customerID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CaseRepository) FindActiveByOrderItem(ctx context.Context, orderItemID string) (*AfterSalesModel.AfterSaleCase, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row AfterSalesModel.AfterSaleCase
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_item_id = ? AND status IN ?", tenantUUID, strings.TrimSpace(orderItemID), []string{"pending", "accepted", "reviewing", "approved"}).
		Order("created_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CaseRepository) ListByCustomer(ctx context.Context, customerID, status string, page, pageSize int) ([]AfterSalesModel.AfterSaleCase, int64, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.DB.WithContext(ctx).
		Model(&AfterSalesModel.AfterSaleCase{}).
		Where("tenant_uuid = ? AND customer_id = ?", tenantUUID, strings.TrimSpace(customerID))
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []AfterSalesModel.AfterSaleCase
	err = query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
