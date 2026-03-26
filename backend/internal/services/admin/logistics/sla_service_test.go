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

func TestSLAService_SnapshotWindowAggregation(t *testing.T) {
	db := setupSLADB(t, "logistics_sla_window")
	now := time.Now().UTC()
	seedSLAFixtures(t, db, now)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	svc := NewSLAService(&app.Deps{DB: db})

	snapshot, err := svc.Snapshot(ctx, "tenant-1", SLAQuery{
		PickupSLAHours:   24,
		DeliverySLAHours: 72,
	})
	require.NoError(t, err)
	require.Len(t, snapshot.Summary, 1)

	item := snapshot.Summary[0]
	require.Equal(t, int64(3), item.WaybillCount)
	require.Equal(t, int64(2), item.PickupOnTimeCount)
	require.Equal(t, int64(2), item.SignOnTimeCount)
	require.Equal(t, int64(1), item.ExceptionCount)
	require.InDelta(t, 66.67, item.PickupOnTimeRate, 0.2)
	require.InDelta(t, 66.67, item.SignOnTimeRate, 0.2)
	require.InDelta(t, 33.33, item.ExceptionRate, 0.2)
}

func TestSLAService_SnapshotTenantIsolation(t *testing.T) {
	db := setupSLADB(t, "logistics_sla_tenant")
	now := time.Now().UTC()
	seedSLAFixtures(t, db, now)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	svc := NewSLAService(&app.Deps{DB: db})

	snapshot, err := svc.Snapshot(ctx, "tenant-1", SLAQuery{})
	require.NoError(t, err)
	require.Equal(t, int64(3), snapshot.Total.WaybillCount)

	ctxTenant2 := authx.ContextWithTenantUUID(context.Background(), "tenant-2")
	snapshot2, err := svc.Snapshot(ctxTenant2, "tenant-2", SLAQuery{})
	require.NoError(t, err)
	require.Equal(t, int64(1), snapshot2.Total.WaybillCount)
}

func seedSLAFixtures(t *testing.T, db *gorm.DB, now time.Time) {
	t.Helper()
	require.NoError(t, db.Create(&LogisticsModel.Carrier{
		ID:         "carrier-1",
		TenantUUID: "tenant-1",
		Name:       "承运商A",
		Code:       "a",
		Type:       "self",
		Status:     "active",
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.Carrier{
		ID:         "carrier-2",
		TenantUUID: "tenant-2",
		Name:       "承运商B",
		Code:       "b",
		Type:       "self",
		Status:     "active",
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.Waybill{
		ID:          "wb-1",
		TenantUUID:  "tenant-1",
		OrderID:     "order-1",
		CarrierID:   "carrier-1",
		ServiceCode: "std",
		WaybillNo:   "WB-1",
		Status:      "delivered",
		CreatedAt:   now.Add(-10 * time.Hour),
		UpdatedAt:   now,
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.Waybill{
		ID:          "wb-2",
		TenantUUID:  "tenant-1",
		OrderID:     "order-2",
		CarrierID:   "carrier-1",
		ServiceCode: "std",
		WaybillNo:   "WB-2",
		Status:      "in_transit",
		CreatedAt:   now.Add(-30 * time.Hour),
		UpdatedAt:   now,
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.Waybill{
		ID:          "wb-3",
		TenantUUID:  "tenant-1",
		OrderID:     "order-3",
		CarrierID:   "carrier-1",
		ServiceCode: "std",
		WaybillNo:   "WB-3",
		Status:      "exception",
		CreatedAt:   now.Add(-50 * time.Hour),
		UpdatedAt:   now,
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.Waybill{
		ID:          "wb-4",
		TenantUUID:  "tenant-2",
		OrderID:     "order-4",
		CarrierID:   "carrier-2",
		ServiceCode: "std",
		WaybillNo:   "WB-4",
		Status:      "delivered",
		CreatedAt:   now.Add(-12 * time.Hour),
		UpdatedAt:   now,
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.TrackingEvent{
		ID:         "te-1",
		TenantUUID: "tenant-1",
		WaybillID:  "wb-1",
		WaybillNo:  "WB-1",
		EventID:    "ev-1",
		Status:     "picked",
		Source:     "provider",
		OccurredAt: ptrTime(now.Add(-9 * time.Hour)),
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.TrackingEvent{
		ID:         "te-2",
		TenantUUID: "tenant-1",
		WaybillID:  "wb-1",
		WaybillNo:  "WB-1",
		EventID:    "ev-2",
		Status:     "delivered",
		Source:     "provider",
		OccurredAt: ptrTime(now.Add(-2 * time.Hour)),
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.TrackingEvent{
		ID:         "te-3",
		TenantUUID: "tenant-1",
		WaybillID:  "wb-2",
		WaybillNo:  "WB-2",
		EventID:    "ev-3",
		Status:     "picked",
		Source:     "provider",
		OccurredAt: ptrTime(now.Add(-2 * time.Hour)),
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.TrackingEvent{
		ID:         "te-4",
		TenantUUID: "tenant-1",
		WaybillID:  "wb-3",
		WaybillNo:  "WB-3",
		EventID:    "ev-4",
		Status:     "picked",
		Source:     "provider",
		OccurredAt: ptrTime(now.Add(-49 * time.Hour)),
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.TrackingEvent{
		ID:         "te-5",
		TenantUUID: "tenant-1",
		WaybillID:  "wb-3",
		WaybillNo:  "WB-3",
		EventID:    "ev-5",
		Status:     "delivered",
		Source:     "provider",
		OccurredAt: ptrTime(now.Add(-1 * time.Hour)),
	}).Error)
	require.NoError(t, db.Create(&LogisticsModel.TrackingEvent{
		ID:         "te-6",
		TenantUUID: "tenant-2",
		WaybillID:  "wb-4",
		WaybillNo:  "WB-4",
		EventID:    "ev-6",
		Status:     "picked",
		Source:     "provider",
		OccurredAt: ptrTime(now.Add(-10 * time.Hour)),
	}).Error)
}

func setupSLADB(t *testing.T, name string) *gorm.DB {
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

func ptrTime(v time.Time) *time.Time {
	return &v
}
