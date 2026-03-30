package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// RiskRuleRepository manages logistics risk-rule persistence.
type RiskRuleRepository struct {
	*Repository[LogisticsModel.RiskRule]
}

func NewRiskRuleRepository(db *gorm.DB) *RiskRuleRepository {
	return &RiskRuleRepository{Repository: NewRepository[LogisticsModel.RiskRule](db)}
}

func (r *RiskRuleRepository) List(ctx context.Context, enabledOnly bool) ([]LogisticsModel.RiskRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	var rows []LogisticsModel.RiskRule
	err = query.Order("priority DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *RiskRuleRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.RiskRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RiskRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RiskRuleRepository) Create(ctx context.Context, row *LogisticsModel.RiskRule) error {
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

func (r *RiskRuleRepository) Save(ctx context.Context, row *LogisticsModel.RiskRule) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
