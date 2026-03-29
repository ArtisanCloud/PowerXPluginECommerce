package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type CustomsRulePackRepository struct {
	*Repository[LogisticsModel.CustomsRulePack]
}

func NewCustomsRulePackRepository(db *gorm.DB) *CustomsRulePackRepository {
	return &CustomsRulePackRepository{Repository: NewRepository[LogisticsModel.CustomsRulePack](db)}
}

func (r *CustomsRulePackRepository) Save(ctx context.Context, row *LogisticsModel.CustomsRulePack) error {
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

func (r *CustomsRulePackRepository) List(ctx context.Context, countryCode, status string, limit int) ([]LogisticsModel.CustomsRulePack, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(countryCode) != "" {
		q = q.Where("country_code = ?", strings.ToUpper(strings.TrimSpace(countryCode)))
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.CustomsRulePack
	err = q.Order("updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *CustomsRulePackRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.CustomsRulePack, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CustomsRulePack
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type CustomsRuleVersionRepository struct {
	*Repository[LogisticsModel.CustomsRuleVersion]
}

func NewCustomsRuleVersionRepository(db *gorm.DB) *CustomsRuleVersionRepository {
	return &CustomsRuleVersionRepository{Repository: NewRepository[LogisticsModel.CustomsRuleVersion](db)}
}

func (r *CustomsRuleVersionRepository) Save(ctx context.Context, row *LogisticsModel.CustomsRuleVersion) error {
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

func (r *CustomsRuleVersionRepository) ListByPackID(ctx context.Context, packID string, limit int) ([]LogisticsModel.CustomsRuleVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []LogisticsModel.CustomsRuleVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND pack_id = ?", tenantUUID, strings.TrimSpace(packID)).
		Order("version_no DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *CustomsRuleVersionRepository) GetLatestByPackID(ctx context.Context, packID string) (*LogisticsModel.CustomsRuleVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CustomsRuleVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND pack_id = ?", tenantUUID, strings.TrimSpace(packID)).
		Order("version_no DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CustomsRuleVersionRepository) GetLatestPublishedByPackID(ctx context.Context, packID string) (*LogisticsModel.CustomsRuleVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.CustomsRuleVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND pack_id = ? AND status = ?", tenantUUID, strings.TrimSpace(packID), "published").
		Order("version_no DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
