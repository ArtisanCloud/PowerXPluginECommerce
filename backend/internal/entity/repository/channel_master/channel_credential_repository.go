package repository

import (
	"context"
	"strings"
	"time"

	credentialmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelCredentialRepository manages persistence for channel credentials.
type ChannelCredentialRepository struct {
	*repo.BaseRepository[credentialmodel.ChannelCredential]
}

// NewChannelCredentialRepository creates repository.
func NewChannelCredentialRepository(db *gorm.DB) *ChannelCredentialRepository {
	return &ChannelCredentialRepository{BaseRepository: repo.NewBaseRepository[credentialmodel.ChannelCredential](db)}
}

// UniqueConstraintColumns enforces tenant+channel+type uniqueness.
func (r *ChannelCredentialRepository) UniqueConstraintColumns() []clause.Column {
	return []clause.Column{{Name: "tenant_uuid"}, {Name: "channel_id"}, {Name: "type"}}
}

// Upsert saves credential ensuring tenant context.
func (r *ChannelCredentialRepository) Upsert(ctx context.Context, credential *credentialmodel.ChannelCredential) (*credentialmodel.ChannelCredential, error) {
	if credential == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(credential.TenantUUID) == "" {
		credential.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Upsert(ctx, credential, r.UniqueConstraintColumns())
}

// ListByChannel returns credentials bound to channel.
func (r *ChannelCredentialRepository) ListByChannel(ctx context.Context, channelID string) ([]*credentialmodel.ChannelCredential, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var creds []*credentialmodel.ChannelCredential
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, channelID).
		Order("updated_at DESC").
		Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

// ListExpiring returns credentials expiring within duration.
func (r *ChannelCredentialRepository) ListExpiring(ctx context.Context, within time.Duration) ([]*credentialmodel.ChannelCredential, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(within)
	var creds []*credentialmodel.ChannelCredential
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND expires_at IS NOT NULL AND expires_at <= ?", tenantUUID, cutoff).
		Where("status <> ?", "expired").
		Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

// Update persists an existing credential.
func (r *ChannelCredentialRepository) Update(ctx context.Context, credential *credentialmodel.ChannelCredential) (*credentialmodel.ChannelCredential, error) {
	if credential == nil {
		return nil, gorm.ErrInvalidData
	}
	return r.BaseRepository.Update(ctx, credential)
}
