package logistics

import (
	"context"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestControlTowerService_OverviewAndDrilldown(t *testing.T) {
	db := setupBillingDB(t, "logistics_control_tower_overview")
	ensureControlTowerTables(t, db)

	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-control")
	seedCarrier(t, db, "tenant-control", "carrier-ct-1", "Carrier-CT-1")
	seedCarrier(t, db, "tenant-control", "carrier-ct-2", "Carrier-CT-2")

	waybillSvc := NewWaybillService(&app.Deps{DB: db})
	svc := NewControlTowerService(&app.Deps{DB: db})

	wb1, _, err := waybillSvc.Create(ctx, "tenant-control", CreateWaybillRequest{
		OrderID:        "order-ct-1",
		CarrierID:      "carrier-ct-1",
		ServiceCode:    "std",
		WaybillNo:      "WB-CT-1",
		PackageKey:     "order-ct-1#1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb1.Status = "in_transit"
	wb1.FeeAmount = 12
	wb1.Metadata = datatypes.JSON([]byte(`{"warehouse_id":"wh-a","destination_zone":"CN-EAST"}`))
	wb1.CreatedAt = time.Now().UTC().Add(-12 * time.Hour)
	wb1.UpdatedAt = time.Now().UTC().Add(-1 * time.Hour)
	require.NoError(t, db.WithContext(ctx).Save(wb1).Error)

	wb2, _, err := waybillSvc.Create(ctx, "tenant-control", CreateWaybillRequest{
		OrderID:        "order-ct-2",
		CarrierID:      "carrier-ct-1",
		ServiceCode:    "std",
		WaybillNo:      "WB-CT-2",
		PackageKey:     "order-ct-2#1",
		ShipmentItems:  []string{"sku-2"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb2.Status = "delay"
	wb2.FeeAmount = 18
	wb2.Metadata = datatypes.JSON([]byte(`{"warehouse_id":"wh-a","destination_zone":"CN-EAST"}`))
	wb2.CreatedAt = time.Now().UTC().Add(-90 * time.Hour)
	wb2.UpdatedAt = time.Now().UTC().Add(-2 * time.Hour)
	require.NoError(t, db.WithContext(ctx).Save(wb2).Error)

	wb3, _, err := waybillSvc.Create(ctx, "tenant-control", CreateWaybillRequest{
		OrderID:        "order-ct-3",
		CarrierID:      "carrier-ct-2",
		ServiceCode:    "std",
		WaybillNo:      "WB-CT-3",
		PackageKey:     "order-ct-3#1",
		ShipmentItems:  []string{"sku-3"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb3.Status = "delivered"
	wb3.ActualFeeAmount = 30
	wb3.Metadata = datatypes.JSON([]byte(`{"warehouse_id":"wh-b","destination_zone":"CN-WEST"}`))
	wb3.CreatedAt = time.Now().UTC().Add(-20 * time.Hour)
	wb3.UpdatedAt = time.Now().UTC().Add(-10 * time.Hour)
	require.NoError(t, db.WithContext(ctx).Save(wb3).Error)

	overview, err := svc.Overview(ctx, "tenant-control", ControlTowerOverviewQuery{
		WarehouseID: "wh-a",
		WindowHours: 120,
	})
	require.NoError(t, err)
	require.Equal(t, 2, overview.Summary.TotalWaybills)
	require.Equal(t, 1, overview.Summary.InTransitCount)
	require.Equal(t, 1, overview.Summary.ExceptionCount)
	require.Equal(t, 1, overview.Summary.TimeoutCount)

	drilldown, err := svc.Drilldown(ctx, "tenant-control", ControlTowerDrilldownQuery{
		WarehouseID: "wh-a",
		Status:      "exception",
		WindowHours: 120,
	})
	require.NoError(t, err)
	require.Len(t, drilldown, 1)
	require.Equal(t, "WB-CT-2", drilldown[0].WaybillNo)
}

func TestControlTowerService_SubscriptionsAndTenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_control_tower_subscriptions")
	ensureControlTowerTables(t, db)
	svc := NewControlTowerService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-cta")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-ctb")

	row, err := svc.UpsertSubscription(ctxA, "tenant-cta", UpsertControlTowerSubscriptionRequest{
		Name:            "主控告警",
		MinOnTimeRate:   96,
		MaxTimeoutCount: 3,
		MaxCostAmount:   2000,
		Enabled:         ptrBool(true),
	})
	require.NoError(t, err)
	require.Equal(t, "主控告警", row.Name)

	rowsA, err := svc.ListSubscriptions(ctxA, "tenant-cta", nil)
	require.NoError(t, err)
	require.Len(t, rowsA, 1)

	rowsB, err := svc.ListSubscriptions(ctxB, "tenant-ctb", nil)
	require.NoError(t, err)
	require.Len(t, rowsB, 0)
}

func ensureControlTowerTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_control_tower_snapshots (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		window_hours INTEGER NOT NULL DEFAULT 24,
		carrier_id TEXT,
		warehouse_id TEXT,
		destination_zone TEXT,
		in_transit_count INTEGER NOT NULL DEFAULT 0,
		exception_count INTEGER NOT NULL DEFAULT 0,
		timeout_count INTEGER NOT NULL DEFAULT 0,
		delivered_count INTEGER NOT NULL DEFAULT 0,
		on_time_rate NUMERIC NOT NULL DEFAULT 0,
		total_cost NUMERIC NOT NULL DEFAULT 0,
		alert_count INTEGER NOT NULL DEFAULT 0,
		summary JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_control_tower_alert_subscriptions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		carrier_id TEXT,
		warehouse_id TEXT,
		destination_zone TEXT,
		min_on_time_rate NUMERIC NOT NULL DEFAULT 95,
		max_timeout_count INTEGER NOT NULL DEFAULT 5,
		max_cost_amount NUMERIC NOT NULL DEFAULT 0,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		config JSON,
		last_notified_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_control_tower_subscription ON logistics_control_tower_alert_subscriptions(tenant_uuid, name)`).Error)
}
