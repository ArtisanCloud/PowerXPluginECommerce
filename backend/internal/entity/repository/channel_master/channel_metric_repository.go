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

// ChannelMetricRepository manages KPI snapshots.
type ChannelMetricRepository struct {
	*repo.BaseRepository[channelmodel.ChannelMetric]
}

// NewChannelMetricRepository builds a repository for metrics.
func NewChannelMetricRepository(db *gorm.DB) *ChannelMetricRepository {
	return &ChannelMetricRepository{BaseRepository: repo.NewBaseRepository[channelmodel.ChannelMetric](db)}
}

// UniqueConstraint enforces tenant+channel+window combination.
func (r *ChannelMetricRepository) UniqueConstraint() []clause.Column {
	return []clause.Column{
		{Name: "tenant_uuid"},
		{Name: "channel_id"},
		{Name: "window"},
	}
}

// Upsert stores a KPI snapshot.
func (r *ChannelMetricRepository) Upsert(ctx context.Context, metric *channelmodel.ChannelMetric) (*channelmodel.ChannelMetric, error) {
	if metric == nil {
		return nil, gorm.ErrInvalidData
	}
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(metric.TenantUUID) == "" {
		metric.TenantUUID = tenantUUID
	}
	return r.BaseRepository.Upsert(ctx, metric, r.UniqueConstraint())
}

// ListByChannel returns ordered metrics for a channel.
func (r *ChannelMetricRepository) ListByChannel(ctx context.Context, channelID string) ([]*channelmodel.ChannelMetric, error) {
	tenantUUID, err := authx.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	var metrics []*channelmodel.ChannelMetric
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND channel_id = ?", tenantUUID, strings.TrimSpace(channelID)).
		Order("window ASC").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}
