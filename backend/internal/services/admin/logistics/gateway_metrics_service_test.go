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

func TestGatewayMetricsService_SnapshotAndAlerts(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "gateway_metrics_snapshot")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	now := time.Now().UTC()
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-1",
		TenantUUID:    "tenant-a",
		CarrierID:     "carrier-a",
		Status:        "success",
		BatchLimit:    20,
		EventLimit:    20,
		TotalWaybills: 20,
		SuccessCount:  20,
		FailedCount:   0,
		P95LatencyMS:  300,
		CreatedAt:     now.Add(-1 * time.Hour),
		UpdatedAt:     now.Add(-1 * time.Hour),
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-2",
		TenantUUID:    "tenant-a",
		CarrierID:     "carrier-a",
		Status:        "partial_failed",
		BatchLimit:    20,
		EventLimit:    20,
		TotalWaybills: 10,
		SuccessCount:  6,
		FailedCount:   4,
		P95LatencyMS:  2600,
		CreatedAt:     now.Add(-30 * time.Minute),
		UpdatedAt:     now.Add(-30 * time.Minute),
	}).Error)

	svc := NewGatewayMetricsService(&app.Deps{DB: db})
	snapshot, err := svc.Snapshot(ctx, "tenant-a", 24)
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	require.Equal(t, 30, snapshot.Summary.TotalRequests)
	require.Equal(t, 26, snapshot.Summary.SuccessRequests)
	require.Equal(t, 4, snapshot.Summary.FailedRequests)
	require.Greater(t, snapshot.Summary.SuccessRate, 80.0)
	require.Equal(t, 2600, snapshot.Summary.P95LatencyMS)
	require.NotEmpty(t, snapshot.Alerts)
}
