package logistics

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

type GatewayCostService struct {
	usageRepo *LogisticsRepo.GatewayUsageRepository
	jobRepo   *LogisticsRepo.TrackingSyncJobRepository
}

type GatewayCostQuery struct {
	WindowHours int     `json:"window_hours,omitempty"`
	CarrierID   string  `json:"carrier_id,omitempty"`
	Provider    string  `json:"provider,omitempty"`
	UnitPrice   float64 `json:"unit_price,omitempty"`
	QuotaLimit  int     `json:"quota_limit,omitempty"`
}

type GatewayCostSummary struct {
	WindowHours    int     `json:"window_hours"`
	TotalRequests  int     `json:"total_requests"`
	SuccessCount   int     `json:"success_count"`
	FailedCount    int     `json:"failed_count"`
	TotalCost      float64 `json:"total_cost"`
	QuotaLimit     int     `json:"quota_limit"`
	QuotaUsed      int     `json:"quota_used"`
	QuotaUsageRate float64 `json:"quota_usage_rate"`
}

type GatewayCostCarrierItem struct {
	CarrierID     string  `json:"carrier_id"`
	Provider      string  `json:"provider"`
	RequestCount  int     `json:"request_count"`
	SuccessCount  int     `json:"success_count"`
	FailedCount   int     `json:"failed_count"`
	CostAmount    float64 `json:"cost_amount"`
	QuotaConsumed int     `json:"quota_consumed"`
}

type GatewayCostAlert struct {
	Level       string  `json:"level"`
	Code        string  `json:"code"`
	Message     string  `json:"message"`
	CarrierID   string  `json:"carrier_id,omitempty"`
	Provider    string  `json:"provider,omitempty"`
	UsageRate   float64 `json:"usage_rate,omitempty"`
	RequestRate float64 `json:"request_rate,omitempty"`
}

type GatewayCostSnapshot struct {
	Summary  GatewayCostSummary       `json:"summary"`
	Carriers []GatewayCostCarrierItem `json:"carriers"`
	Alerts   []GatewayCostAlert       `json:"alerts"`
}

func NewGatewayCostService(deps *app.Deps) *GatewayCostService {
	if deps == nil || deps.DB == nil {
		return &GatewayCostService{}
	}
	return &GatewayCostService{
		usageRepo: LogisticsRepo.NewGatewayUsageRepository(deps.DB),
		jobRepo:   LogisticsRepo.NewTrackingSyncJobRepository(deps.DB),
	}
}

func (s *GatewayCostService) Snapshot(ctx context.Context, tenantUUID string, query GatewayCostQuery) (*GatewayCostSnapshot, error) {
	if s == nil || s.usageRepo == nil || s.jobRepo == nil {
		return nil, errors.New("gateway cost service unavailable")
	}
	windowHours := query.WindowHours
	if windowHours <= 0 {
		windowHours = 24
	}
	if windowHours > 24*30 {
		windowHours = 24 * 30
	}
	quotaLimit := query.QuotaLimit
	if quotaLimit <= 0 {
		quotaLimit = 5000
	}
	unitPrice := query.UnitPrice
	if unitPrice <= 0 {
		unitPrice = 0.02
	}

	ctx = withTenantContext(ctx, tenantUUID)
	now := time.Now().UTC()
	from := now.Add(-time.Duration(windowHours) * time.Hour)

	if err := s.ingestUsageFromJobs(ctx, tenantUUID, from, now, unitPrice); err != nil {
		return nil, err
	}

	rows, err := s.usageRepo.List(ctx, LogisticsRepo.GatewayUsageFilter{
		CarrierID: strings.TrimSpace(query.CarrierID),
		Provider:  strings.TrimSpace(query.Provider),
		From:      &from,
		To:        &now,
		Limit:     2000,
	})
	if err != nil {
		return nil, err
	}

	summary := GatewayCostSummary{
		WindowHours: windowHours,
		QuotaLimit:  quotaLimit,
	}
	type carrierAgg struct {
		item GatewayCostCarrierItem
	}
	carrierMap := make(map[string]*carrierAgg)
	for _, row := range rows {
		key := strings.TrimSpace(row.CarrierID) + "|" + strings.TrimSpace(row.Provider)
		agg := carrierMap[key]
		if agg == nil {
			agg = &carrierAgg{
				item: GatewayCostCarrierItem{
					CarrierID: strings.TrimSpace(row.CarrierID),
					Provider:  strings.TrimSpace(row.Provider),
				},
			}
			carrierMap[key] = agg
		}
		agg.item.RequestCount += row.RequestCount
		agg.item.SuccessCount += row.SuccessCount
		agg.item.FailedCount += row.FailedCount
		agg.item.CostAmount += row.CostAmount
		agg.item.QuotaConsumed += row.QuotaConsumed

		summary.TotalRequests += row.RequestCount
		summary.SuccessCount += row.SuccessCount
		summary.FailedCount += row.FailedCount
		summary.TotalCost += row.CostAmount
		summary.QuotaUsed += row.QuotaConsumed
	}
	summary.QuotaUsageRate = usageRate(summary.QuotaUsed, summary.QuotaLimit)

	carriers := make([]GatewayCostCarrierItem, 0, len(carrierMap))
	for _, agg := range carrierMap {
		carriers = append(carriers, agg.item)
	}
	sort.Slice(carriers, func(i, j int) bool {
		if carriers[i].CostAmount == carriers[j].CostAmount {
			return carriers[i].RequestCount > carriers[j].RequestCount
		}
		return carriers[i].CostAmount > carriers[j].CostAmount
	})

	alerts := s.buildAlerts(summary, carriers)
	return &GatewayCostSnapshot{
		Summary:  summary,
		Carriers: carriers,
		Alerts:   alerts,
	}, nil
}

