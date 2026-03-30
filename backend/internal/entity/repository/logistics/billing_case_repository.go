package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type BillingCaseFilter struct {
	CarrierID string
	Status    string
}

// BillingCaseRepository manages reconciliation case persistence.
type BillingCaseRepository struct {
	*Repository[LogisticsModel.BillingCase]
}

func NewBillingCaseRepository(db *gorm.DB) *BillingCaseRepository {
	return &BillingCaseRepository{Repository: NewRepository[LogisticsModel.BillingCase](db)}
}

func (r *BillingCaseRepository) List(ctx context.Context, filter BillingCaseFilter) ([]LogisticsModel.BillingCase, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		query = query.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.BillingCase
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *BillingCaseRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.BillingCase, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.BillingCase
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BillingCaseRepository) GetActiveByWaybillID(ctx context.Context, waybillID string) (*LogisticsModel.BillingCase, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.BillingCase
	err = r.DB.WithContext(ctx).
		Where(
			"tenant_uuid = ? AND waybill_id = ? AND status IN ?",
			tenantUUID,
			strings.TrimSpace(waybillID),
			[]string{"open", "confirmed", "appealed"},
		).
		Order("created_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BillingCaseRepository) Create(ctx context.Context, row *LogisticsModel.BillingCase) error {
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

func (r *BillingCaseRepository) Save(ctx context.Context, row *LogisticsModel.BillingCase) error {
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
	return r.DB.WithContext(ctx).Save(row).Error
}
