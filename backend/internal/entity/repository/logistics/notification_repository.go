package logistics

import (
	"context"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/gorm"
)

// NotificationTemplateRepository manages logistics notification templates.
type NotificationTemplateRepository struct {
	*Repository[LogisticsModel.NotificationTemplate]
}

func NewNotificationTemplateRepository(db *gorm.DB) *NotificationTemplateRepository {
	return &NotificationTemplateRepository{Repository: NewRepository[LogisticsModel.NotificationTemplate](db)}
}

func (r *NotificationTemplateRepository) List(ctx context.Context) ([]LogisticsModel.NotificationTemplate, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []LogisticsModel.NotificationTemplate
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ?", tenantUUID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *NotificationTemplateRepository) FindEnabledByEvent(ctx context.Context, event string) (*LogisticsModel.NotificationTemplate, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.NotificationTemplate
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND event = ? AND enabled = ?", tenantUUID, strings.TrimSpace(event), true).
		Order("updated_at DESC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *NotificationTemplateRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.NotificationTemplate, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.NotificationTemplate
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *NotificationTemplateRepository) Create(ctx context.Context, row *LogisticsModel.NotificationTemplate) error {
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

func (r *NotificationTemplateRepository) Save(ctx context.Context, row *LogisticsModel.NotificationTemplate) error {
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

// NotificationRecordRepository manages logistics notification delivery records.
type NotificationRecordRepository struct {
	*Repository[LogisticsModel.NotificationRecord]
}

func NewNotificationRecordRepository(db *gorm.DB) *NotificationRecordRepository {
	return &NotificationRecordRepository{Repository: NewRepository[LogisticsModel.NotificationRecord](db)}
}

func (r *NotificationRecordRepository) List(ctx context.Context, status string) ([]LogisticsModel.NotificationRecord, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []LogisticsModel.NotificationRecord
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *NotificationRecordRepository) GetByID(ctx context.Context, id string) (*LogisticsModel.NotificationRecord, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.NotificationRecord
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *NotificationRecordRepository) GetByIdempotencyKey(ctx context.Context, idem string) (*LogisticsModel.NotificationRecord, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row LogisticsModel.NotificationRecord
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND idempotency_key = ?", tenantUUID, strings.TrimSpace(idem)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *NotificationRecordRepository) Create(ctx context.Context, row *LogisticsModel.NotificationRecord) error {
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

func (r *NotificationRecordRepository) Save(ctx context.Context, row *LogisticsModel.NotificationRecord) error {
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
