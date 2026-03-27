package logistics

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type GatewayMetricsService struct {
	jobRepo *LogisticsRepo.TrackingSyncJobRepository
}

type GatewayHealthSummary struct {
	WindowHours     int     `json:"window_hours"`
	TotalRequests   int     `json:"total_requests"`
	SuccessRequests int     `json:"success_requests"`
	FailedRequests  int     `json:"failed_requests"`
	SuccessRate     float64 `json:"success_rate"`
	P95LatencyMS    int     `json:"p95_latency_ms"`
}

type GatewayCarrierHealthItem struct {
	CarrierID       string  `json:"carrier_id"`
	TotalRequests   int     `json:"total_requests"`
	SuccessRequests int     `json:"success_requests"`
	FailedRequests  int     `json:"failed_requests"`
	SuccessRate     float64 `json:"success_rate"`
	P95LatencyMS    int     `json:"p95_latency_ms"`
}

type GatewayAlertItem struct {
	Level   string `json:"level"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type GatewayHealthSnapshot struct {
	Summary  GatewayHealthSummary       `json:"summary"`
	Carriers []GatewayCarrierHealthItem `json:"carriers"`
	Alerts   []GatewayAlertItem         `json:"alerts"`
}

func NewGatewayMetricsService(deps *app.Deps) *GatewayMetricsService {
	if deps == nil || deps.DB == nil {
		return &GatewayMetricsService{}
	}
	return &GatewayMetricsService{jobRepo: LogisticsRepo.NewTrackingSyncJobRepository(deps.DB)}
}

func (s *GatewayMetricsService) Snapshot(ctx context.Context, tenantUUID string, windowHours int) (*GatewayHealthSnapshot, error) {
	if s == nil || s.jobRepo == nil {
		return nil, errors.New("gateway metrics service unavailable")
	}
	if windowHours <= 0 {
		windowHours = 24
	}
	if windowHours > 24*30 {
		windowHours = 24 * 30
	}
	ctx = withTenantContext(ctx, tenantUUID)
	since := time.Now().UTC().Add(-time.Duration(windowHours) * time.Hour)
	jobs, err := s.jobRepo.List(ctx, LogisticsRepo.TrackingSyncJobFilter{
		Limit: 500,
		Since: &since,
	})
	if err != nil {
		return nil, err
	}
	summary := GatewayHealthSummary{WindowHours: windowHours}
	type carrierAgg struct {
		total   int
		success int
		failed  int
		p95     int
	}
	carrierMap := map[string]*carrierAgg{}
	for _, row := range jobs {
		key := strings.TrimSpace(row.CarrierID)
		if key == "" {
			key = "all"
		}
		item := carrierMap[key]
		if item == nil {
			item = &carrierAgg{}
			carrierMap[key] = item
		}
		item.total += row.TotalWaybills
		item.success += row.SuccessCount
		item.failed += row.FailedCount
		if row.P95LatencyMS > item.p95 {
			item.p95 = row.P95LatencyMS
		}

		summary.TotalRequests += row.TotalWaybills
		summary.SuccessRequests += row.SuccessCount
		summary.FailedRequests += row.FailedCount
		if row.P95LatencyMS > summary.P95LatencyMS {
			summary.P95LatencyMS = row.P95LatencyMS
		}
	}
	summary.SuccessRate = ratio(summary.SuccessRequests, summary.TotalRequests)
	carriers := make([]GatewayCarrierHealthItem, 0, len(carrierMap))
	for carrierID, item := range carrierMap {
		carriers = append(carriers, GatewayCarrierHealthItem{
			CarrierID:       carrierID,
			TotalRequests:   item.total,
			SuccessRequests: item.success,
			FailedRequests:  item.failed,
			SuccessRate:     ratio(item.success, item.total),
			P95LatencyMS:    item.p95,
		})
	}
	alerts := buildGatewayAlerts(summary, jobs)
	return &GatewayHealthSnapshot{
		Summary:  summary,
		Carriers: carriers,
		Alerts:   alerts,
	}, nil
}

func ratio(success, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(success) * 100 / float64(total)
}

func buildGatewayAlerts(summary GatewayHealthSummary, jobs []LogisticsModel.TrackingSyncJob) []GatewayAlertItem {
	alerts := make([]GatewayAlertItem, 0, 3)
	if summary.TotalRequests > 0 && summary.SuccessRate < 95 {
		alerts = append(alerts, GatewayAlertItem{
			Level:   "warning",
			Code:    "success_rate_low",
			Message: fmt.Sprintf("网关成功率 %.2f%% 低于阈值 95%%", summary.SuccessRate),
		})
	}
	if summary.P95LatencyMS > 2000 {
		alerts = append(alerts, GatewayAlertItem{
			Level:   "warning",
			Code:    "latency_high",
			Message: fmt.Sprintf("网关 P95 延迟 %dms 超过阈值 2000ms", summary.P95LatencyMS),
		})
	}
	recentFailed := 0
	for _, row := range jobs {
		if strings.EqualFold(row.Status, "failed") || strings.EqualFold(row.Status, "partial_failed") {
			recentFailed++
		}
	}
	if recentFailed > 0 {
		alerts = append(alerts, GatewayAlertItem{
			Level:   "info",
			Code:    "failed_jobs_detected",
			Message: fmt.Sprintf("窗口内存在 %d 个失败/部分失败同步任务", recentFailed),
		})
	}
	return alerts
}
