package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type ReconciliationBatchRepository struct {
	*Repository[LogisticsModel.ReconciliationBatch]
}

func NewReconciliationBatchRepository(db *gorm.DB) *ReconciliationBatchRepository {
	return &ReconciliationBatchRepository{Repository: NewRepository[LogisticsModel.ReconciliationBatch](db)}
}

func (r *ReconciliationBatchRepository) List(ctx context.Context, carrierID, status string, limit int) ([]LogisticsModel.ReconciliationBatch, error) {
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
	var rows []LogisticsModel.ReconciliationBatch
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ReconciliationBatchRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.ReconciliationBatch, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ReconciliationBatch
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReconciliationBatchRepository) Create(ctx context.Context, row *LogisticsModel.ReconciliationBatch) error {
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

func (r *ReconciliationBatchRepository) Save(ctx context.Context, row *LogisticsModel.ReconciliationBatch) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

type ReconciliationRecordRepository struct {
	*Repository[LogisticsModel.ReconciliationRecord]
}

func NewReconciliationRecordRepository(db *gorm.DB) *ReconciliationRecordRepository {
	return &ReconciliationRecordRepository{Repository: NewRepository[LogisticsModel.ReconciliationRecord](db)}
}

func (r *ReconciliationRecordRepository) List(ctx context.Context, batchID, status string, limit int) ([]LogisticsModel.ReconciliationRecord, error) {
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
	var rows []LogisticsModel.ReconciliationRecord
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ReconciliationRecordRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.ReconciliationRecord, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ReconciliationRecord
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReconciliationRecordRepository) CreateBatch(ctx context.Context, rows []LogisticsModel.ReconciliationRecord) error {
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

func (r *ReconciliationRecordRepository) Save(ctx context.Context, row *LogisticsModel.ReconciliationRecord) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

type ReconciliationCaseRepository struct {
	*Repository[LogisticsModel.ReconciliationCase]
}

func NewReconciliationCaseRepository(db *gorm.DB) *ReconciliationCaseRepository {
	return &ReconciliationCaseRepository{Repository: NewRepository[LogisticsModel.ReconciliationCase](db)}
}

func (r *ReconciliationCaseRepository) List(ctx context.Context, batchID, status string, limit int) ([]LogisticsModel.ReconciliationCase, error) {
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
	var rows []LogisticsModel.ReconciliationCase
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ReconciliationCaseRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.ReconciliationCase, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ReconciliationCase
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReconciliationCaseRepository) CreateBatch(ctx context.Context, rows []LogisticsModel.ReconciliationCase) error {
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

func (r *ReconciliationCaseRepository) Save(ctx context.Context, row *LogisticsModel.ReconciliationCase) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
