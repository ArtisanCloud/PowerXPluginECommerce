package repository

import (
	"context"
	"strings"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// ChannelNoteRepository persists operator notes.
type ChannelNoteRepository struct {
	*repo.BaseRepository[channelmodel.ChannelNote]
}

// NewChannelNoteRepository constructs repository.
func NewChannelNoteRepository(db *gorm.DB) *ChannelNoteRepository {
	return &ChannelNoteRepository{BaseRepository: repo.NewBaseRepository[channelmodel.ChannelNote](db)}
}

// Create inserts a note scoped by tenant.
func (r *ChannelNoteRepository) Create(ctx context.Context, note *channelmodel.ChannelNote) (*channelmodel.ChannelNote, error) {
	if note == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(note.TenantUUID) == "" {
		note.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Create(ctx, note)
}

// ListByChannel fetches notes ordered by recency.
func (r *ChannelNoteRepository) ListByChannel(ctx context.Context, channelID string, limit int) ([]*channelmodel.ChannelNote, error) {
	if limit <= 0 {
		limit = 20
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var notes []*channelmodel.ChannelNote
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, strings.TrimSpace(channelID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}
