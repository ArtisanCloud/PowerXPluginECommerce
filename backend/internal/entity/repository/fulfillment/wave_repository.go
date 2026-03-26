package fulfillment

import (
	"context"
	"strings"

	FulfillmentModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/fulfillment"
	"gorm.io/gorm"
)

// WaveRepository manages fulfillment wave persistence.
type WaveRepository struct {
	*Repository[FulfillmentModel.Wave]
}

func NewWaveRepository(db *gorm.DB) *WaveRepository {
	return &WaveRepository{Repository: NewRepository[FulfillmentModel.Wave](db)}
}

func (r *WaveRepository) List(ctx context.Context, status string) ([]FulfillmentModel.Wave, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var rows []FulfillmentModel.Wave
	err = query.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *WaveRepository) GetByID(ctx context.Context, id string) (*FulfillmentModel.Wave, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.Wave
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, strings.TrimSpace(id)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaveRepository) Create(ctx context.Context, wave *FulfillmentModel.Wave) error {
	if wave == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(wave.TenantUUID) == "" {
		wave.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(wave).Error
}

func (r *WaveRepository) Save(ctx context.Context, wave *FulfillmentModel.Wave) error {
	if wave == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(wave.TenantUUID) == "" {
		wave.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Save(wave).Error
}

// WaveTaskLinkRepository manages task links in a wave.
type WaveTaskLinkRepository struct {
	*Repository[FulfillmentModel.WaveTaskLink]
}

func NewWaveTaskLinkRepository(db *gorm.DB) *WaveTaskLinkRepository {
	return &WaveTaskLinkRepository{Repository: NewRepository[FulfillmentModel.WaveTaskLink](db)}
}

func (r *WaveTaskLinkRepository) ListByWaveID(ctx context.Context, waveID string) ([]FulfillmentModel.WaveTaskLink, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var rows []FulfillmentModel.WaveTaskLink
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND wave_id = ?", tenantUUID, strings.TrimSpace(waveID)).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *WaveTaskLinkRepository) GetByWaveAndTask(ctx context.Context, waveID, taskID string) (*FulfillmentModel.WaveTaskLink, error) {
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var row FulfillmentModel.WaveTaskLink
	err = r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND wave_id = ? AND task_id = ?", tenantUUID, strings.TrimSpace(waveID), strings.TrimSpace(taskID)).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *WaveTaskLinkRepository) Create(ctx context.Context, link *FulfillmentModel.WaveTaskLink) error {
	if link == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(link.TenantUUID) == "" {
		link.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(link).Error
}

func (r *WaveTaskLinkRepository) Save(ctx context.Context, link *FulfillmentModel.WaveTaskLink) error {
	if link == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(link.TenantUUID) == "" {
		link.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Save(link).Error
}
