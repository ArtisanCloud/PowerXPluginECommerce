package channel_master

import (
	"context"
	"math"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/sirupsen/logrus"
)

// HealthComputation holds score + derived labels to surface on UI.
type HealthComputation struct {
	Score  int      `json:"score"`
	Labels []string `json:"labels"`
}

// MetricsService aggregates KPI snapshots and computes health summary.
type MetricsService struct {
	metricsRepo    *channelrepo.ChannelMetricRepository
	credentialRepo *channelrepo.ChannelCredentialRepository
	logger         *logrus.Entry
	metrics        *channelobs.Metrics
}

// NewMetricsService creates the service dependencies.
func NewMetricsService(deps *app.Deps, metrics *channelobs.Metrics) *MetricsService {
	if deps == nil || deps.DB == nil {
		panic("metrics service requires database dependency")
	}
	logger := deps.RuntimeLogger(context.TODO(), "channel-metrics-service", nil)
	if metrics == nil {
		metrics = channelobs.NewMetrics(logger)
	}
	return &MetricsService{
		metricsRepo:    channelrepo.NewChannelMetricRepository(deps.DB),
		credentialRepo: channelrepo.NewChannelCredentialRepository(deps.DB),
		logger:         logger,
		metrics:        metrics,
	}
}

// LoadMetrics returns all recorded KPI snapshots for a channel.
func (s *MetricsService) LoadMetrics(ctx context.Context, channelID string) ([]*channelmodel.ChannelMetric, error) {
	if s == nil || s.metricsRepo == nil {
		return nil, nil
	}
	start := time.Now()
	metrics, err := s.metricsRepo.ListByChannel(ctx, channelID)
	if err == nil && s.metrics != nil {
		s.metrics.ObserveKPILoad(time.Since(start))
	}
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// ComputeHealthScore evaluates gating rules + weighted metrics.
func (s *MetricsService) ComputeHealthScore(ctx context.Context, channelID string, metrics []*channelmodel.ChannelMetric) HealthComputation {
	result := HealthComputation{Score: 80}
	if len(metrics) == 0 {
		result.Score = 50
		result.Labels = append(result.Labels, "no_metrics")
	} else {
		windowMetric := pickPrimaryMetric(metrics)
		result.Score = calculateBaseScore(windowMetric)
		result.Labels = append(result.Labels, windowMetricLabels(windowMetric)...)
	}
	if s != nil && s.credentialRepo != nil {
		if labels := s.credentialLabels(ctx, channelID); len(labels) > 0 {
			result.Labels = append(result.Labels, labels...)
			result.Score -= 30
		}
	}
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 100 {
		result.Score = 100
	}
	result.Labels = dedupeLabels(result.Labels)
	return result
}

func (s *MetricsService) credentialLabels(ctx context.Context, channelID string) []string {
	creds, err := s.credentialRepo.ListByChannel(ctx, channelID)
	if err != nil {
		if s.logger != nil {
			s.logger.WithError(err).WithField("channel_id", channelID).Warn("list credentials for health score failed")
		}
		return nil
	}
	if len(creds) == 0 {
		return []string{"credential_missing"}
	}
	var labels []string
	for _, cred := range creds {
		switch strings.ToLower(cred.Status) {
		case "expired":
			labels = append(labels, "credential_expired")
		case "expiring":
			labels = append(labels, "credential_expiring")
		case "test_failed":
			labels = append(labels, "credential_test_failed")
		}
	}
	return dedupeLabels(labels)
}

func pickPrimaryMetric(metrics []*channelmodel.ChannelMetric) *channelmodel.ChannelMetric {
	if len(metrics) == 0 {
		return nil
	}
	preferredOrder := map[string]int{"d7": 0, "d30": 1, "d1": 2}
	selected := metrics[0]
	bestRank := math.MaxInt
	for _, metric := range metrics {
		if metric == nil {
			continue
		}
		rank, ok := preferredOrder[strings.ToLower(metric.Window)]
		if !ok {
			rank = 5
		}
		if rank < bestRank {
			bestRank = rank
			selected = metric
		}
	}
	return selected
}

func calculateBaseScore(metric *channelmodel.ChannelMetric) int {
	if metric == nil {
		return 50
	}
	score := 85
	if metric.SyncSuccessRate < 0.95 {
		score -= 20
	}
	if metric.ErrorRate > 0.05 {
		score -= 15
	}
	if metric.InventoryCoverage < 0.85 {
		score -= 10
	}
	if metric.HealthScore > 0 {
		score = int(math.Round((float64(score) + float64(metric.HealthScore)) / 2))
	}
	if metric.GMVGrowthRate < 0 {
		score -= 5
	}
	return score
}

func windowMetricLabels(metric *channelmodel.ChannelMetric) []string {
	if metric == nil {
		return nil
	}
	var labels []string
	if metric.SyncSuccessRate < 0.95 {
		labels = append(labels, "sync_unstable")
	}
	if metric.ErrorRate > 0.05 {
		labels = append(labels, "error_rate_high")
	}
	if metric.InventoryCoverage < 0.85 {
		labels = append(labels, "inventory_low")
	}
	return labels
}

func dedupeLabels(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
