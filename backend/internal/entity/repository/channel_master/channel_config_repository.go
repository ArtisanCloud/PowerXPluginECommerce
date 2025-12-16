package repository

import (
	"context"
	"strings"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelConfigRepository persists strategy/team configuration per channel.
type ChannelConfigRepository struct {
	*repo.BaseRepository[channelmodel.ChannelConfig]
}

// NewChannelConfigRepository constructs the repository.
func NewChannelConfigRepository(db *gorm.DB) *ChannelConfigRepository {
	return &ChannelConfigRepository{BaseRepository: repo.NewBaseRepository[channelmodel.ChannelConfig](db)}
}

var channelConfigUniqueKey = []clause.Column{{Name: "tenant_uuid"}, {Name: "channel_id"}}

// Upsert stores configuration scoped by tenant/channel.
func (r *ChannelConfigRepository) Upsert(ctx context.Context, cfg *channelmodel.ChannelConfig) (*channelmodel.ChannelConfig, error) {
	if cfg == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.TenantUUID) == "" {
		cfg.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Upsert(ctx, cfg, channelConfigUniqueKey)
}

// FindByChannel fetches the latest config row for a channel.
func (r *ChannelConfigRepository) FindByChannel(ctx context.Context, channelID string) (*channelmodel.ChannelConfig, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var cfg channelmodel.ChannelConfig
	tx := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, strings.TrimSpace(channelID)).
		Order("updated_at DESC").
		Limit(1).
		First(&cfg)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &cfg, nil
}
