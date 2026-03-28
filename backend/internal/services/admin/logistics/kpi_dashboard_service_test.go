package logistics

import (
	"context"
	"strings"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestKPIDashboardService_OverviewTrendDrilldown(t *testing.T) {
	db := setupBillingDB(t, "logistics_kpi_dashboard")
	ensureKPIDashboardTables(t, db)

	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-kpi")
	seedCarrier(t, db, "tenant-kpi", "carrier-kpi-1", "Carrier-KPI-1")
	seedCarrier(t, db, "tenant-kpi", "carrier-kpi-2", "Carrier-KPI-2")

	waybillSvc := NewWaybillService(&app.Deps{DB: db})
	svc := NewKPIDashboardService(&app.Deps{DB: db})

	wb1, _, err := waybillSvc.Create(ctx, "tenant-kpi", CreateWaybillRequest{
		OrderID:        "order-kpi-1",
		CarrierID:      "carrier-kpi-1",
		ServiceCode:    "std",
		WaybillNo:      "WB-KPI-1",
		PackageKey:     "order-kpi-1#1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb1.Status = "delivered"
	wb1.FeeAmount = 10
	wb1.Metadata = datatypes.JSON([]byte(`{"warehouse_id":"wh-a","destination_zone":"CN-EAST"}`))
	wb1.CreatedAt = time.Now().UTC().Add(-10 * time.Hour)
	wb1.UpdatedAt = time.Now().UTC().Add(-1 * time.Hour)
	require.NoError(t, db.WithContext(ctx).Save(wb1).Error)

	wb2, _, err := waybillSvc.Create(ctx, "tenant-kpi", CreateWaybillRequest{
		OrderID:        "order-kpi-2",
		CarrierID:      "carrier-kpi-2",
		ServiceCode:    "std",
		WaybillNo:      "WB-KPI-2",
		PackageKey:     "order-kpi-2#1",
		ShipmentItems:  []string{"sku-2"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb2.Status = "delay"
	wb2.ActualFeeAmount = 20
	wb2.Metadata = datatypes.JSON([]byte(`{"warehouse_id":"wh-b","destination_zone":"CN-WEST"}`))
	wb2.CreatedAt = time.Now().UTC().Add(-90 * time.Hour)
	wb2.UpdatedAt = time.Now().UTC().Add(-2 * time.Hour)
	require.NoError(t, db.WithContext(ctx).Save(wb2).Error)

	overview, err := svc.Overview(ctx, "tenant-kpi", KPIDashboardQuery{
		WindowHours: 120,
		Dimension:   "carrier",
	})
	require.NoError(t, err)
	require.Equal(t, 2, overview.TotalWaybills)
	require.Equal(t, 1, overview.DeliveredCount)
	require.Equal(t, 1, overview.ExceptionCount)
	require.Equal(t, 1, overview.TimeoutCount)

	trends, err := svc.Trends(ctx, "tenant-kpi", KPIDashboardQuery{
		WindowHours: 120,
		Dimension:   "warehouse",
	})
	require.NoError(t, err)
	require.Len(t, trends, 2)

	drilldown, err := svc.Drilldown(ctx, "tenant-kpi", KPIDashboardQuery{
		WindowHours: 120,
		CarrierID:   "carrier-kpi-2",
		Limit:       20,
	})
	require.NoError(t, err)
	require.Len(t, drilldown, 1)
	require.Equal(t, "WB-KPI-2", drilldown[0].WaybillNo)
}

func TestKPIDashboardService_ExportAndTenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_kpi_dashboard_tenant")
	ensureKPIDashboardTables(t, db)

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-kpi-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-kpi-b")
	svc := NewKPIDashboardService(&app.Deps{DB: db})
	waybillSvc := NewWaybillService(&app.Deps{DB: db})

	seedCarrier(t, db, "tenant-kpi-a", "carrier-kpi-a", "Carrier-KPI-A")
	_, _, err := waybillSvc.Create(ctxA, "tenant-kpi-a", CreateWaybillRequest{
		OrderID:        "order-kpi-a",
		CarrierID:      "carrier-kpi-a",
		ServiceCode:    "std",
		WaybillNo:      "WB-KPI-A",
		PackageKey:     "order-kpi-a#1",
		ShipmentItems:  []string{"sku-a"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)

	content, err := svc.ExportCSV(ctxA, "tenant-kpi-a", KPIDashboardQuery{Dimension: "carrier", WindowHours: 72})
	require.NoError(t, err)
	require.True(t, strings.Contains(content, "dimension_key"))

	overviewB, err := svc.Overview(ctxB, "tenant-kpi-b", KPIDashboardQuery{Dimension: "carrier", WindowHours: 72})
	require.NoError(t, err)
	require.Equal(t, 0, overviewB.TotalWaybills)
}

func ensureKPIDashboardTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_kpi_snapshots (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		window_hours INTEGER NOT NULL DEFAULT 24,
		dimension_type TEXT NOT NULL,
		dimension_key TEXT NOT NULL,
		total_waybills INTEGER NOT NULL DEFAULT 0,
		delivered_count INTEGER NOT NULL DEFAULT 0,
		exception_count INTEGER NOT NULL DEFAULT 0,
		timeout_count INTEGER NOT NULL DEFAULT 0,
		on_time_rate NUMERIC NOT NULL DEFAULT 0,
		delivery_success_rate NUMERIC NOT NULL DEFAULT 0,
		avg_transit_hours NUMERIC NOT NULL DEFAULT 0,
		avg_cost NUMERIC NOT NULL DEFAULT 0,
		total_cost NUMERIC NOT NULL DEFAULT 0,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
}
