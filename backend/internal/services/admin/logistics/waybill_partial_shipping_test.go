package logistics

import (
	"context"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWaybillService_PartialShippingAggregationConsistency(t *testing.T) {
	db := setupWaybillPartialShippingDB(t, "logistics_partial_shipping")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-1",
		TenantUUID: "tenant-1",
		Name:       "自营",
		Code:       "self",
		Type:       "self",
		Status:     "active",
	}).Error)

	svc := NewWaybillService(&app.Deps{DB: db})
	repo := LogisticsRepo.NewWaybillRepository(db)

	wb1, status1, err := svc.Create(ctx, "tenant-1", CreateWaybillRequest{
		OrderID:        "order-1",
		CarrierID:      "carrier-1",
		ServiceCode:    "std",
		PackageKey:     "order-1-pkg-1",
		ShipmentItems:  []string{"sku-a", "sku-b"},
		OrderItemCount: 3,
		WaybillNo:      "WB-ORDER1-1",
		PackageNo:      1,
	})
	require.NoError(t, err)
	require.Equal(t, "created", status1)
	require.Equal(t, "partial_shipped", wb1.OrderFulfillmentStatus)

	wb2, status2, err := svc.Create(ctx, "tenant-1", CreateWaybillRequest{
		OrderID:        "order-1",
		CarrierID:      "carrier-1",
		ServiceCode:    "std",
		PackageKey:     "order-1-pkg-2",
		ShipmentItems:  []string{"sku-c"},
		OrderItemCount: 3,
		WaybillNo:      "WB-ORDER1-2",
		PackageNo:      2,
	})
	require.NoError(t, err)
	require.Equal(t, "created", status2)
	require.Equal(t, "fully_shipped", wb2.OrderFulfillmentStatus)

	savedWb1, err := repo.GetByID(ctx, wb1.ID)
	require.NoError(t, err)
	require.Equal(t, "fully_shipped", savedWb1.OrderFulfillmentStatus)
	require.Equal(t, 3, savedWb1.OrderItemCount)
}

func TestWaybillService_CreateWaybillReplayByPackageKey(t *testing.T) {
	db := setupWaybillPartialShippingDB(t, "logistics_package_key_idem")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-1",
		TenantUUID: "tenant-1",
		Name:       "自营",
		Code:       "self",
		Type:       "self",
		Status:     "active",
	}).Error)

	svc := NewWaybillService(&app.Deps{DB: db})

	first, firstStatus, err := svc.Create(ctx, "tenant-1", CreateWaybillRequest{
		OrderID:        "order-2",
		CarrierID:      "carrier-1",
		ServiceCode:    "std",
		PackageKey:     "order-2-pkg-1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 2,
		WaybillNo:      "WB-ORDER2-1",
	})
	require.NoError(t, err)
	require.Equal(t, "created", firstStatus)

	second, secondStatus, err := svc.Create(ctx, "tenant-1", CreateWaybillRequest{
		OrderID:        "order-2",
		CarrierID:      "carrier-1",
		ServiceCode:    "std",
		PackageKey:     "order-2-pkg-1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 2,
		WaybillNo:      "WB-ORDER2-1-DUP",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", secondStatus)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, first.WaybillNo, second.WaybillNo)
}

func setupWaybillPartialShippingDB(t *testing.T, name string) *gorm.DB {
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
	return db
}
