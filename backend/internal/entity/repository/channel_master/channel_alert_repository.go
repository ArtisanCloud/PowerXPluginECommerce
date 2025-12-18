package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	alertmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// ChannelAlertRepository persists alerts.
type ChannelAlertRepository struct {
	*repo.BaseRepository[alertmodel.ChannelAlert]
}

// NewChannelAlertRepository creates repo.
func NewChannelAlertRepository(db *gorm.DB) *ChannelAlertRepository {
	return &ChannelAlertRepository{BaseRepository: repo.NewBaseRepository[alertmodel.ChannelAlert](db)}
}

// Create inserts alert with tenant scope.
func (r *ChannelAlertRepository) Create(ctx context.Context, alert *alertmodel.ChannelAlert) (*alertmodel.ChannelAlert, error) {
	if alert == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(alert.TenantUUID) == "" {
		alert.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, alert)
}

// UpsertStatus ensures we only keep one open alert per type/channel.
func (r *ChannelAlertRepository) UpsertStatus(ctx context.Context, alert *alertmodel.ChannelAlert) (*alertmodel.ChannelAlert, error) {
	if alert == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(alert.TenantUUID) == "" {
		alert.TenantUUID = tenantUUID
	}
	var existing alertmodel.ChannelAlert
	tx := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ? AND type = ? AND status = ?", alert.TenantUUID, alert.ChannelID, alert.Type, "open").
		First(&existing)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, tx.Error
	}
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return r.Create(ctx, alert)
	}
	existing.Severity = alert.Severity
	existing.Title = alert.Title
	existing.Description = alert.Description
	existing.Metadata = alert.Metadata
	existing.TriggeredAt = alert.TriggeredAt
	existing.Status = "open"
	existing.ResolvedAt = nil
	if err := r.DB.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// Resolve marks open alerts of given type as resolved.
func (r *ChannelAlertRepository) Resolve(ctx context.Context, channelID, alertType string) error {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	tx := r.DB.WithContext(ctx).Model(&alertmodel.ChannelAlert{}).
		Where("tenant_uuid = ? AND channel_id = ? AND type = ? AND status = ?", tenantUUID, channelID, alertType, "open").
		Updates(map[string]any{
			"status":      "resolved",
			"resolved_at": &now,
		})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByChannel returns alerts ordered by triggered time.
func (r *ChannelAlertRepository) ListByChannel(ctx context.Context, channelID string, limit int) ([]*alertmodel.ChannelAlert, error) {
	if limit <= 0 {
		limit = 50
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var alerts []*alertmodel.ChannelAlert
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, strings.TrimSpace(channelID)).
		Order("triggered_at DESC").
		Limit(limit).
		Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// UpdateAttrs updates mutable alert fields (status, assignee, task).
func (r *ChannelAlertRepository) UpdateAttrs(ctx context.Context, alertID string, attrs map[string]any) error {
	if len(attrs) == 0 {
		return nil
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if status, ok := attrs["status"]; ok && strings.EqualFold(status.(string), "resolved") {
		now := time.Now().UTC()
		attrs["resolved_at"] = &now
	}
	return r.DB.WithContext(ctx).
		Model(&alertmodel.ChannelAlert{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, alertID).
		Updates(attrs).Error
}
