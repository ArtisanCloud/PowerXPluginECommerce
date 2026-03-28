package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRoutingOptimizerService_StableScoreOrder(t *testing.T) {
	db := setupBillingDB(t, "logistics_routing_optimizer_stable")
	ensureRoutingOptimizerTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-optimizer")
	seedCarrier(t, db, "tenant-optimizer", "carrier-a", "Carrier-A")
	seedCarrier(t, db, "tenant-optimizer", "carrier-b", "Carrier-B")
	require.NoError(t, db.WithContext(ctx).Exec(`UPDATE logistics_carriers SET config = ? WHERE id = ?`, `{"onTimeRate":98,"cost_score":0.6,"quota_remain":0.7,"risk_score":0.2}`, "carrier-a").Error)
	require.NoError(t, db.WithContext(ctx).Exec(`UPDATE logistics_carriers SET config = ? WHERE id = ?`, `{"onTimeRate":92,"cost_score":0.3,"quota_remain":0.5,"risk_score":0.4}`, "carrier-b").Error)
	svc := NewRoutingOptimizerService(&app.Deps{DB: db})

	_, err := svc.UpsertStrategy(ctx, "tenant-optimizer", UpsertRoutingOptimizerStrategyRequest{
		Name:             "default",
		TimelinessWeight: 0.4,
		CostWeight:       0.3,
		QuotaWeight:      0.2,
		RiskWeight:       0.1,
	})
	require.NoError(t, err)

	first, err := svc.Simulate(ctx, "tenant-optimizer", RoutingOptimizerSimulationRequest{})
	require.NoError(t, err)
	second, err := svc.Simulate(ctx, "tenant-optimizer", RoutingOptimizerSimulationRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, first.Candidates)
	require.Equal(t, first.Candidates[0].CarrierID, second.Candidates[0].CarrierID)
}

func TestRoutingOptimizerService_DegradeFallback(t *testing.T) {
	db := setupBillingDB(t, "logistics_routing_optimizer_degrade")
	ensureRoutingOptimizerTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-optimizer")
	seedCarrier(t, db, "tenant-optimizer", "carrier-a", "Carrier-A")
	svc := NewRoutingOptimizerService(&app.Deps{DB: db})

	_, err := svc.UpsertStrategy(ctx, "tenant-optimizer", UpsertRoutingOptimizerStrategyRequest{FallbackStrategy: "highest_timeliness"})
	require.NoError(t, err)
	result, err := svc.Simulate(ctx, "tenant-optimizer", RoutingOptimizerSimulationRequest{RealtimeDegraded: true})
	require.NoError(t, err)
	require.True(t, result.Degraded)
	require.Equal(t, "degraded", result.Strategy)
}

func TestRoutingOptimizerService_InvalidWeightsIsolated(t *testing.T) {
	db := setupBillingDB(t, "logistics_routing_optimizer_invalid")
	ensureRoutingOptimizerTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-optimizer")
	seedCarrier(t, db, "tenant-optimizer", "carrier-a", "Carrier-A")
	svc := NewRoutingOptimizerService(&app.Deps{DB: db})

	row, err := svc.UpsertStrategy(ctx, "tenant-optimizer", UpsertRoutingOptimizerStrategyRequest{
		TimelinessWeight: -1,
		CostWeight:       -1,
		QuotaWeight:      -1,
		RiskWeight:       -1,
	})
	require.NoError(t, err)
	require.InDelta(t, 0.4, row.TimelinessWeight, 0.0001)
	require.InDelta(t, 0.3, row.CostWeight, 0.0001)

	_, err = svc.Simulate(ctx, "tenant-optimizer", RoutingOptimizerSimulationRequest{AvailableCarrierIDs: []string{"missing"}})
	require.Error(t, err)
}

func ensureRoutingOptimizerTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_routing_score_profiles (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		timeliness_weight NUMERIC NOT NULL DEFAULT 0.4,
		cost_weight NUMERIC NOT NULL DEFAULT 0.3,
		quota_weight NUMERIC NOT NULL DEFAULT 0.2,
		risk_weight NUMERIC NOT NULL DEFAULT 0.1,
		fallback_strategy TEXT NOT NULL DEFAULT 'highest_timeliness',
		enabled BOOLEAN NOT NULL DEFAULT 1,
		config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_routing_score_profile ON logistics_routing_score_profiles(tenant_uuid, name)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_routing_score_simulations (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		profile_id TEXT,
		preferred_carrier_id TEXT,
		chosen_carrier_id TEXT,
		chosen_carrier_name TEXT,
		strategy TEXT NOT NULL,
		degraded BOOLEAN NOT NULL DEFAULT 0,
		reason TEXT,
		input_payload JSON,
		candidates JSON,
		explain_payload JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
}
