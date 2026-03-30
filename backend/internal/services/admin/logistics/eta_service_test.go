package logistics

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	logisticsmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestETAService_WindowAndDeadlineComputation(t *testing.T) {
	db := setupETADB(t, "logistics_eta_window")
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	created := time.Date(2026, 3, 14, 11, 30, 0, 0, time.UTC) // 19:30 +08, over cutoff.
	seedETAFixtures(t, db, "tenant-1", created)
	require.NoError(t, db.Create(&logisticsmodel.ETAPolicy{
		ID:               "policy-1",
		TenantUUID:       "tenant-1",
		CarrierID:        "carrier-1",
		ServiceCode:      "std",
		DestinationZone:  "GLOBAL",
		Timezone:         "Asia/Shanghai",
		CutoffHourLocal:  18,
		PickupSLAHours:   24,
		DeliverySLAHours: 48,
		Metadata:         datatypes.JSON([]byte(`{"note":"test"}`)),
	}).Error)

	svc := NewETAService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	result, err := svc.GetByWaybillID(ctx, "tenant-1", "wb-1", ETAQuery{})
	require.NoError(t, err)
	require.NotNil(t, result.PromisedAt)
	require.NotNil(t, result.PickupDeadlineAt)

	expectedPickup := time.Date(2026, 3, 16, 0, 0, 0, 0, loc).UTC()
	expectedPromised := time.Date(2026, 3, 18, 0, 0, 0, 0, loc).UTC()
	require.True(t, result.PickupDeadlineAt.UTC().Equal(expectedPickup))
	require.True(t, result.PromisedAt.UTC().Equal(expectedPromised))
}

func TestETAService_CrossTimezoneQueryAffectsPromise(t *testing.T) {
	db := setupETADB(t, "logistics_eta_timezone")
	created := time.Date(2026, 3, 14, 12, 30, 0, 0, time.UTC)
	seedETAFixtures(t, db, "tenant-1", created)
	// Default policy in UTC.
	require.NoError(t, db.Create(&logisticsmodel.ETAPolicy{
		ID:               "policy-2",
		TenantUUID:       "tenant-1",
		CarrierID:        "carrier-1",
		ServiceCode:      "std",
		DestinationZone:  "GLOBAL",
		Timezone:         "UTC",
		CutoffHourLocal:  18,
		PickupSLAHours:   24,
		DeliverySLAHours: 48,
	}).Error)

	svc := NewETAService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	utcResult, err := svc.GetByWaybillID(ctx, "tenant-1", "wb-1", ETAQuery{ForceRecompute: true})
	require.NoError(t, err)
	shResult, err := svc.GetByWaybillID(ctx, "tenant-1", "wb-1", ETAQuery{Timezone: "Asia/Shanghai", ForceRecompute: true})
	require.NoError(t, err)

	require.NotNil(t, utcResult.PromisedAt)
	require.NotNil(t, shResult.PromisedAt)
	require.NotEqual(t, utcResult.PromisedAt.UTC(), shResult.PromisedAt.UTC())
	require.Equal(t, "Asia/Shanghai", shResult.Timezone)
}

func TestETAService_TenantIsolation(t *testing.T) {
	db := setupETADB(t, "logistics_eta_tenant")
	created := time.Now().UTC().Add(-2 * time.Hour)
	seedETAFixtures(t, db, "tenant-1", created)
	seedETAFixtures(t, db, "tenant-2", created)

	svc := NewETAService(&app.Deps{DB: db})
	ctx1 := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	rows, err := svc.ListByWaybillIDs(ctx1, "tenant-1", []string{"wb-1", "wb-2"}, ETAQuery{ForceRecompute: true})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "wb-1", rows[0].WaybillID)

	ctx2 := authx.ContextWithTenantUUID(context.Background(), "tenant-2")
	rows2, err := svc.ListByWaybillIDs(ctx2, "tenant-2", []string{"wb-1", "wb-2"}, ETAQuery{ForceRecompute: true})
	require.NoError(t, err)
	require.Len(t, rows2, 1)
	require.Equal(t, "wb-2", rows2[0].WaybillID)
}

func setupETADB(t *testing.T, name string) *gorm.DB {
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
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_eta_policies (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		service_code TEXT NOT NULL,
		destination_zone TEXT NOT NULL,
		timezone TEXT NOT NULL,
		cutoff_hour_local INTEGER NOT NULL DEFAULT 18,
		pickup_sla_hours INTEGER NOT NULL DEFAULT 24,
		delivery_sla_hours INTEGER NOT NULL DEFAULT 72,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_eta_records (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		service_code TEXT NOT NULL,
		timezone TEXT NOT NULL,
		pickup_deadline_at DATETIME,
		delivery_deadline_at DATETIME,
		promised_at DATETIME,
		estimated_at DATETIME,
		source TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		metadata JSON,
		last_computed_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_waybill_no ON logistics_waybills(tenant_uuid, waybill_no)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_tracking_event ON logistics_tracking_events(tenant_uuid, waybill_no, event_id)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_eta_policy ON logistics_eta_policies(tenant_uuid, carrier_id, service_code, destination_zone)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_eta_waybill ON logistics_eta_records(tenant_uuid, waybill_id)`).Error)
	return db
}

func seedETAFixtures(t *testing.T, db *gorm.DB, tenant string, createdAt time.Time) {
	t.Helper()
	carrierID := "carrier-1"
	waybillID := "wb-1"
	waybillNo := "WB-001"
	if tenant == "tenant-2" {
		carrierID = "carrier-2"
		waybillID = "wb-2"
		waybillNo = "WB-002"
	}
	require.NoError(t, db.Create(&logisticsmodel.Carrier{
		ID:         carrierID,
		TenantUUID: tenant,
		Name:       "承运商",
		Code:       "self",
		Type:       "self",
		Status:     "active",
	}).Error)
	require.NoError(t, db.Create(&logisticsmodel.Waybill{
		ID:          waybillID,
		TenantUUID:  tenant,
		OrderID:     "order-1",
		CarrierID:   carrierID,
		ServiceCode: "std",
		WaybillNo:   waybillNo,
		Status:      "in_transit",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}).Error)
}
