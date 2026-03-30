package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// RoutingRuleRepository manages routing rule persistence.
type RoutingRuleRepository struct {
	*Repository[LogisticsModel.RoutingRule]
}

func NewRoutingRuleRepository(db *gorm.DB) *RoutingRuleRepository {
	return &RoutingRuleRepository{Repository: NewRepository[LogisticsModel.RoutingRule](db)}
}

func (r *RoutingRuleRepository) List(ctx context.Context, enabledOnly bool) ([]LogisticsModel.RoutingRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	var rows []LogisticsModel.RoutingRule
	err = query.Order("priority DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *RoutingRuleRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.RoutingRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RoutingRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RoutingRuleRepository) Create(ctx context.Context, row *LogisticsModel.RoutingRule) error {
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

func (r *RoutingRuleRepository) Save(ctx context.Context, row *LogisticsModel.RoutingRule) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *RoutingRuleRepository) DeleteByID(ctx context.Context, id string) error {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	return r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		Delete(&LogisticsModel.RoutingRule{}).Error
}
