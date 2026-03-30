package logistics

import (
	"context"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSLOGuardService_ThresholdTriggering(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_slo_guard_trigger")
	ensureSLOGuardTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-slo-a")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-slo-trigger-1",
		TenantUUID:    "tenant-slo-a",
		CarrierID:     "carrier-a",
		Status:        "partial_failed",
		TotalWaybills: 100,
		SuccessCount:  80,
		FailedCount:   20,
		P95LatencyMS:  2500,
		CreatedAt:     now.Add(-10 * time.Minute),
		UpdatedAt:     now.Add(-10 * time.Minute),
	}).Error)

	svc := NewSLOGuardService(&app.Deps{DB: db})
	_, err := svc.UpsertPolicy(ctx, "tenant-slo-a", UpsertSLOGuardPolicyRequest{
		Name:              "网关SLO守卫",
		CarrierID:         "carrier-a",
		WindowHours:       2,
		MinSuccessRate:    95,
		MaxP95LatencyMS:   2000,
		MaxFailedRequests: 10,
		Action:            "throttle",
		ThrottleRatio:     60,
	})
	require.NoError(t, err)

	result, err := svc.Evaluate(ctx, "tenant-slo-a", SLOGuardEvaluateRequest{WindowHours: 2, CarrierID: "carrier-a"})
	require.NoError(t, err)
	require.True(t, result.ThrottleRequired)
	require.Equal(t, 1, result.ThrottledPolicies)
	require.NotEmpty(t, result.Triggered)
	require.Equal(t, "threshold_breached", result.Triggered[0].ReasonCode)
}

func TestSLOGuardService_AutoRecoverAndManualRelease(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_slo_guard_recover")
	ensureSLOGuardTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-slo-b")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-slo-recover-1",
		TenantUUID:    "tenant-slo-b",
		CarrierID:     "carrier-b",
		Status:        "failed",
		TotalWaybills: 20,
		SuccessCount:  8,
		FailedCount:   12,
		P95LatencyMS:  3200,
		CreatedAt:     now.Add(-2 * time.Hour),
		UpdatedAt:     now.Add(-2 * time.Hour),
	}).Error)

	svc := NewSLOGuardService(&app.Deps{DB: db})
	policy, err := svc.UpsertPolicy(ctx, "tenant-slo-b", UpsertSLOGuardPolicyRequest{
		Name:              "延迟守卫",
		CarrierID:         "carrier-b",
		WindowHours:       3,
		MinSuccessRate:    90,
		MaxP95LatencyMS:   2500,
		MaxFailedRequests: 5,
		Action:            "throttle",
		ThrottleRatio:     40,
	})
	require.NoError(t, err)

	first, err := svc.Evaluate(ctx, "tenant-slo-b", SLOGuardEvaluateRequest{WindowHours: 3, CarrierID: "carrier-b"})
	require.NoError(t, err)
	require.True(t, first.ThrottleRequired)
	require.Equal(t, 1, first.ThrottledPolicies)

	released, err := svc.ManualRelease(ctx, "tenant-slo-b", SLOGuardReleaseRequest{
		PolicyID:   policy.ID,
		OperatorID: "ops-01",
		Reason:     "确认误报，先放行",
	})
	require.NoError(t, err)
	require.Equal(t, "released", released.Status)

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.TrackingSyncJob{
		ID:            "job-slo-recover-2",
		TenantUUID:    "tenant-slo-b",
		CarrierID:     "carrier-b",
		Status:        "success",
		TotalWaybills: 120,
		SuccessCount:  120,
		FailedCount:   0,
		P95LatencyMS:  800,
		CreatedAt:     now.Add(-10 * time.Minute),
		UpdatedAt:     now.Add(-10 * time.Minute),
	}).Error)

	second, err := svc.Evaluate(ctx, "tenant-slo-b", SLOGuardEvaluateRequest{WindowHours: 1, CarrierID: "carrier-b"})
	require.NoError(t, err)
	require.False(t, second.ThrottleRequired)
	require.Equal(t, 0, second.ThrottledPolicies)
}

func TestSLOGuardService_TenantIsolation(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_slo_guard_isolation")
	ensureSLOGuardTables(t, db)
	svc := NewSLOGuardService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-slo-a")
	_, err := svc.UpsertPolicy(ctxA, "tenant-slo-a", UpsertSLOGuardPolicyRequest{
		Name:              "tenant-a-policy",
		WindowHours:       24,
		MinSuccessRate:    95,
		MaxP95LatencyMS:   2000,
		MaxFailedRequests: 10,
		Action:            "throttle",
		ThrottleRatio:     50,
	})
	require.NoError(t, err)

	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-slo-b")
	policiesB, err := svc.ListPolicies(ctxB, "tenant-slo-b", "", nil, 20)
	require.NoError(t, err)
	require.Len(t, policiesB, 0)
}

func ensureSLOGuardTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_slo_guard_policies (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		carrier_id TEXT,
		window_hours INTEGER NOT NULL DEFAULT 24,
		min_success_rate NUMERIC NOT NULL DEFAULT 95,
		max_p95_latency_ms INTEGER NOT NULL DEFAULT 2000,
		max_failed_requests INTEGER NOT NULL DEFAULT 10,
		action TEXT NOT NULL DEFAULT 'throttle',
		throttle_ratio INTEGER NOT NULL DEFAULT 50,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_slo_guard_policy ON logistics_slo_guard_policies(tenant_uuid, name)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_slo_guard_states (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		policy_id TEXT NOT NULL,
		carrier_id TEXT,
		status TEXT NOT NULL DEFAULT 'normal',
		action TEXT NOT NULL DEFAULT 'throttle',
		throttle_ratio INTEGER NOT NULL DEFAULT 0,
		reason_code TEXT,
		reason_message TEXT,
		trigger_metrics JSON,
		activated_at DATETIME,
		recovered_at DATETIME,
		manual_released_by TEXT,
		manual_reason TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_logistics_slo_guard_state_policy ON logistics_slo_guard_states(tenant_uuid, policy_id, created_at)`).Error)
}
