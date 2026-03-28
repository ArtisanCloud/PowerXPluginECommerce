package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAllocationService_AutoAllocateAndOverride(t *testing.T) {
	db := setupBillingDB(t, "logistics_allocation_auto")
	ensureAllocationTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-alloc")
	seedCarrier(t, db, "tenant-alloc", "carrier-aa", "Carrier-AA")
	seedCarrier(t, db, "tenant-alloc", "carrier-bb", "Carrier-BB")

	svc := NewAllocationService(&app.Deps{DB: db})

	_, err := svc.UpsertPlan(ctx, "tenant-alloc", UpsertCapacityPlanRequest{
		Name:             "A计划",
		CarrierID:        "carrier-aa",
		WarehouseID:      "wh-1",
		DestinationZone:  "CN-EAST",
		DailyCapacity:    20,
		ReservedCapacity: 5,
		Status:           "active",
	})
	require.NoError(t, err)
	_, err = svc.UpsertPlan(ctx, "tenant-alloc", UpsertCapacityPlanRequest{
		Name:             "B计划",
		CarrierID:        "carrier-bb",
		WarehouseID:      "wh-1",
		DestinationZone:  "CN-EAST",
		DailyCapacity:    10,
		ReservedCapacity: 8,
		Status:           "active",
	})
	require.NoError(t, err)

	result, err := svc.Allocate(ctx, "tenant-alloc", AllocationRequest{
		RequestKey:      "alloc#1",
		OrderID:         "order-1",
		WarehouseID:     "wh-1",
		DestinationZone: "CN-EAST",
		OperatorID:      "admin",
	})
	require.NoError(t, err)
	require.Equal(t, "carrier-aa", result.CarrierID)
	require.False(t, result.ManualOverride)
	require.NotEmpty(t, result.Candidates)

	override, err := svc.Override(ctx, "tenant-alloc", OverrideAllocationRequest{
		RequestKey: "alloc#1",
		CarrierID:  "carrier-bb",
		OperatorID: "ops-1",
		Reason:     "manual assign",
	})
	require.NoError(t, err)
	require.Equal(t, "carrier-bb", override.CarrierID)
	require.True(t, override.ManualOverride)
}

func TestAllocationService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_allocation_tenant")
	ensureAllocationTables(t, db)
	svc := NewAllocationService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	seedCarrier(t, db, "tenant-a", "carrier-a", "Carrier-A")

	_, err := svc.UpsertPlan(ctxA, "tenant-a", UpsertCapacityPlanRequest{
		Name:          "A计划",
		CarrierID:     "carrier-a",
		DailyCapacity: 10,
		Status:        "active",
	})
	require.NoError(t, err)
	plansB, err := svc.ListPlans(ctxB, "tenant-b", AllocationQuery{})
	require.NoError(t, err)
	require.Len(t, plansB, 0)
}

func ensureAllocationTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_capacity_plans (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		warehouse_id TEXT,
		destination_zone TEXT,
		daily_capacity INTEGER NOT NULL DEFAULT 0,
		reserved_capacity INTEGER NOT NULL DEFAULT 0,
		used_capacity INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'active',
		config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_capacity_plan ON logistics_capacity_plans(tenant_uuid, name)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_allocation_decisions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		waybill_id TEXT,
		order_id TEXT,
		carrier_id TEXT NOT NULL,
		warehouse_id TEXT,
		destination_zone TEXT,
		strategy TEXT NOT NULL DEFAULT 'capacity_first',
		reason TEXT,
		manual_override BOOLEAN NOT NULL DEFAULT 0,
		previous_carrier_id TEXT,
		candidates JSON,
		metadata JSON,
		created_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_allocation_request ON logistics_allocation_decisions(tenant_uuid, request_key)`).Error)
}
