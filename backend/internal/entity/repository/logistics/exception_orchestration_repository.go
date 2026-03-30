package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type ExceptionOrchestrationRuleRepository struct {
	*Repository[LogisticsModel.ExceptionOrchestrationRule]
}

func NewExceptionOrchestrationRuleRepository(db *gorm.DB) *ExceptionOrchestrationRuleRepository {
	return &ExceptionOrchestrationRuleRepository{
		Repository: NewRepository[LogisticsModel.ExceptionOrchestrationRule](db),
	}
}

func (r *ExceptionOrchestrationRuleRepository) Create(ctx context.Context, row *LogisticsModel.ExceptionOrchestrationRule) error {
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

func (r *ExceptionOrchestrationRuleRepository) Save(ctx context.Context, row *LogisticsModel.ExceptionOrchestrationRule) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	return r.DB.WithContext(ctx).Save(row).Error
}

func (r *ExceptionOrchestrationRuleRepository) List(ctx context.Context, enabled *bool) ([]LogisticsModel.ExceptionOrchestrationRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabled != nil {
		db = db.Where("enabled = ?", *enabled)
	}
	var rows []LogisticsModel.ExceptionOrchestrationRule
	err = db.Order("priority DESC, created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *ExceptionOrchestrationRuleRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.ExceptionOrchestrationRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.ExceptionOrchestrationRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

type ExceptionOrchestrationRunRepository struct {
	*Repository[LogisticsModel.ExceptionOrchestrationRun]
}

func NewExceptionOrchestrationRunRepository(db *gorm.DB) *ExceptionOrchestrationRunRepository {
	return &ExceptionOrchestrationRunRepository{
		Repository: NewRepository[LogisticsModel.ExceptionOrchestrationRun](db),
	}
}

func (r *ExceptionOrchestrationRunRepository) Create(ctx context.Context, row *LogisticsModel.ExceptionOrchestrationRun) error {
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

func (r *ExceptionOrchestrationRunRepository) List(ctx context.Context, waybillNo string, limit int) ([]LogisticsModel.ExceptionOrchestrationRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	db := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(waybillNo) != "" {
		db = db.Where("waybill_no = ?", strings.TrimSpace(waybillNo))
	}
	var rows []LogisticsModel.ExceptionOrchestrationRun
	err = db.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
