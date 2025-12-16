package repository

import (
	"context"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// ChannelTaskLinkRepository manages task associations.
type ChannelTaskLinkRepository struct {
	*repo.BaseRepository[channelmodel.ChannelTaskLink]
}

// NewChannelTaskLinkRepository builds repository.
func NewChannelTaskLinkRepository(db *gorm.DB) *ChannelTaskLinkRepository {
	return &ChannelTaskLinkRepository{BaseRepository: repo.NewBaseRepository[channelmodel.ChannelTaskLink](db)}
}

// Create stores a link scoped by tenant.
func (r *ChannelTaskLinkRepository) Create(ctx context.Context, link *channelmodel.ChannelTaskLink) (*channelmodel.ChannelTaskLink, error) {
	if link == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(link.TenantUUID) == "" {
		link.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, link)
}

// ListByChannel returns task associations.
func (r *ChannelTaskLinkRepository) ListByChannel(ctx context.Context, channelID string) ([]*channelmodel.ChannelTaskLink, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var links []*channelmodel.ChannelTaskLink
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, strings.TrimSpace(channelID)).
		Order("linked_at DESC").
		Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

// UpdateStatus updates task status/note fields.
func (r *ChannelTaskLinkRepository) UpdateStatus(ctx context.Context, id string, status string, note string) error {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"status": status,
		"note":   note,
	}
	if strings.EqualFold(status, "done") {
		now := time.Now().UTC()
		updates["resolved_at"] = &now
	}
	return r.DB.WithContext(ctx).
		Model(&channelmodel.ChannelTaskLink{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id).
		Updates(updates).Error
}

// Delete removes a task link.
func (r *ChannelTaskLinkRepository) Delete(ctx context.Context, id string) error {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	return r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id).
		Delete(&channelmodel.ChannelTaskLink{}).Error
}
