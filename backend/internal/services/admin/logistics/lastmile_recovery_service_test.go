package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLastmileRecoveryService_IdempotencyRetryTakeover(t *testing.T) {
	db := setupBillingDB(t, "logistics_lastmile_recovery")
	ensureLastmileRecoveryTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-lastmile")
	svc := NewLastmileRecoveryService(&app.Deps{DB: db})

	rule, err := svc.UpsertRule(ctx, "tenant-lastmile", UpsertLastmileRecoveryRuleRequest{
		Name:         "延误改派",
		TriggerEvent: "delay",
		Action:       "redispatch",
		MaxRetries:   3,
		Enabled:      boolPtr(true),
	})
	require.NoError(t, err)
	require.Equal(t, "delay", rule.TriggerEvent)

	run, idem, err := svc.Execute(ctx, "tenant-lastmile", ExecuteLastmileRecoveryRequest{
		RequestKey:   "lm#1",
		RuleID:       rule.ID,
		WaybillID:    "wb-1",
		WaybillNo:    "WB-1",
		TriggerEvent: "lost",
	})
	require.NoError(t, err)
	require.Equal(t, "created", idem)
	require.Equal(t, "retrying", run.Status)
	require.Equal(t, 1, run.RetryCount)

	replay, idem, err := svc.Execute(ctx, "tenant-lastmile", ExecuteLastmileRecoveryRequest{
		RequestKey: "lm#1",
		RuleID:     rule.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", idem)
	require.Equal(t, run.ID, replay.ID)

	taken, err := svc.Takeover(ctx, "tenant-lastmile", run.ID, TakeoverLastmileRecoveryRequest{
		Action:     "takeover",
		OperatorID: "ops-1",
		Reason:     "manual check",
	})
	require.NoError(t, err)
	require.Equal(t, "manual_taken", taken.Status)
	require.True(t, taken.ManualTaken)

	resumed, err := svc.Takeover(ctx, "tenant-lastmile", run.ID, TakeoverLastmileRecoveryRequest{
		Action:     "resume",
		OperatorID: "ops-2",
	})
	require.NoError(t, err)
	require.Equal(t, "retrying", resumed.Status)
	require.False(t, resumed.ManualTaken)
}

func TestLastmileRecoveryService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_lastmile_recovery_tenant")
	ensureLastmileRecoveryTables(t, db)
	svc := NewLastmileRecoveryService(&app.Deps{DB: db})
	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-b")

	rule, err := svc.UpsertRule(ctxA, "tenant-a", UpsertLastmileRecoveryRuleRequest{
		Name:         "A规则",
		TriggerEvent: "delay",
		Action:       "redispatch",
		Enabled:      boolPtr(true),
	})
	require.NoError(t, err)

	_, _, err = svc.Execute(ctxA, "tenant-a", ExecuteLastmileRecoveryRequest{
		RequestKey: "lm#a",
		RuleID:     rule.ID,
		WaybillNo:  "WB-A",
	})
	require.NoError(t, err)

	rulesB, err := svc.ListRules(ctxB, "tenant-b", nil)
	require.NoError(t, err)
	require.Len(t, rulesB, 0)
	runsB, err := svc.ListRuns(ctxB, "tenant-b", "", "", 20)
	require.NoError(t, err)
	require.Len(t, runsB, 0)
}

func ensureLastmileRecoveryTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_lastmile_recovery_rules (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		trigger_event TEXT NOT NULL,
		action TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 100,
		max_retries INTEGER NOT NULL DEFAULT 3,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_lastmile_recovery_rule ON logistics_lastmile_recovery_rules(tenant_uuid, name)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_lastmile_recovery_runs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		rule_id TEXT NOT NULL,
		waybill_id TEXT,
		waybill_no TEXT,
		trigger_event TEXT NOT NULL,
		action TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		retry_count INTEGER NOT NULL DEFAULT 0,
		max_retries INTEGER NOT NULL DEFAULT 3,
		message TEXT,
		manual_taken BOOLEAN NOT NULL DEFAULT 0,
		taken_by TEXT,
		taken_reason TEXT,
		taken_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_lastmile_recovery_run ON logistics_lastmile_recovery_runs(tenant_uuid, request_key)`).Error)
}
