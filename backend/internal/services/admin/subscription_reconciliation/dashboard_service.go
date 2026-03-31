package subscription_reconciliation

import (
	"context"
	"math"
	"strings"
	"time"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	SubscriptionReconciliationRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/subscription_reconciliation"
)

type DashboardDeltaTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

func (s *Service) buildDashboard(ctx context.Context, query DashboardQuery) (map[string]any, error) {
	from, to := normalizeDashboardDateRange(query.From, query.To)
	filters := dashboardFilters{
		channel:       strings.TrimSpace(query.Channel),
		plan:          strings.TrimSpace(query.Plan),
		region:        strings.TrimSpace(query.Region),
		failureReason: strings.TrimSpace(query.FailureReason),
	}
	batches, err := s.queryDashboardBatches(ctx, from, to)
	if err != nil {
		return nil, err
	}
	batchIDs := make([]string, 0, len(batches))
	var expectedAmount int64
	for _, batch := range batches {
		batchIDs = append(batchIDs, batch.ID)
		expectedAmount += batch.ExpectedAmountMinor
	}

	deltas, err := s.queryDashboardDeltas(ctx, batchIDs, from, to, filters)
	if err != nil {
		return nil, err
	}

	deltaAbsAmount := int64(0)
	byDeltaTypeMap := map[string]int{}
	for _, delta := range deltas {
		byDeltaTypeMap[delta.DeltaType]++
		deltaAbsAmount += int64(math.Abs(float64(delta.DeltaAmountMinor)))
	}
	byDeltaType := make([]DashboardDeltaTypeCount, 0, len(byDeltaTypeMap))
	for k, v := range byDeltaTypeMap {
		byDeltaType = append(byDeltaType, DashboardDeltaTypeCount{Type: k, Count: v})
	}

	deltaRate := 0.0
	if expectedAmount > 0 {
		deltaRate = roundRatio(float64(deltaAbsAmount) / float64(expectedAmount))
	}

	recoveryRate, err := s.queryRecoveryRate(ctx, from, to)
	if err != nil {
		return nil, err
	}
	avgHandleHours, pendingTasks, err := s.queryTaskMetrics(ctx, from, to)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"deltaRate":      deltaRate,
		"recoveryRate":   recoveryRate,
		"avgHandleHours": avgHandleHours,
		"pendingTasks":   pendingTasks,
		"byDeltaType":    byDeltaType,
	}, nil
}

func (s *Service) queryRecoveryRate(ctx context.Context, fromDate, toDate string) (float64, error) {
	tenantUUID, err := SubscriptionReconciliationRepo.RequireTenantUUID(ctx)
	if err != nil {
		return 0, err
	}
	q := s.executionRepo.DB.WithContext(ctx).
		Model(&SubscriptionReconciliationModel.RenewalExecutionLog{}).
		Where("tenant_uuid = ? AND action_type = ?", tenantUUID, "retry")
	if strings.TrimSpace(fromDate) != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			q = q.Where("executed_at >= ?", t.UTC())
		}
	}
	if strings.TrimSpace(toDate) != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			q = q.Where("executed_at <= ?", t.Add(24*time.Hour).UTC())
		}
	}
	var rows []SubscriptionReconciliationModel.RenewalExecutionLog
	if err := q.Order("executed_at DESC").Limit(5000).Find(&rows).Error; err != nil {
		return 0, err
	}
	totalSet := map[string]struct{}{}
	successSet := map[string]struct{}{}
	for _, row := range rows {
		ref := strings.TrimSpace(row.SubscriptionRef)
		if ref == "" {
			continue
		}
		totalSet[ref] = struct{}{}
		if row.Result == "success" {
			successSet[ref] = struct{}{}
		}
	}
	if len(totalSet) == 0 {
		return 0, nil
	}
	return roundRatio(float64(len(successSet)) / float64(len(totalSet))), nil
}

func (s *Service) queryTaskMetrics(ctx context.Context, fromDate, toDate string) (float64, int64, error) {
	tenantUUID, err := SubscriptionReconciliationRepo.RequireTenantUUID(ctx)
	if err != nil {
		return 0, 0, err
	}
	taskQ := s.taskRepo.DB.WithContext(ctx).
		Model(&SubscriptionReconciliationModel.DeltaTask{}).
		Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(fromDate) != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			taskQ = taskQ.Where("created_at >= ?", t.UTC())
		}
	}
	if strings.TrimSpace(toDate) != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			taskQ = taskQ.Where("created_at <= ?", t.Add(24*time.Hour).UTC())
		}
	}
	var tasks []SubscriptionReconciliationModel.DeltaTask
	if err := taskQ.Order("created_at DESC").Limit(2000).Find(&tasks).Error; err != nil {
		return 0, 0, err
	}

	pending := int64(0)
	totalHours := 0.0
	closedCount := 0
	for _, task := range tasks {
		if task.Status != "closed" {
			pending++
			continue
		}
		if task.ClosedAt == nil {
			continue
		}
		hours := task.ClosedAt.Sub(task.CreatedAt).Hours()
		if hours < 0 {
			hours = 0
		}
		totalHours += hours
		closedCount++
	}
	if closedCount == 0 {
		return 0, pending, nil
	}
	return roundRatio(totalHours / float64(closedCount)), pending, nil
}

func normalizeDashboardDateRange(from, to string) (string, string) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" && to == "" {
		now := time.Now().UTC()
		return now.AddDate(0, 0, -7).Format("2006-01-02"), now.Format("2006-01-02")
	}
	return from, to
}

func roundRatio(v float64) float64 {
	return math.Round(v*10000) / 10000
}
