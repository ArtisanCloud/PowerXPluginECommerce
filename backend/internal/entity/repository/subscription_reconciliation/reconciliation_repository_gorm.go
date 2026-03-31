package subscription_reconciliation

import (
	"context"
	"errors"
	"strings"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	"gorm.io/gorm"
)

type ReconciliationBatchRepository struct {
	*Repository[SubscriptionReconciliationModel.ReconciliationBatch]
}

func NewReconciliationBatchRepository(db *gorm.DB) *ReconciliationBatchRepository {
	return &ReconciliationBatchRepository{Repository: NewRepository[SubscriptionReconciliationModel.ReconciliationBatch](db)}
}

func (r *ReconciliationBatchRepository) Create(ctx context.Context, row *SubscriptionReconciliationModel.ReconciliationBatch) error {
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

func (r *ReconciliationBatchRepository) List(ctx context.Context, billingCycle, status string, limit int) ([]SubscriptionReconciliationModel.ReconciliationBatch, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(billingCycle) != "" {
		q = q.Where("billing_cycle = ?", strings.TrimSpace(billingCycle))
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []SubscriptionReconciliationModel.ReconciliationBatch
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ReconciliationBatchRepository) GetByID(ctx context.Context, id string) (*SubscriptionReconciliationModel.ReconciliationBatch, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row SubscriptionReconciliationModel.ReconciliationBatch
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReconciliationBatchRepository) FindByCycleRunType(ctx context.Context, billingCycle, runType string) (*SubscriptionReconciliationModel.ReconciliationBatch, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row SubscriptionReconciliationModel.ReconciliationBatch
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND billing_cycle = ? AND run_type = ?", tenantUUID, strings.TrimSpace(billingCycle), strings.TrimSpace(runType)).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

type ReconciliationDeltaRepository struct {
	*Repository[SubscriptionReconciliationModel.ReconciliationDelta]
}

func NewReconciliationDeltaRepository(db *gorm.DB) *ReconciliationDeltaRepository {
	return &ReconciliationDeltaRepository{Repository: NewRepository[SubscriptionReconciliationModel.ReconciliationDelta](db)}
}

func (r *ReconciliationDeltaRepository) CreateBatch(ctx context.Context, rows []SubscriptionReconciliationModel.ReconciliationDelta) error {
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

func (r *ReconciliationDeltaRepository) ListByBatchID(ctx context.Context, batchID, deltaType, riskLevel string, limit int) ([]SubscriptionReconciliationModel.ReconciliationDelta, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ? AND batch_id = ?", tenantUUID, strings.TrimSpace(batchID))
	if strings.TrimSpace(deltaType) != "" {
		q = q.Where("delta_type = ?", strings.TrimSpace(deltaType))
	}
	if strings.TrimSpace(riskLevel) != "" {
		q = q.Where("risk_level = ?", strings.TrimSpace(riskLevel))
	}
	var rows []SubscriptionReconciliationModel.ReconciliationDelta
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *ReconciliationDeltaRepository) GetByID(ctx context.Context, id string) (*SubscriptionReconciliationModel.ReconciliationDelta, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row SubscriptionReconciliationModel.ReconciliationDelta
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *ReconciliationDeltaRepository) Save(ctx context.Context, row *SubscriptionReconciliationModel.ReconciliationDelta) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}
