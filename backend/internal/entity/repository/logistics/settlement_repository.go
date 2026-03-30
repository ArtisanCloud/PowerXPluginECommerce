package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type SettlementBatchRepository struct {
	*Repository[LogisticsModel.SettlementBatch]
}

func NewSettlementBatchRepository(db *gorm.DB) *SettlementBatchRepository {
	return &SettlementBatchRepository{Repository: NewRepository[LogisticsModel.SettlementBatch](db)}
}

func (r *SettlementBatchRepository) List(ctx context.Context, carrierID, status string, limit int) ([]LogisticsModel.SettlementBatch, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(carrierID) != "" {
		q = q.Where("carrier_id = ?", strings.TrimSpace(carrierID))
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.SettlementBatch
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *SettlementBatchRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.SettlementBatch, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.SettlementBatch
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SettlementBatchRepository) Create(ctx context.Context, row *LogisticsModel.SettlementBatch) error {
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

func (r *SettlementBatchRepository) Save(ctx context.Context, row *LogisticsModel.SettlementBatch) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

type SettlementDiffRepository struct {
	*Repository[LogisticsModel.SettlementDiff]
}

func NewSettlementDiffRepository(db *gorm.DB) *SettlementDiffRepository {
	return &SettlementDiffRepository{Repository: NewRepository[LogisticsModel.SettlementDiff](db)}
}

func (r *SettlementDiffRepository) List(ctx context.Context, batchID, status string, limit int) ([]LogisticsModel.SettlementDiff, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(batchID) != "" {
		q = q.Where("batch_id = ?", strings.TrimSpace(batchID))
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.SettlementDiff
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *SettlementDiffRepository) ListByBatchID(ctx context.Context, batchID string) ([]LogisticsModel.SettlementDiff, error) {
	return r.List(ctx, batchID, "", 500)
}

func (r *SettlementDiffRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.SettlementDiff, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.SettlementDiff
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SettlementDiffRepository) CreateBatch(ctx context.Context, rows []LogisticsModel.SettlementDiff) error {
	if len(rows) == 0 {
		return nil
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.TrimSpace(rows[i].TenantUUID) == "" {
			rows[i].TenantUUID = tenantUUID
		}
	}
	return r.DB.WithContext(ctx).Create(&rows).Error
}

func (r *SettlementDiffRepository) Save(ctx context.Context, row *LogisticsModel.SettlementDiff) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
