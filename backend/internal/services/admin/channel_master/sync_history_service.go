package channel_master

import (
	"context"
	"errors"
	"strings"
	"time"

	channelmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/channel_master"
	channelrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/channel_master"
	channelobs "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/channel/master"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/sirupsen/logrus"
)

// SyncTriggerInput captures manual trigger payload.
type SyncTriggerInput struct {
	TriggerType string
	Payload     map[string]any
}

// SyncHistoryService persists executions and exposes listings.
type SyncHistoryService struct {
	repo    *channelrepo.ChannelSyncHistoryRepository
	metrics *channelobs.Metrics
	logger  *logrus.Entry
}

// NewSyncHistoryService creates service.
func NewSyncHistoryService(deps *app.Deps, metrics *channelobs.Metrics) *SyncHistoryService {
	if deps == nil || deps.DB == nil {
		panic("sync history service requires database dependency")
	}
	logger := deps.RuntimeLogger(context.TODO(), "channel-sync-service", nil)
	if metrics == nil {
		metrics = channelobs.NewMetrics(logger)
	}
	return &SyncHistoryService{
		repo:    channelrepo.NewChannelSyncHistoryRepository(deps.DB),
		metrics: metrics,
		logger:  logger,
	}
}

// TriggerManual inserts a history row to indicate manual sync request.
func (s *SyncHistoryService) TriggerManual(ctx context.Context, channelID string, input SyncTriggerInput) (*channelmodel.ChannelSyncHistory, error) {
	if stringsTrim(channelID) == "" {
		return nil, errors.New("channel id required")
	}
	triggerType := stringsTrim(input.TriggerType)
	if triggerType == "" {
		triggerType = "manual"
	}
	history := &channelmodel.ChannelSyncHistory{
		ID:          utils.NewUUID(),
		ChannelID:   channelID,
		TriggerType: triggerType,
		TriggeredBy: actorFromContext(ctx),
		Result:      "pending",
		Payload:     jsonMap(input.Payload),
	}
	saved, err := s.repo.Create(ctx, history)
	if err != nil {
		return nil, err
	}
	s.logger.WithFields(logrus.Fields{
		"channel_id":   channelID,
		"trigger_type": triggerType,
		"history_id":   saved.ID,
	}).Info("channel sync triggered")
	return saved, nil
}

// Complete updates status + metrics counters.
func (s *SyncHistoryService) Complete(ctx context.Context, historyID string, duration time.Duration, success bool, payload map[string]any) error {
	if stringsTrim(historyID) == "" {
		return errors.New("history id required")
	}
	result := "failed"
	if success {
		result = "success"
	}
	if err := s.repo.DB.WithContext(ctx).Model(&channelmodel.ChannelSyncHistory{}).
		Where("id = ?", historyID).
		Updates(map[string]any{
			"duration_ms": duration.Milliseconds(),
			"result":      result,
			"payload":     jsonMap(payload),
		}).Error; err != nil {
		return err
	}
	s.metrics.RecordSyncOutcome(success)
	return nil
}

// ListRecent returns recent executions.
func (s *SyncHistoryService) ListRecent(ctx context.Context, channelID string, limit int) ([]*channelmodel.ChannelSyncHistory, error) {
	return s.repo.ListRecent(ctx, channelID, limit)
}

func stringsTrim(val string) string {
	return strings.TrimSpace(val)
}
