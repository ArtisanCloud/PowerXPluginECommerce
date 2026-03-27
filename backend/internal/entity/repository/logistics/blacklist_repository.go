package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// BlacklistRepository manages logistics risk blacklist entries.
type BlacklistRepository struct {
	*Repository[LogisticsModel.BlacklistEntry]
}

func NewBlacklistRepository(db *gorm.DB) *BlacklistRepository {
	return &BlacklistRepository{Repository: NewRepository[LogisticsModel.BlacklistEntry](db)}
}

func (r *BlacklistRepository) List(ctx context.Context, status string) ([]LogisticsModel.BlacklistEntry, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.BlacklistEntry
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *BlacklistRepository) ListActive(ctx context.Context, now time.Time) ([]LogisticsModel.BlacklistEntry, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.BlacklistEntry
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND status = ?", tenantUUID, "active").
		Where("expires_at IS NULL OR expires_at >= ?", now.UTC()).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *BlacklistRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.BlacklistEntry, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.BlacklistEntry
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BlacklistRepository) Create(ctx context.Context, row *LogisticsModel.BlacklistEntry) error {
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

func (r *BlacklistRepository) Save(ctx context.Context, row *LogisticsModel.BlacklistEntry) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
