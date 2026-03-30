package logistics

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBillingService_DiffCalculationAndTenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_billing_service")
	ctx := context.Background()

	svc := NewBillingService(&app.Deps{DB: db})
	wbRepo := NewWaybillService(&app.Deps{DB: db})

	tenant1 := authx.ContextWithTenantUUID(ctx, "tenant-1")
	tenant2 := authx.ContextWithTenantUUID(ctx, "tenant-2")

	require.NoError(t, db.WithContext(tenant1).Create(&LogisticsModel.Carrier{
		ID:         "carrier-1",
		TenantUUID: "tenant-1",
		Name:       "Carrier A",
		Code:       "a",
		Type:       "self",
		Status:     "active",
	}).Error)
	require.NoError(t, db.WithContext(tenant2).Create(&LogisticsModel.Carrier{
		ID:         "carrier-2",
		TenantUUID: "tenant-2",
		Name:       "Carrier B",
		Code:       "b",
		Type:       "self",
		Status:     "active",
	}).Error)

	wb1, _, err := wbRepo.Create(tenant1, "tenant-1", CreateWaybillRequest{
		OrderID:        "order-1",
		CarrierID:      "carrier-1",
		ServiceCode:    "std",
		WaybillNo:      "WB-T1-1",
		PackageKey:     "order-1-pkg-1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb1.FeeAmount = 10
	require.NoError(t, db.WithContext(tenant1).Save(wb1).Error)

	_, _, err = wbRepo.Create(tenant2, "tenant-2", CreateWaybillRequest{
		OrderID:        "order-2",
		CarrierID:      "carrier-2",
		ServiceCode:    "std",
		WaybillNo:      "WB-T2-1",
		PackageKey:     "order-2-pkg-1",
		ShipmentItems:  []string{"sku-9"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)

	updated, err := svc.UpdateWaybillCost(tenant1, "tenant-1", wb1.ID, UpdateWaybillCostRequest{ActualFeeAmount: 12.5})
	require.NoError(t, err)
	require.Equal(t, 12.5, updated.ActualFeeAmount)
	require.Equal(t, 2.5, updated.FeeDiffAmount)
	require.Equal(t, "settled", updated.BillingStatus)
	require.NotNil(t, updated.SettledAt)

	snapshot, err := svc.Snapshot(tenant1, "tenant-1", BillingQuery{})
	require.NoError(t, err)
	require.Len(t, snapshot.Items, 1)
	require.Equal(t, "WB-T1-1", snapshot.Items[0].WaybillNo)
	require.Len(t, snapshot.Summary, 1)
	require.Equal(t, 2.5, snapshot.Summary[0].DiffFee)
}

func TestBillingService_ExportValidation(t *testing.T) {
	db := setupBillingDB(t, "logistics_billing_export_validate")
	svc := NewBillingService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	_, err := svc.ExportPayload(ctx, "tenant-1", BillingQuery{From: "bad-time"}, "csv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "from must be RFC3339")

	_, err = svc.ExportPayload(ctx, "tenant-1", BillingQuery{}, "xlsx")
	require.Error(t, err)
	require.Contains(t, err.Error(), "format must be csv or json")
}

func setupBillingDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_carriers (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		contact_name TEXT,
		contact_phone TEXT,
		capabilities JSON,
		config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_waybills (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		order_id TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		service_code TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		package_no INTEGER NOT NULL DEFAULT 1,
		package_key TEXT NOT NULL DEFAULT '',
		shipment_items JSON,
		order_item_count INTEGER NOT NULL DEFAULT 0,
		order_fulfillment_status TEXT NOT NULL DEFAULT 'partial_shipped',
		status TEXT NOT NULL,
		fee_amount NUMERIC NOT NULL DEFAULT 0,
		actual_fee_amount NUMERIC NOT NULL DEFAULT 0,
		fee_diff_amount NUMERIC NOT NULL DEFAULT 0,
		billing_status TEXT NOT NULL DEFAULT 'pending',
		settled_at DATETIME,
		label_url TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_waybill_no ON logistics_waybills(tenant_uuid, waybill_no)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_tracking_events (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		event_id TEXT NOT NULL,
		status TEXT NOT NULL,
		source TEXT NOT NULL,
		description TEXT,
		occurred_at DATETIME,
		payload JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_tracking_event ON logistics_tracking_events(tenant_uuid, waybill_no, event_id)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_billing_cases (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		case_no TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'open',
		diff_amount NUMERIC NOT NULL DEFAULT 0,
		reason TEXT,
		resolution TEXT,
		metadata JSON,
		closed_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_billing_case_no ON logistics_billing_cases(tenant_uuid, case_no)`).Error)

	// ensure deterministic created_at for billing window filtering if needed
	require.NoError(t, db.Exec(`UPDATE logistics_waybills SET created_at = ? WHERE created_at IS NULL`, time.Now().UTC()).Error)
	return db
}
