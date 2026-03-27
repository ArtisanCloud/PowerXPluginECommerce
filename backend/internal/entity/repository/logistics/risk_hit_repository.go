package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// RiskHitRepository manages persisted risk-hit records.
type RiskHitRepository struct {
	*Repository[LogisticsModel.RiskHit]
}

func NewRiskHitRepository(db *gorm.DB) *RiskHitRepository {
	return &RiskHitRepository{Repository: NewRepository[LogisticsModel.RiskHit](db)}
}

func (r *RiskHitRepository) List(ctx context.Context, waybillID, status string) ([]LogisticsModel.RiskHit, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(waybillID) != "" {
		query = query.Where("waybill_id = ?", strings.TrimSpace(waybillID))
	}
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.RiskHit
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *RiskHitRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.RiskHit, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.RiskHit
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RiskHitRepository) ListReleasedByFingerprint(ctx context.Context, fingerprint string) ([]LogisticsModel.RiskHit, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.RiskHit
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND fingerprint = ? AND status = ?", tenantUUID, strings.TrimSpace(fingerprint), "released").
		Order("released_at DESC, updated_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *RiskHitRepository) Create(ctx context.Context, row *LogisticsModel.RiskHit) error {
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

func (r *RiskHitRepository) Save(ctx context.Context, row *LogisticsModel.RiskHit) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
