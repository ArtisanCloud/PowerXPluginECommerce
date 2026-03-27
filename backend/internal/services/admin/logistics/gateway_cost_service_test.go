package logistics

import (
	"context"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestGatewayCostService_SnapshotAggregatesAndAlerts(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "gateway_cost_snapshot")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-cost-1",
		TenantUUID:    "tenant-a",
		CarrierID:     "carrier-a",
		Provider:      "gateway",
		Status:        "success",
		TotalWaybills: 120,
		SuccessCount:  110,
		FailedCount:   10,
		CreatedAt:     now.Add(-30 * time.Minute),
		UpdatedAt:     now.Add(-20 * time.Minute),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-cost-2",
		TenantUUID:    "tenant-a",
		CarrierID:     "carrier-b",
		Provider:      "gateway",
		Status:        "success",
		TotalWaybills: 40,
		SuccessCount:  40,
		FailedCount:   0,
		CreatedAt:     now.Add(-10 * time.Minute),
		UpdatedAt:     now.Add(-5 * time.Minute),
	}).Error)

	otherTenantCtx := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	require.NoError(t, db.WithContext(otherTenantCtx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-cost-other",
		TenantUUID:    "tenant-b",
		CarrierID:     "carrier-z",
		Provider:      "gateway",
		Status:        "success",
		TotalWaybills: 999,
		SuccessCount:  999,
		FailedCount:   0,
		CreatedAt:     now.Add(-15 * time.Minute),
		UpdatedAt:     now.Add(-15 * time.Minute),
	}).Error)

	svc := NewGatewayCostService(&app.Deps{DB: db})
	snapshot, err := svc.Snapshot(ctx, "tenant-a", GatewayCostQuery{
		WindowHours: 24,
		UnitPrice:   0.1,
		QuotaLimit:  100,
	})
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	require.Equal(t, 160, snapshot.Summary.TotalRequests)
	require.Equal(t, 150, snapshot.Summary.SuccessCount)
	require.Equal(t, 10, snapshot.Summary.FailedCount)
	require.InDelta(t, 16.0, snapshot.Summary.TotalCost, 0.0001)
	require.InDelta(t, 160.0, snapshot.Summary.QuotaUsageRate, 0.0001)
	require.Len(t, snapshot.Carriers, 2)
	require.NotEmpty(t, snapshot.Alerts)

	hasQuotaExceeded := false
	for _, alert := range snapshot.Alerts {
		if alert.Code == "quota_exceeded" {
			hasQuotaExceeded = true
			break
		}
	}
	require.True(t, hasQuotaExceeded)
}

func TestGatewayCostService_IdempotentIngestAndTenantIsolation(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "gateway_cost_idempotent")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-idem-1",
		TenantUUID:    "tenant-a",
		CarrierID:     "carrier-a",
		Provider:      "gateway",
		Status:        "success",
		TotalWaybills: 20,
		SuccessCount:  20,
		FailedCount:   0,
		CreatedAt:     now.Add(-30 * time.Minute),
		UpdatedAt:     now.Add(-10 * time.Minute),
	}).Error)

	svc := NewGatewayCostService(&app.Deps{DB: db})
	first, err := svc.Snapshot(ctx, "tenant-a", GatewayCostQuery{WindowHours: 24, UnitPrice: 0.05, QuotaLimit: 1000})
	require.NoError(t, err)
	second, err := svc.Snapshot(ctx, "tenant-a", GatewayCostQuery{WindowHours: 24, UnitPrice: 0.05, QuotaLimit: 1000})
	require.NoError(t, err)
	require.Equal(t, first.Summary.TotalRequests, second.Summary.TotalRequests)
	require.InDelta(t, first.Summary.TotalCost, second.Summary.TotalCost, 0.0001)

	otherTenantCtx := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	other, err := svc.Snapshot(otherTenantCtx, "tenant-b", GatewayCostQuery{WindowHours: 24, UnitPrice: 0.05, QuotaLimit: 1000})
	require.NoError(t, err)
	require.Equal(t, 0, other.Summary.TotalRequests)
	require.InDelta(t, 0.0, other.Summary.TotalCost, 0.0001)
}
