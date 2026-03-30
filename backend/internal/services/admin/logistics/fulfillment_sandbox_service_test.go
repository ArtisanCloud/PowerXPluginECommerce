package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestFulfillmentSandboxService_ReproducibleRun(t *testing.T) {
	db := setupBillingDB(t, "logistics_sandbox_reproducible")
	ensureFulfillmentSandboxTables(t, db)
	svc := NewFulfillmentSandboxService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-sbx-a")

	scenario, err := svc.UpsertScenario(ctx, "tenant-sbx-a", UpsertFulfillmentSandboxScenarioRequest{
		Name: "大促压力测试",
		BaselineConfig: map[string]any{
			"timeliness_rate": 92.5,
			"cost_index":      1.08,
			"exception_rate":  4.2,
			"recovery_hours":  14,
		},
	})
	require.NoError(t, err)

	runA, err := svc.Run(ctx, "tenant-sbx-a", RunFulfillmentSandboxRequest{
		ScenarioID: scenario.ID,
		WindowDays: 14,
		Strategy:   "balanced",
	})
	require.NoError(t, err)

	runB, err := svc.Run(ctx, "tenant-sbx-a", RunFulfillmentSandboxRequest{
		ScenarioID: scenario.ID,
		WindowDays: 14,
		Strategy:   "balanced",
	})
	require.NoError(t, err)

	require.Equal(t, runA.ID, runB.ID)
	require.Equal(t, runA.Score, runB.Score)
	require.Equal(t, runA.TimelinessRate, runB.TimelinessRate)
	require.Equal(t, runA.CostIndex, runB.CostIndex)
}

func TestFulfillmentSandboxService_StrategyDiffExplainable(t *testing.T) {
	db := setupBillingDB(t, "logistics_sandbox_strategy_diff")
	ensureFulfillmentSandboxTables(t, db)
	svc := NewFulfillmentSandboxService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-sbx-b")

	scenario, err := svc.UpsertScenario(ctx, "tenant-sbx-b", UpsertFulfillmentSandboxScenarioRequest{
		Name: "策略对比样本",
		BaselineConfig: map[string]any{
			"timeliness_rate": 94.0,
			"cost_index":      1.0,
			"exception_rate":  3.0,
			"recovery_hours":  10.0,
		},
	})
	require.NoError(t, err)

	baseRun, err := svc.Run(ctx, "tenant-sbx-b", RunFulfillmentSandboxRequest{
		ScenarioID: scenario.ID,
		Strategy:   "balanced",
		RequestKey: "sandbox#base",
	})
	require.NoError(t, err)
	costRun, err := svc.Run(ctx, "tenant-sbx-b", RunFulfillmentSandboxRequest{
		ScenarioID: scenario.ID,
		Strategy:   "cost_first",
		RequestKey: "sandbox#cost",
	})
	require.NoError(t, err)

	compare, err := svc.Compare(ctx, "tenant-sbx-b", CompareFulfillmentSandboxRequest{
		BaselineRunID:  baseRun.ID,
		CandidateRunID: costRun.ID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, compare.Recommendation)
	require.Less(t, compare.CostDelta, 0.0)
}

func TestFulfillmentSandboxService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_sandbox_tenant")
	ensureFulfillmentSandboxTables(t, db)
	svc := NewFulfillmentSandboxService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-sbx-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-sbx-b")

	_, err := svc.UpsertScenario(ctxA, "tenant-sbx-a", UpsertFulfillmentSandboxScenarioRequest{
		Name: "租户A场景",
	})
	require.NoError(t, err)

	rowsB, err := svc.ListScenarios(ctxB, "tenant-sbx-b", FulfillmentSandboxScenarioQuery{Limit: 20})
	require.NoError(t, err)
	require.Len(t, rowsB, 0)
}

func ensureFulfillmentSandboxTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_fulfillment_sandbox_scenarios (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		carrier_id TEXT,
		warehouse_id TEXT,
		destination_zone TEXT,
		baseline_config JSON,
		strategy_config JSON,
		status TEXT NOT NULL DEFAULT 'active',
		description TEXT,
		created_by TEXT,
		updated_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_sandbox_scenario ON logistics_fulfillment_sandbox_scenarios(tenant_uuid, name)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_fulfillment_sandbox_runs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		scenario_id TEXT NOT NULL,
		request_key TEXT NOT NULL,
		window_days INTEGER NOT NULL DEFAULT 7,
		strategy TEXT NOT NULL DEFAULT 'balanced',
		timeliness_rate NUMERIC NOT NULL DEFAULT 0,
		cost_index NUMERIC NOT NULL DEFAULT 0,
		exception_rate NUMERIC NOT NULL DEFAULT 0,
		recovery_hours NUMERIC NOT NULL DEFAULT 0,
		score NUMERIC NOT NULL DEFAULT 0,
		snapshot JSON,
		recommendation TEXT,
		status TEXT NOT NULL DEFAULT 'completed',
		created_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_sandbox_run_request ON logistics_fulfillment_sandbox_runs(tenant_uuid, request_key)`).Error)
}
