package subscription_reconciliation

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	SubscriptionReconciliationModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/subscription_reconciliation"
	SubscriptionReconciliationRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/subscription_reconciliation"
)

type dashboardFilters struct {
	channel       string
	plan          string
	region        string
	failureReason string
}

func (s *Service) queryDashboardBatches(ctx context.Context, fromDate, toDate string) ([]SubscriptionReconciliationModel.ReconciliationBatch, error) {
	tenantUUID, err := SubscriptionReconciliationRepo.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	q := s.batchRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND status = ?", tenantUUID, "completed")
	if strings.TrimSpace(fromDate) != "" {
		q = q.Where("billing_cycle >= ?", strings.TrimSpace(fromDate))
	}
	if strings.TrimSpace(toDate) != "" {
		q = q.Where("billing_cycle <= ?", strings.TrimSpace(toDate))
	}
	var batches []SubscriptionReconciliationModel.ReconciliationBatch
	if err := q.Order("billing_cycle DESC").Limit(200).Find(&batches).Error; err != nil {
		return nil, err
	}
	return batches, nil
}

func (s *Service) queryDashboardDeltas(
	ctx context.Context,
	batchIDs []string,
	fromDate, toDate string,
	filters dashboardFilters,
) ([]SubscriptionReconciliationModel.ReconciliationDelta, error) {
	if len(batchIDs) == 0 {
		return []SubscriptionReconciliationModel.ReconciliationDelta{}, nil
	}
	tenantUUID, err := SubscriptionReconciliationRepo.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	q := s.deltaRepo.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND batch_id IN ?", tenantUUID, batchIDs)
	if strings.TrimSpace(fromDate) != "" {
		if t, err := time.Parse("2006-01-02", fromDate); err == nil {
			q = q.Where("detected_at >= ?", t.UTC())
		}
	}
	if strings.TrimSpace(toDate) != "" {
		if t, err := time.Parse("2006-01-02", toDate); err == nil {
			q = q.Where("detected_at <= ?", t.Add(24*time.Hour).UTC())
		}
	}
	if strings.TrimSpace(filters.failureReason) != "" {
		q = q.Where("reason_code = ?", strings.TrimSpace(filters.failureReason))
	}
	var rows []SubscriptionReconciliationModel.ReconciliationDelta
	if err := q.Order("detected_at DESC").Limit(2000).Find(&rows).Error; err != nil {
		return nil, err
	}
	rows = filterDeltasByMetadata(rows, filters)
	return rows, nil
}

func filterDeltasByMetadata(rows []SubscriptionReconciliationModel.ReconciliationDelta, filters dashboardFilters) []SubscriptionReconciliationModel.ReconciliationDelta {
	channel := strings.TrimSpace(filters.channel)
	plan := strings.TrimSpace(filters.plan)
	region := strings.TrimSpace(filters.region)
	if channel == "" && plan == "" && region == "" {
		return rows
	}

	filtered := make([]SubscriptionReconciliationModel.ReconciliationDelta, 0, len(rows))
	for _, row := range rows {
		meta := map[string]any{}
		if len(row.Metadata) > 0 {
			_ = json.Unmarshal(row.Metadata, &meta)
		}
		if channel != "" && strings.TrimSpace(asString(meta["channel"])) != channel {
			continue
		}
		if plan != "" && strings.TrimSpace(asString(meta["plan"])) != plan {
			continue
		}
		if region != "" && strings.TrimSpace(asString(meta["region"])) != region {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
