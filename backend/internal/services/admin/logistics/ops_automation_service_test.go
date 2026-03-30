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

func TestOpsAutomationService_SuppressionAccuracy(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_ops_automation_suppress")
	ensureOpsAutomationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-ops-a")
	now := time.Now().UTC()

	for i := 0; i < 2; i++ {
		require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.GatewayFailureEvent{
			ID:           "evt-suppress-" + string(rune('a'+i)),
			TenantUUID:   "tenant-ops-a",
			CarrierID:    "carrier-a",
			ErrorClass:   "timeout",
			ErrorCode:    "E_TIMEOUT",
			ErrorMessage: "timeout",
			Status:       "pending",
			CreatedAt:    now,
			UpdatedAt:    now,
		}).Error)
	}

	svc := NewOpsAutomationService(&app.Deps{DB: db})
	_, err := svc.UpsertPolicy(ctx, "tenant-ops-a", UpsertOpsAutomationPolicyRequest{
		Name:            "抑制策略",
		CarrierID:       "carrier-a",
		SuppressionRule: map[string]any{"min_failed_events": 3},
		RetryStrategy:   map[string]any{"enabled": true, "max_attempts": 1},
		EscalationChain: map[string]any{"failed_event_threshold": 4, "slo_throttled_required": true},
		Enabled:         boolPtrOps(true),
	})
	require.NoError(t, err)

	snapshot, err := svc.Evaluate(ctx, "tenant-ops-a", EvaluateOpsAutomationRequest{CarrierID: "carrier-a"})
	require.NoError(t, err)
	require.Equal(t, 1, snapshot.SuppressionHits)
	require.Equal(t, 0, snapshot.Escalated)
	require.NotEmpty(t, snapshot.Runs)
	require.True(t, snapshot.Runs[0].Suppressed)
}

func TestOpsAutomationService_AutoRecoveryStability(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_ops_automation_recover")
	ensureOpsAutomationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-ops-b")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.GatewayFailureEvent{
		ID:           "evt-recover-1",
		TenantUUID:   "tenant-ops-b",
		CarrierID:    "carrier-b",
		ErrorClass:   "timeout",
		ErrorCode:    "E_TIMEOUT",
		ErrorMessage: "timeout",
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error)

	svc := NewOpsAutomationService(&app.Deps{DB: db})
	_, err := svc.UpsertPolicy(ctx, "tenant-ops-b", UpsertOpsAutomationPolicyRequest{
		Name:            "恢复策略",
		CarrierID:       "carrier-b",
		SuppressionRule: map[string]any{"min_failed_events": 1},
		RetryStrategy:   map[string]any{"enabled": true, "max_attempts": 1},
		EscalationChain: map[string]any{"failed_event_threshold": 9, "slo_throttled_required": false},
		Enabled:         boolPtrOps(true),
	})
	require.NoError(t, err)

	snapshot, err := svc.Evaluate(ctx, "tenant-ops-b", EvaluateOpsAutomationRequest{CarrierID: "carrier-b"})
	require.NoError(t, err)
	require.Equal(t, 1, snapshot.AutoRecovered)
	require.Len(t, snapshot.Runs, 1)
	require.True(t, snapshot.Runs[0].AutoRecovered)

	var row LogisticsModel.GatewayFailureEvent
	require.NoError(t, db.WithContext(ctx).Where("id = ?", "evt-recover-1").First(&row).Error)
	require.Equal(t, "recovered", row.Status)
	require.NotNil(t, row.RecoveredAt)
}

func TestOpsAutomationService_TakeoverPermissionControl(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_ops_automation_takeover")
	ensureOpsAutomationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-ops-c")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.GatewayFailureEvent{
		ID:           "evt-takeover-1",
		TenantUUID:   "tenant-ops-c",
		CarrierID:    "carrier-c",
		ErrorClass:   "server_5xx",
		ErrorCode:    "E_5XX",
		ErrorMessage: "5xx",
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}).Error)

	svc := NewOpsAutomationService(&app.Deps{DB: db})
	_, err := svc.UpsertPolicy(ctx, "tenant-ops-c", UpsertOpsAutomationPolicyRequest{
		Name:            "接管策略",
		CarrierID:       "carrier-c",
		SuppressionRule: map[string]any{"min_failed_events": 1},
		RetryStrategy:   map[string]any{"enabled": false, "max_attempts": 0},
		EscalationChain: map[string]any{"failed_event_threshold": 1, "slo_throttled_required": false},
		Enabled:         boolPtrOps(true),
	})
	require.NoError(t, err)

	snapshot, err := svc.Evaluate(ctx, "tenant-ops-c", EvaluateOpsAutomationRequest{CarrierID: "carrier-c"})
	require.NoError(t, err)
	require.NotEmpty(t, snapshot.Runs)
	runID := snapshot.Runs[0].ID

	_, err = svc.Takeover(ctx, "tenant-ops-c", runID, OpsAutomationTakeoverRequest{
		Action:     "manual_takeover",
		OperatorID: "",
		Reason:     "no operator",
	})
	require.Error(t, err)

	run, err := svc.Takeover(ctx, "tenant-ops-c", runID, OpsAutomationTakeoverRequest{
		Action:     "manual_takeover",
		OperatorID: "ops-admin",
		Reason:     "manual intervention",
	})
	require.NoError(t, err)
	require.Equal(t, "manual_takeover", run.Status)
	require.Equal(t, "ops-admin", run.TakeoverBy)
}

func ensureOpsAutomationTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	ensureSLOGuardTables(t, db)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_ops_automation_policies (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		carrier_id TEXT,
		retry_strategy JSON,
		circuit_breaker_strategy JSON,
		suppression_rule JSON,
		escalation_chain JSON,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		last_evaluated_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_ops_automation_policy ON logistics_ops_automation_policies(tenant_uuid, name)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_ops_automation_runs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		policy_id TEXT NOT NULL,
		carrier_id TEXT,
		trigger_source TEXT NOT NULL,
		trigger_key TEXT,
		status TEXT NOT NULL,
		suppressed BOOLEAN NOT NULL DEFAULT 0,
		auto_recovered BOOLEAN NOT NULL DEFAULT 0,
		escalated BOOLEAN NOT NULL DEFAULT 0,
		takeover_by TEXT,
		takeover_reason TEXT,
		payload JSON,
		started_at DATETIME,
		finished_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
}

func boolPtrOps(v bool) *bool {
	return &v
}
