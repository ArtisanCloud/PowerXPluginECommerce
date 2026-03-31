package subscription_reconciliation

import (
	"context"
	"errors"
	"strings"
	"time"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	"gorm.io/gorm"
)

type DeltaTaskRepository struct {
	*Repository[SubscriptionReconciliationModel.DeltaTask]
}

func NewDeltaTaskRepository(db *gorm.DB) *DeltaTaskRepository {
	return &DeltaTaskRepository{Repository: NewRepository[SubscriptionReconciliationModel.DeltaTask](db)}
}

func (r *DeltaTaskRepository) Create(ctx context.Context, row *SubscriptionReconciliationModel.DeltaTask) error {
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

func (r *DeltaTaskRepository) GetByID(ctx context.Context, id string) (*SubscriptionReconciliationModel.DeltaTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row SubscriptionReconciliationModel.DeltaTask
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DeltaTaskRepository) FindOpenByFingerprint(ctx context.Context, fingerprint string) (*SubscriptionReconciliationModel.DeltaTask, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row SubscriptionReconciliationModel.DeltaTask
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND delta_fingerprint = ? AND status <> ?", tenantUUID, strings.TrimSpace(fingerprint), "closed").
		Order("created_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *DeltaTaskRepository) Close(ctx context.Context, id, resolution, note string, closedAt time.Time) (*SubscriptionReconciliationModel.DeltaTask, error) {
	row, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	row.Status = "closed"
	row.Resolution = strings.TrimSpace(resolution)
	row.ResolutionNote = strings.TrimSpace(note)
	row.ClosedAt = &closedAt
	if err := r.DB.WithContext(ctx).Save(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

type RenewalGovernancePolicyRepository struct {
	*Repository[SubscriptionReconciliationModel.RenewalGovernancePolicy]
}

func NewRenewalGovernancePolicyRepository(db *gorm.DB) *RenewalGovernancePolicyRepository {
	return &RenewalGovernancePolicyRepository{Repository: NewRepository[SubscriptionReconciliationModel.RenewalGovernancePolicy](db)}
}

func (r *RenewalGovernancePolicyRepository) GetLatestEnabled(ctx context.Context) (*SubscriptionReconciliationModel.RenewalGovernancePolicy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row SubscriptionReconciliationModel.RenewalGovernancePolicy
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND enabled = ?", tenantUUID, true).
		Order("version DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

type RenewalExecutionLogRepository struct {
	*Repository[SubscriptionReconciliationModel.RenewalExecutionLog]
}

func NewRenewalExecutionLogRepository(db *gorm.DB) *RenewalExecutionLogRepository {
	return &RenewalExecutionLogRepository{Repository: NewRepository[SubscriptionReconciliationModel.RenewalExecutionLog](db)}
}

func (r *RenewalExecutionLogRepository) CreateBatch(ctx context.Context, rows []SubscriptionReconciliationModel.RenewalExecutionLog) error {
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

func (r *RenewalExecutionLogRepository) CountFailedRetries(ctx context.Context, subscriptionRef string, since time.Time) (int64, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return 0, err
	}
	var count int64
	q := r.DB.WithContext(ctx).
		Model(&SubscriptionReconciliationModel.RenewalExecutionLog{}).
		Where("tenant_uuid = ? AND subscription_ref = ? AND action_type = ? AND result = ?", tenantUUID, strings.TrimSpace(subscriptionRef), "retry", "failed")
	if !since.IsZero() {
		q = q.Where("executed_at >= ?", since)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
