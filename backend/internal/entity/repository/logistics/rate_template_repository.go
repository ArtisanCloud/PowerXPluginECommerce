package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// RateTemplateRepository manages freight template persistence.
type RateTemplateRepository struct {
	*Repository[LogisticsModel.RateTemplate]
}

func NewRateTemplateRepository(db *gorm.DB) *RateTemplateRepository {
	return &RateTemplateRepository{Repository: NewRepository[LogisticsModel.RateTemplate](db)}
}

func (r *RateTemplateRepository) List(ctx context.Context) ([]LogisticsModel.RateTemplate, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.RateTemplate
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *RateTemplateRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.RateTemplate, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RateTemplate
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RateTemplateRepository) Create(ctx context.Context, tpl *LogisticsModel.RateTemplate) error {
	if tpl == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(tpl.TenantUUID) == "" {
		tpl.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(tpl).Error
}

func (r *RateTemplateRepository) Save(ctx context.Context, tpl *LogisticsModel.RateTemplate) error {
	if tpl == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(tpl).Error
}

// RateZoneRepository handles zone-level freight rows.
type RateZoneRepository struct {
	*Repository[LogisticsModel.RateZone]
}

func NewRateZoneRepository(db *gorm.DB) *RateZoneRepository {
	return &RateZoneRepository{Repository: NewRepository[LogisticsModel.RateZone](db)}
}

func (r *RateZoneRepository) ListByTemplateID(ctx context.Context, templateID string) ([]LogisticsModel.RateZone, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.RateZone
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND template_id = ?", tenantUUID, strings.TrimSpace(templateID)).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}
