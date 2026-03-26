package reverse

import (
	"context"
	"strings"

	ReverseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
	"gorm.io/gorm"
)

// InspectionRuleRepository manages reverse inspection-rule persistence.
type InspectionRuleRepository struct {
	*Repository[ReverseModel.InspectionRule]
}

func NewInspectionRuleRepository(db *gorm.DB) *InspectionRuleRepository {
	return &InspectionRuleRepository{Repository: NewRepository[ReverseModel.InspectionRule](db)}
}

func (r *InspectionRuleRepository) ListEnabled(ctx context.Context) ([]ReverseModel.InspectionRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []ReverseModel.InspectionRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND enabled = ?", tenantUUID, true).
		Order("priority ASC, created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *InspectionRuleRepository) ListAll(ctx context.Context) ([]ReverseModel.InspectionRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []ReverseModel.InspectionRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("priority ASC, created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *InspectionRuleRepository) Create(ctx context.Context, row *ReverseModel.InspectionRule) error {
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

// WaybillInspectionRepository manages reverse-waybill inspection outcomes.
type WaybillInspectionRepository struct {
	*Repository[ReverseModel.WaybillInspection]
}

func NewWaybillInspectionRepository(db *gorm.DB) *WaybillInspectionRepository {
	return &WaybillInspectionRepository{Repository: NewRepository[ReverseModel.WaybillInspection](db)}
}

func (r *WaybillInspectionRepository) GetByWaybillID(ctx context.Context, waybillID string) (*ReverseModel.WaybillInspection, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row ReverseModel.WaybillInspection
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND waybill_id = ?", tenantUUID, strings.TrimSpace(waybillID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaybillInspectionRepository) Create(ctx context.Context, row *ReverseModel.WaybillInspection) error {
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
