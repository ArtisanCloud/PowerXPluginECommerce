package master

import (
	"context"
	"math/rand"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// MetricRefreshJob periodically refreshes KPI snapshots for demo purposes.
type MetricRefreshJob struct {
	db          *gorm.DB
	metricRepo  *channelrepo.ChannelMetricRepository
	channelRepo *channelrepo.ChannelMasterRepository
	logger      *logrus.Entry
	interval    time.Duration
}

// NewMetricRefreshJob creates a background job.
func NewMetricRefreshJob(deps *app.Deps, interval time.Duration) *MetricRefreshJob {
	if deps == nil || deps.DB == nil {
		return nil
	}
	if interval <= 0 {
		interval = 30 * time.Minute
	}
	logger := deps.RuntimeLogger(nil, "channel-metric-refresh", nil)
	return &MetricRefreshJob{
		db:          deps.DB,
		metricRepo:  channelrepo.NewChannelMetricRepository(deps.DB),
		channelRepo: channelrepo.NewChannelMasterRepository(deps.DB),
		logger:      logger,
		interval:    interval,
	}
}

// Run starts ticker loop.
func (j *MetricRefreshJob) Run(ctx context.Context) {
	if j == nil || j.db == nil {
		return
	}
	j.logger.Info("channel metric refresh job started")
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	j.refreshOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			j.logger.Info("channel metric refresh job stopped")
			return
		case <-ticker.C:
			j.refreshOnce(ctx)
		}
	}
}

func (j *MetricRefreshJob) refreshOnce(ctx context.Context) {
	var tenantIDs []string
	if err := j.db.WithContext(ctx).
		Model(&channelmodel.ChannelMaster{}).
		Distinct().
		Pluck("tenant_uuid", &tenantIDs).Error; err != nil {
		j.logger.WithError(err).Warn("metric refresh: failed to list tenants")
		return
	}
	for _, tenant := range tenantIDs {
		if tenant == "" {
			continue
		}
		tenantCtx := authx.ContextWithTenantUUID(ctx, tenant)
		var channels []channelmodel.ChannelMaster
		if err := j.db.WithContext(tenantCtx).
			Where("tenant_uuid = ?", tenant).
			Limit(50).
			Find(&channels).Error; err != nil {
			j.logger.WithError(err).WithField("tenant", tenant).Warn("metric refresh: list channels failed")
			continue
		}
		for _, channel := range channels {
			j.upsertMetric(tenantCtx, &channel)
		}
	}
}

func (j *MetricRefreshJob) upsertMetric(ctx context.Context, channel *channelmodel.ChannelMaster) {
	if channel == nil {
		return
	}
	now := time.Now().UTC()
	score := 70
	if channel.HealthScore != nil {
		score = *channel.HealthScore
	}
	randomGrowth := (rand.Float64() * 0.2) - 0.1
	metric := &channelmodel.ChannelMetric{
		ChannelID:         channel.ID,
		Window:            "d7",
		GMV:               float64(score) * 1000,
		Orders:            int64(50 + rand.Intn(200)),
		GMVGrowthRate:     randomGrowth,
		InventoryCoverage: 0.8 + rand.Float64()*0.2,
		ErrorRate:         rand.Float64() * 0.1,
		SyncSuccessRate:   0.9 + rand.Float64()*0.1,
		HealthScore:       score,
		SourceTimestamp:   &now,
		SourceJobID:       "metric-refresh",
	}
	if _, err := j.metricRepo.Upsert(ctx, metric); err != nil {
		j.logger.WithError(err).WithField("channel_id", channel.ID).Warn("metric refresh: upsert failed")
	}
}
