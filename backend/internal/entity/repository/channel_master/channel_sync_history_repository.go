package repository

import (
	"context"
	"strings"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// ChannelSyncHistoryRepository persists sync executions.
type ChannelSyncHistoryRepository struct {
	*repo.BaseRepository[channelmodel.ChannelSyncHistory]
}

// NewChannelSyncHistoryRepository instantiates repository.
func NewChannelSyncHistoryRepository(db *gorm.DB) *ChannelSyncHistoryRepository {
	return &ChannelSyncHistoryRepository{BaseRepository: repo.NewBaseRepository[channelmodel.ChannelSyncHistory](db)}
}

// Create stores a sync history row.
func (r *ChannelSyncHistoryRepository) Create(ctx context.Context, history *channelmodel.ChannelSyncHistory) (*channelmodel.ChannelSyncHistory, error) {
	if history == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(history.TenantUUID) == "" {
		history.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, history)
}

// ListRecent returns recent history rows limited by count.
func (r *ChannelSyncHistoryRepository) ListRecent(ctx context.Context, channelID string, limit int) ([]*channelmodel.ChannelSyncHistory, error) {
	if limit <= 0 {
		limit = 10
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var history []*channelmodel.ChannelSyncHistory
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, strings.TrimSpace(channelID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}
