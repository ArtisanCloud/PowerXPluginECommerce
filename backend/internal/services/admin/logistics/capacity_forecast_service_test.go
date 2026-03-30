package logistics

import (
	"context"
	"fmt"
	"testing"
	"time"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCapacityForecastService_ForecastConsistency(t *testing.T) {
	db := setupBillingDB(t, "logistics_capacity_forecast_consistency")
	ensureCapacityForecastTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-cf-a")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.CapacityPlan{
		ID:               "plan-cf-1",
		TenantUUID:       "tenant-cf-a",
		Name:             "预测计划A",
		CarrierID:        "carrier-cf-a",
		WarehouseID:      "wh-a",
		DestinationZone:  "CN-EAST",
		DailyCapacity:    100,
		ReservedCapacity: 20,
		UsedCapacity:     40,
		Status:           "active",
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error)

	for i := 0; i < 14; i++ {
		require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.AllocationDecision{
			ID:              fmt.Sprintf("decision-cf-%d", i+1),
			TenantUUID:      "tenant-cf-a",
			RequestKey:      fmt.Sprintf("req-cf-%d", i+1),
			CarrierID:       "carrier-cf-a",
			WarehouseID:     "wh-a",
			DestinationZone: "CN-EAST",
			Strategy:        "capacity_first",
			CreatedAt:       now,
			UpdatedAt:       now,
		}).Error)
	}

	svc := NewCapacityForecastService(&app.Deps{DB: db})
	rows, err := svc.Generate(ctx, "tenant-cf-a", CapacityForecastGenerateRequest{
		CarrierID:       "carrier-cf-a",
		WarehouseID:     "wh-a",
		DestinationZone: "CN-EAST",
		WindowDays:      7,
		OperatorID:      "ops-a",
	})
	require.NoError(t, err)
	require.Len(t, rows, 1)

	row := rows[0]
	require.Equal(t, 100, row.CurrentDailyCapacity)
	require.Equal(t, 65, row.PredictedDailyVolume)
	require.Equal(t, 75, row.TargetCapacity)
	require.Equal(t, 75, row.RecommendedQuota)
	require.Equal(t, "low", row.RiskLevel)
	require.Equal(t, "keep_quota", row.Strategy)
	require.Equal(t, "suggested", row.Status)
	require.InDelta(t, 69.0, row.Confidence, 0.01)

	listed, err := svc.List(ctx, "tenant-cf-a", CapacityForecastQuery{Status: "suggested", Limit: 10})
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, row.ID, listed[0].ID)
}

