package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type OpsAutomationPolicyFilter struct {
	CarrierID string
	Enabled   *bool
	Limit     int
}

type OpsAutomationPolicyRepository struct {
	*Repository[LogisticsModel.OpsAutomationPolicy]
}

func NewOpsAutomationPolicyRepository(db *gorm.DB) *OpsAutomationPolicyRepository {
	return &OpsAutomationPolicyRepository{Repository: NewRepository[LogisticsModel.OpsAutomationPolicy](db)}
}

func (r *OpsAutomationPolicyRepository) Save(ctx context.Context, row *LogisticsModel.OpsAutomationPolicy) error {
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

func (r *OpsAutomationPolicyRepository) List(ctx context.Context, filter OpsAutomationPolicyFilter) ([]LogisticsModel.OpsAutomationPolicy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.CarrierID); v != "" {
		q = q.Where("(carrier_id = ? OR carrier_id = '')", v)
	}
	if filter.Enabled != nil {
		q = q.Where("enabled = ?", *filter.Enabled)
	}
	var rows []LogisticsModel.OpsAutomationPolicy
	err = q.Order("updated_at DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *OpsAutomationPolicyRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.OpsAutomationPolicy, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.OpsAutomationPolicy
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type OpsAutomationRunFilter struct {
	PolicyID  string
	CarrierID string
	Status    string
	Limit     int
}

type OpsAutomationRunRepository struct {
	*Repository[LogisticsModel.OpsAutomationRun]
}

func NewOpsAutomationRunRepository(db *gorm.DB) *OpsAutomationRunRepository {
	return &OpsAutomationRunRepository{Repository: NewRepository[LogisticsModel.OpsAutomationRun](db)}
}

func (r *OpsAutomationRunRepository) Save(ctx context.Context, row *LogisticsModel.OpsAutomationRun) error {
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

func (r *OpsAutomationRunRepository) List(ctx context.Context, filter OpsAutomationRunFilter) ([]LogisticsModel.OpsAutomationRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.PolicyID); v != "" {
		q = q.Where("policy_id = ?", v)
	}
	if v := strings.TrimSpace(filter.CarrierID); v != "" {
		q = q.Where("carrier_id = ?", v)
	}
	if v := strings.TrimSpace(filter.Status); v != "" {
		q = q.Where("status = ?", v)
	}
	var rows []LogisticsModel.OpsAutomationRun
	err = q.Order("created_at DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *OpsAutomationRunRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.OpsAutomationRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.OpsAutomationRun
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
