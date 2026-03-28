package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type KPISnapshotQuery struct {
	WindowHours   int
	DimensionType string
	DimensionKey  string
	Limit         int
}

type KPISnapshotRepository struct {
	*Repository[LogisticsModel.KPISnapshot]
}

func NewKPISnapshotRepository(db *gorm.DB) *KPISnapshotRepository {
	return &KPISnapshotRepository{Repository: NewRepository[LogisticsModel.KPISnapshot](db)}
}

func (r *KPISnapshotRepository) Create(ctx context.Context, row *LogisticsModel.KPISnapshot) error {
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

func (r *KPISnapshotRepository) List(ctx context.Context, query KPISnapshotQuery) ([]LogisticsModel.KPISnapshot, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := query.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if query.WindowHours > 0 {
		q = q.Where("window_hours = ?", query.WindowHours)
	}
	if strings.TrimSpace(query.DimensionType) != "" {
		q = q.Where("dimension_type = ?", strings.TrimSpace(query.DimensionType))
	}
	if strings.TrimSpace(query.DimensionKey) != "" {
		q = q.Where("dimension_key = ?", strings.TrimSpace(query.DimensionKey))
	}
	var rows []LogisticsModel.KPISnapshot
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
