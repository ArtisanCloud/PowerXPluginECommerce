package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type ComplianceKBVersionFilter struct {
	CountryCode string
	Status      string
	Limit       int
}

type ComplianceKBVersionRepository struct {
	*Repository[LogisticsModel.ComplianceKBVersion]
}

func NewComplianceKBVersionRepository(db *gorm.DB) *ComplianceKBVersionRepository {
	return &ComplianceKBVersionRepository{Repository: NewRepository[LogisticsModel.ComplianceKBVersion](db)}
}

func (r *ComplianceKBVersionRepository) Save(ctx context.Context, row *LogisticsModel.ComplianceKBVersion) error {
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

func (r *ComplianceKBVersionRepository) List(ctx context.Context, filter ComplianceKBVersionFilter) ([]LogisticsModel.ComplianceKBVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 300 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.ToUpper(strings.TrimSpace(filter.CountryCode)); v != "" {
		q = q.Where("country_code = ?", v)
	}
	if v := strings.TrimSpace(filter.Status); v != "" {
		q = q.Where("status = ?", v)
	}
	var rows []LogisticsModel.ComplianceKBVersion
	err = q.Order("country_code ASC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ComplianceKBVersionRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.ComplianceKBVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ComplianceKBVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ComplianceKBVersionRepository) GetLatestByCountry(ctx context.Context, countryCode string) (*LogisticsModel.ComplianceKBVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ComplianceKBVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND country_code = ?", tenantUUID, strings.ToUpper(strings.TrimSpace(countryCode))).
		Order("updated_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ComplianceKBVersionRepository) ListByCountry(ctx context.Context, countryCode string, limit int) ([]LogisticsModel.ComplianceKBVersion, error) {
	return r.List(ctx, ComplianceKBVersionFilter{
		CountryCode: countryCode,
		Limit:       limit,
	})
}
