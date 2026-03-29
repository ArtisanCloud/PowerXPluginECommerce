package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type SLOGuardPolicyFilter struct {
	CarrierID string
	Enabled   *bool
	Limit     int
}

type SLOGuardPolicyRepository struct {
	*Repository[LogisticsModel.SLOGuardPolicy]
}

func NewSLOGuardPolicyRepository(db *gorm.DB) *SLOGuardPolicyRepository {
	return &SLOGuardPolicyRepository{Repository: NewRepository[LogisticsModel.SLOGuardPolicy](db)}
}

func (r *SLOGuardPolicyRepository) Save(ctx context.Context, row *LogisticsModel.SLOGuardPolicy) error {
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

func (r *SLOGuardPolicyRepository) List(ctx context.Context, filter SLOGuardPolicyFilter) ([]LogisticsModel.SLOGuardPolicy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.CarrierID) != "" {
		q = q.Where("(carrier_id = ? OR carrier_id = '')", strings.TrimSpace(filter.CarrierID))
	}
	if filter.Enabled != nil {
		q = q.Where("enabled = ?", *filter.Enabled)
	}
	var rows []LogisticsModel.SLOGuardPolicy
	err = q.Order("updated_at DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *SLOGuardPolicyRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.SLOGuardPolicy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.SLOGuardPolicy
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type SLOGuardStateFilter struct {
	PolicyID  string
	CarrierID string
	Status    string
	Limit     int
}

type SLOGuardStateRepository struct {
	*Repository[LogisticsModel.SLOGuardState]
}

func NewSLOGuardStateRepository(db *gorm.DB) *SLOGuardStateRepository {
	return &SLOGuardStateRepository{Repository: NewRepository[LogisticsModel.SLOGuardState](db)}
}

func (r *SLOGuardStateRepository) Save(ctx context.Context, row *LogisticsModel.SLOGuardState) error {
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

func (r *SLOGuardStateRepository) List(ctx context.Context, filter SLOGuardStateFilter) ([]LogisticsModel.SLOGuardState, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.PolicyID) != "" {
		q = q.Where("policy_id = ?", strings.TrimSpace(filter.PolicyID))
	}
	if strings.TrimSpace(filter.CarrierID) != "" {
		q = q.Where("carrier_id = ?", strings.TrimSpace(filter.CarrierID))
	}
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.SLOGuardState
	err = q.Order("created_at DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *SLOGuardStateRepository) GetLatestByPolicyID(ctx context.Context, policyID string) (*LogisticsModel.SLOGuardState, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.SLOGuardState
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND policy_id = ?", tenantUUID, strings.TrimSpace(policyID)).
		Order("created_at DESC, updated_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
