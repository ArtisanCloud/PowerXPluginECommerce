package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

type LastmileRecoveryRuleRepository struct {
	*Repository[LogisticsModel.LastmileRecoveryRule]
}

func NewLastmileRecoveryRuleRepository(db *gorm.DB) *LastmileRecoveryRuleRepository {
	return &LastmileRecoveryRuleRepository{Repository: NewRepository[LogisticsModel.LastmileRecoveryRule](db)}
}

func (r *LastmileRecoveryRuleRepository) List(ctx context.Context, enabled *bool) ([]LogisticsModel.LastmileRecoveryRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if enabled != nil {
		q = q.Where("enabled = ?", *enabled)
	}
	var rows []LogisticsModel.LastmileRecoveryRule
	err = q.Order("priority ASC, updated_at DESC").Find(&rows).Error
	return rows, err
}

func (r *LastmileRecoveryRuleRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.LastmileRecoveryRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.LastmileRecoveryRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LastmileRecoveryRuleRepository) GetByName(ctx context.Context, name string) (*LogisticsModel.LastmileRecoveryRule, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.LastmileRecoveryRule
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND name = ?", tenantUUID, strings.TrimSpace(name)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LastmileRecoveryRuleRepository) Save(ctx context.Context, row *LogisticsModel.LastmileRecoveryRule) error {
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

type LastmileRecoveryRunRepository struct {
	*Repository[LogisticsModel.LastmileRecoveryRun]
}

func NewLastmileRecoveryRunRepository(db *gorm.DB) *LastmileRecoveryRunRepository {
	return &LastmileRecoveryRunRepository{Repository: NewRepository[LogisticsModel.LastmileRecoveryRun](db)}
}

func (r *LastmileRecoveryRunRepository) GetByRequestKey(ctx context.Context, requestKey string) (*LogisticsModel.LastmileRecoveryRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.LastmileRecoveryRun
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND request_key = ?", tenantUUID, strings.TrimSpace(requestKey)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LastmileRecoveryRunRepository) List(ctx context.Context, waybillNo, status string, limit int) ([]LogisticsModel.LastmileRecoveryRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(waybillNo) != "" {
		q = q.Where("waybill_no = ?", strings.TrimSpace(waybillNo))
	}
	if strings.TrimSpace(status) != "" {
		q = q.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.LastmileRecoveryRun
	err = q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *LastmileRecoveryRunRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.LastmileRecoveryRun, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.LastmileRecoveryRun
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *LastmileRecoveryRunRepository) Save(ctx context.Context, row *LogisticsModel.LastmileRecoveryRun) error {
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
