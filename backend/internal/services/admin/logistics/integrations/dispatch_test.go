package integrations_test

import (
	"context"
	"encoding/json"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	logisticssvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWaybillService_ProviderSwitchAndWebhookIdempotency(t *testing.T) {
	db := setupLogisticsIntegrationDB(t, "logistics_us4_dispatch")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-1",
		TenantUUID: "tenant-1",
		Name:       "顺丰",
		Code:       "sf",
		Type:       "self",
		Status:     "active",
		Config:     []byte(`{"provider":"sf","service_code_map":{"std":"SF_STD"}}`),
	}).Error)

	svc := logisticssvc.NewWaybillService(&app.Deps{DB: db})
	wb1, _, err := svc.Create(ctx, "tenant-1", logisticssvc.CreateWaybillRequest{
		OrderID:     "order-1",
		CarrierID:   "carrier-1",
		ServiceCode: "std",
	})
	require.NoError(t, err)
	require.Equal(t, "SF_STD", wb1.ServiceCode)

	cfg := map[string]any{"provider": "jd", "service_code_map": map[string]any{"std": "JD_STD"}}
	raw, _ := json.Marshal(cfg)
	require.NoError(t, db.WithContext(ctx).Model(&LogisticsModel.Carrier{}).
		Where("id = ?", "carrier-1").
		Update("config", raw).Error)

	wb2, _, err := svc.Create(ctx, "tenant-1", logisticssvc.CreateWaybillRequest{
		OrderID:     "order-2",
		CarrierID:   "carrier-1",
		ServiceCode: "std",
	})
	require.NoError(t, err)
	require.Equal(t, "JD_STD", wb2.ServiceCode)

	_, status, err := svc.AppendTracking(ctx, "tenant-1", wb2.ID, logisticssvc.AppendTrackingRequest{
		EventID: "evt-1",
		Status:  "in_transit",
		Source:  "provider",
	})
	require.NoError(t, err)
	require.Equal(t, "created", status)

	_, status, err = svc.AppendTracking(ctx, "tenant-1", wb2.ID, logisticssvc.AppendTrackingRequest{
		EventID: "evt-1",
		Status:  "in_transit",
		Source:  "provider",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", status)
}

func setupLogisticsIntegrationDB(t *testing.T, name string) *gorm.DB {
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
