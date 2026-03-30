package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type QualityAuditReportFilter struct {
	CarrierID       string
	WarehouseID     string
	DestinationZone string
	Status          string
	Limit           int
}

type QualityAuditReportRepository struct {
	*Repository[LogisticsModel.QualityAuditReport]
}

func NewQualityAuditReportRepository(db *gorm.DB) *QualityAuditReportRepository {
	return &QualityAuditReportRepository{Repository: NewRepository[LogisticsModel.QualityAuditReport](db)}
}

func (r *QualityAuditReportRepository) List(ctx context.Context, filter QualityAuditReportFilter) ([]LogisticsModel.QualityAuditReport, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		q = q.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.WarehouseID) != "" {
		q = q.Where("warehouse_id = ?", strings.TrimSpace(filter.WarehouseID))
	}
	if strings.TrimSpace(filter.DestinationZone) != "" {
		q = q.Where("destination_zone = ?", strings.TrimSpace(filter.DestinationZone))
	}
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.QualityAuditReport
	err = q.Order("report_period_to DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *QualityAuditReportRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.QualityAuditReport, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.QualityAuditReport
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *QualityAuditReportRepository) Save(ctx context.Context, row *LogisticsModel.QualityAuditReport) error {
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
