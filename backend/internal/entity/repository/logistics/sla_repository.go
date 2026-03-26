package logistics

import (
	"context"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type SLAFilter struct {
	CarrierID string
	From      *time.Time
	To        *time.Time
}

type SLAWaybillMetric struct {
	WaybillID    string    `json:"waybill_id"`
	CarrierID    string    `json:"carrier_id"`
	CarrierName  string    `json:"carrier_name"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status"`
	HasException bool      `json:"has_exception"`
}

// SLARepository provides aggregated projections for logistics SLA dashboards.
type SLARepository struct {
	waybillRepo *WaybillRepository
}

func NewSLARepository(waybillRepo *WaybillRepository) *SLARepository {
	return &SLARepository{waybillRepo: waybillRepo}
}

func (r *SLARepository) ListWaybillMetrics(ctx context.Context, filter SLAFilter) ([]SLAWaybillMetric, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.waybillRepo.DB.WithContext(ctx).
		Table(LogisticsModel.Waybill{}.TableName()+" AS wb").
		Select(`
			wb.id AS waybill_id,
			wb.carrier_id AS carrier_id,
			COALESCE(c.name, wb.carrier_id) AS carrier_name,
			wb.created_at AS created_at,
			wb.status AS status,
			CASE
				WHEN wb.status IN ('delay','exception') THEN 1
				WHEN EXISTS (
					SELECT 1
					FROM `+LogisticsModel.TrackingEvent{}.TableName()+` AS te2
					WHERE te2.tenant_uuid = wb.tenant_uuid
					  AND te2.waybill_id = wb.id
					  AND te2.status IN ('delay','exception')
				) THEN 1
				ELSE 0
			END AS has_exception
		`).
		Joins("LEFT JOIN "+LogisticsModel.Carrier{}.TableName()+" AS c ON c.id = wb.carrier_id").
		Where("wb.tenant_uuid = ?", tenantUUID).
		Where("wb.deleted_at IS NULL")
	db = applySLAFilter(db, filter)

	var rows []SLAWaybillMetric
	if err := db.Order("wb.created_at DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func applySLAFilter(db *gorm.DB, filter SLAFilter) *gorm.DB {
	if strings.TrimSpace(filter.CarrierID) != "" {
		db = db.Where("wb.carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if filter.From != nil {
		db = db.Where("wb.created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		db = db.Where("wb.created_at <= ?", *filter.To)
	}
	return db
}
