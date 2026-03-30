package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PolicyOrchestrationFlowFilter struct {
	Status string
	Limit  int
}

type PolicyOrchestrationFlowRepository struct {
	*Repository[LogisticsModel.PolicyOrchestrationFlow]
}

func NewPolicyOrchestrationFlowRepository(db *gorm.DB) *PolicyOrchestrationFlowRepository {
	return &PolicyOrchestrationFlowRepository{Repository: NewRepository[LogisticsModel.PolicyOrchestrationFlow](db)}
}

func (r *PolicyOrchestrationFlowRepository) List(ctx context.Context, filter PolicyOrchestrationFlowFilter) ([]LogisticsModel.PolicyOrchestrationFlow, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.Status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(filter.Status))
	}
	var rows []LogisticsModel.PolicyOrchestrationFlow
	err = q.Order("priority DESC, updated_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *PolicyOrchestrationFlowRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.PolicyOrchestrationFlow, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.PolicyOrchestrationFlow
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PolicyOrchestrationFlowRepository) GetByName(ctx context.Context, name string) (*LogisticsModel.PolicyOrchestrationFlow, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.PolicyOrchestrationFlow
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND name = ?", tenantUUID, strings.TrimSpace(name)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PolicyOrchestrationFlowRepository) GetByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*LogisticsModel.PolicyOrchestrationFlow, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		tx = r.DB
	}
	q := tx.WithContext(ctx)
	if tx.Dialector != nil && !strings.EqualFold(tx.Dialector.Name(), "sqlite") {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row LogisticsModel.PolicyOrchestrationFlow
	err = q.Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PolicyOrchestrationFlowRepository) Save(ctx context.Context, row *LogisticsModel.PolicyOrchestrationFlow) error {
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

type PolicyOrchestrationVersionFilter struct {
	FlowID string
	Limit  int
}

type PolicyOrchestrationVersionRepository struct {
	*Repository[LogisticsModel.PolicyOrchestrationVersion]
}

func NewPolicyOrchestrationVersionRepository(db *gorm.DB) *PolicyOrchestrationVersionRepository {
	return &PolicyOrchestrationVersionRepository{Repository: NewRepository[LogisticsModel.PolicyOrchestrationVersion](db)}
}

func (r *PolicyOrchestrationVersionRepository) List(ctx context.Context, filter PolicyOrchestrationVersionFilter) ([]LogisticsModel.PolicyOrchestrationVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filter.FlowID) != "" {
		q = q.Where("flow_id = ?", strings.TrimSpace(filter.FlowID))
	}
	var rows []LogisticsModel.PolicyOrchestrationVersion
	err = q.Order("version_no DESC, created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *PolicyOrchestrationVersionRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.PolicyOrchestrationVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.PolicyOrchestrationVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PolicyOrchestrationVersionRepository) GetByRequestKey(ctx context.Context, flowID, requestKey string) (*LogisticsModel.PolicyOrchestrationVersion, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.PolicyOrchestrationVersion
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND flow_id = ? AND request_key = ?", tenantUUID, strings.TrimSpace(flowID), strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PolicyOrchestrationVersionRepository) NextVersionNo(ctx context.Context, tx *gorm.DB, flowID string) (int, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return 0, err
	}
	if tx == nil {
		tx = r.DB
	}
	q := tx.WithContext(ctx)
	if tx.Dialector != nil && !strings.EqualFold(tx.Dialector.Name(), "sqlite") {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var latest LogisticsModel.PolicyOrchestrationVersion
	err = q.Where("tenant_uuid = ? AND flow_id = ?", tenantUUID, strings.TrimSpace(flowID)).
		Order("version_no DESC").
		First(&latest).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 1, nil
		}
		return 0, err
	}
	return latest.VersionNo + 1, nil
}

func (r *PolicyOrchestrationVersionRepository) Save(ctx context.Context, row *LogisticsModel.PolicyOrchestrationVersion) error {
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