func (s *GatewayCostService) Alerts(ctx context.Context, tenantUUID string, query GatewayCostQuery) ([]GatewayCostAlert, error) {
	snapshot, err := s.Snapshot(ctx, tenantUUID, query)
	if err != nil {
		return nil, err
	}
	return snapshot.Alerts, nil
}

func (s *GatewayCostService) ingestUsageFromJobs(
	ctx context.Context,
	tenantUUID string,
	from, to time.Time,
	unitPrice float64,
) error {
	jobs, err := s.jobRepo.List(ctx, LogisticsRepo.TrackingSyncJobFilter{
		CarrierID: "",
		Status:    "",
		Limit:     2000,
		Since:     &from,
	})
	if err != nil {
		return err
	}
	for _, job := range jobs {
		jobID := strings.TrimSpace(job.ID)
		if jobID == "" {
			continue
		}
		_, getErr := s.usageRepo.GetBySourceJobID(ctx, jobID)
		if getErr == nil {
			continue
		}
		if !errors.Is(getErr, gorm.ErrRecordNotFound) {
			return getErr
		}
		requestCount := job.TotalWaybills
		if requestCount < 0 {
			requestCount = 0
		}
		successCount := job.SuccessCount
		if successCount < 0 {
			successCount = 0
		}
		failedCount := job.FailedCount
		if failedCount < 0 {
			failedCount = 0
		}
		windowEnd := job.UpdatedAt.UTC()
		if windowEnd.IsZero() {
			windowEnd = to
		}
		windowStart := job.CreatedAt.UTC()
		if windowStart.IsZero() {
			windowStart = windowEnd
		}
		usage := &LogisticsModel.GatewayUsage{
			ID:            utils.NewUUID(),
			TenantUUID:    strings.TrimSpace(tenantUUID),
			CarrierID:     strings.TrimSpace(job.CarrierID),
			Provider:      strings.TrimSpace(job.Provider),
			SourceJobID:   jobID,
			RequestCount:  requestCount,
			SuccessCount:  successCount,
			FailedCount:   failedCount,
			UnitPrice:     unitPrice,
			CostAmount:    float64(requestCount) * unitPrice,
			QuotaConsumed: requestCount,
			WindowStartAt: windowStart,
			WindowEndAt:   windowEnd,
			Metadata:      []byte("{}"),
		}
		if err := s.usageRepo.Create(ctx, usage); err != nil {
			return err
		}
	}
	return nil
}

func (s *GatewayCostService) buildAlerts(summary GatewayCostSummary, carriers []GatewayCostCarrierItem) []GatewayCostAlert {
	alerts := make([]GatewayCostAlert, 0, 8)
	if summary.QuotaLimit > 0 && summary.QuotaUsed > summary.QuotaLimit {
		alerts = append(alerts, GatewayCostAlert{
			Level:     "critical",
			Code:      "quota_exceeded",
			Message:   fmt.Sprintf("网关配额超限：%d/%d", summary.QuotaUsed, summary.QuotaLimit),
			UsageRate: summary.QuotaUsageRate,
		})
	} else if summary.QuotaUsageRate >= 80 {
		alerts = append(alerts, GatewayCostAlert{
			Level:     "warning",
			Code:      "quota_high_usage",
			Message:   fmt.Sprintf("网关配额使用率 %.2f%%，接近上限", summary.QuotaUsageRate),
			UsageRate: summary.QuotaUsageRate,
		})
	}
	failRate := usageRate(summary.FailedCount, summary.TotalRequests)
	if summary.TotalRequests > 0 && failRate >= 10 {
		alerts = append(alerts, GatewayCostAlert{
			Level:       "warning",
			Code:        "gateway_failure_rate_high",
			Message:     fmt.Sprintf("网关失败率 %.2f%% 超过阈值 10%%", failRate),
			RequestRate: failRate,
		})
	}
	for _, item := range carriers {
		if summary.QuotaLimit <= 0 {
			continue
		}
		carrierRate := usageRate(item.QuotaConsumed, summary.QuotaLimit)
		if carrierRate < 50 {
			continue
		}
		level := "info"
		if carrierRate >= 80 {
			level = "warning"
		}
		alerts = append(alerts, GatewayCostAlert{
			Level:     level,
			Code:      "carrier_quota_hotspot",
			Message:   fmt.Sprintf("承运商 %s 占用配额 %.2f%%", item.CarrierID, carrierRate),
			CarrierID: item.CarrierID,
			Provider:  item.Provider,
			UsageRate: carrierRate,
		})
	}
	return alerts
}

func usageRate(current, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(current) * 100 / float64(total)
}