func TestCapacityForecastService_ApplyIdempotent(t *testing.T) {
	db := setupBillingDB(t, "logistics_capacity_forecast_apply")
	ensureCapacityForecastTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-cf-b")
	now := time.Now().UTC()

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.CapacityPlan{
		ID:               "plan-cf-2",
		TenantUUID:       "tenant-cf-b",
		Name:             "预测计划B",
		CarrierID:        "carrier-cf-b",
		WarehouseID:      "wh-b",
		DestinationZone:  "CN-SOUTH",
		DailyCapacity:    80,
		ReservedCapacity: 30,
		UsedCapacity:     20,
		Status:           "active",
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.CapacityForecast{
		ID:                   "forecast-cf-1",
		TenantUUID:           "tenant-cf-b",
		PlanID:               "plan-cf-2",
		CarrierID:            "carrier-cf-b",
		WarehouseID:          "wh-b",
		DestinationZone:      "CN-SOUTH",
		WindowDays:           14,
		CurrentDailyCapacity: 80,
		PredictedDailyVolume: 95,
		TargetCapacity:       120,
		RecommendedQuota:     120,
		Confidence:           82,
		RiskLevel:            "high",
		Strategy:             "increase_quota",
		Status:               "suggested",
		CreatedAt:            now,
		UpdatedAt:            now,
	}).Error)

	svc := NewCapacityForecastService(&app.Deps{DB: db})
	first, err := svc.Apply(ctx, "tenant-cf-b", CapacityForecastApplyRequest{
		ForecastID: "forecast-cf-1",
		OperatorID: "ops-b",
	})
	require.NoError(t, err)
	require.Equal(t, "applied", first.Status)
	require.NotNil(t, first.AppliedAt)

	var plan LogisticsModel.CapacityPlan
	require.NoError(t, db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", "tenant-cf-b", "plan-cf-2").First(&plan).Error)
	require.Equal(t, 120, plan.DailyCapacity)
	require.Equal(t, 120, plan.ReservedCapacity)

	second, err := svc.Apply(ctx, "tenant-cf-b", CapacityForecastApplyRequest{
		ForecastID: "forecast-cf-1",
		OperatorID: "ops-b",
	})
	require.NoError(t, err)
	require.Equal(t, "applied", second.Status)
	require.NotNil(t, second.AppliedAt)

	var planReloaded LogisticsModel.CapacityPlan
	require.NoError(t, db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", "tenant-cf-b", "plan-cf-2").First(&planReloaded).Error)
	require.Equal(t, 120, planReloaded.DailyCapacity)
	require.Equal(t, 120, planReloaded.ReservedCapacity)
}

func TestCapacityForecastService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_capacity_forecast_tenant")
	ensureCapacityForecastTables(t, db)
	svc := NewCapacityForecastService(&app.Deps{DB: db})
	now := time.Now().UTC()

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-cf-x")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-cf-y")

	require.NoError(t, db.WithContext(ctxA).Create(&LogisticsModel.CapacityPlan{
		ID:               "plan-cf-x",
		TenantUUID:       "tenant-cf-x",
		Name:             "预测计划X",
		CarrierID:        "carrier-cf-x",
		DailyCapacity:    50,
		ReservedCapacity: 10,
		UsedCapacity:     5,
		Status:           "active",
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error)
	require.NoError(t, db.WithContext(ctxA).Create(&LogisticsModel.CapacityForecast{
		ID:                   "forecast-cf-x",
		TenantUUID:           "tenant-cf-x",
		PlanID:               "plan-cf-x",
		CarrierID:            "carrier-cf-x",
		WindowDays:           7,
		CurrentDailyCapacity: 50,
		PredictedDailyVolume: 40,
		TargetCapacity:       45,
		RecommendedQuota:     45,
		Confidence:           72,
		RiskLevel:            "medium",
		Strategy:             "rebalance_quota",
		Status:               "suggested",
		CreatedAt:            now,
		UpdatedAt:            now,
	}).Error)

	listB, err := svc.List(ctxB, "tenant-cf-y", CapacityForecastQuery{Limit: 20})
	require.NoError(t, err)
	require.Len(t, listB, 0)

	_, err = svc.Apply(ctxB, "tenant-cf-y", CapacityForecastApplyRequest{ForecastID: "forecast-cf-x"})
	require.Error(t, err)
}

func ensureCapacityForecastTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	ensureAllocationTables(t, db)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_capacity_forecasts (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		plan_id TEXT,
		carrier_id TEXT,
		warehouse_id TEXT,
		destination_zone TEXT,
		window_days INTEGER NOT NULL DEFAULT 7,
		current_daily_capacity INTEGER NOT NULL DEFAULT 0,
		predicted_daily_volume INTEGER NOT NULL DEFAULT 0,
		target_capacity INTEGER NOT NULL DEFAULT 0,
		recommended_quota INTEGER NOT NULL DEFAULT 0,
		confidence NUMERIC NOT NULL DEFAULT 0,
		risk_level TEXT NOT NULL DEFAULT 'low',
		strategy TEXT NOT NULL DEFAULT 'keep_quota',
		status TEXT NOT NULL DEFAULT 'suggested',
		metrics JSON,
		applied_at DATETIME,
		created_by TEXT,
		updated_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_logistics_capacity_forecast_scope ON logistics_capacity_forecasts(tenant_uuid, carrier_id, warehouse_id, destination_zone)`).Error)
}
