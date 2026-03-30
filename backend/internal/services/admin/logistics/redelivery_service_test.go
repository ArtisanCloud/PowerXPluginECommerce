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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRedeliveryService_StateTransitionsAndIdempotency(t *testing.T) {
	db := setupRedeliveryDB(t, "logistics_redelivery_state")
	seedRedeliveryWaybill(t, db, "tenant-1", "wb-1", "WB-1")
	svc := NewRedeliveryService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	task, idem, err := svc.Initiate(ctx, "tenant-1", InitiateRedeliveryRequest{
		WaybillID:  "wb-1",
		RequestKey: "req-init-1",
		Reason:     "failed_delivery",
	})
	require.NoError(t, err)
	require.Equal(t, "created", idem)
	require.Equal(t, "initiated", task.Status)
	require.Equal(t, 1, task.AttemptNo)

	replayed, idem, err := svc.Initiate(ctx, "tenant-1", InitiateRedeliveryRequest{
		WaybillID:  "wb-1",
		RequestKey: "req-init-1",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", idem)
	require.Equal(t, task.ID, replayed.ID)

	updated, err := svc.UpdateAddress(ctx, "tenant-1", task.ID, UpdateRedeliveryAddressRequest{
		Address: map[string]any{"line1": "new address"},
		Reason:  "address_changed",
	})
	require.NoError(t, err)
	require.Equal(t, "address_updated", updated.Status)

	redispatched, idem, err := svc.Redispatch(ctx, "tenant-1", task.ID, RedispatchRedeliveryRequest{
		RequestKey: "req-redispatch-1",
		Reason:     "retry_delivery",
	})
	require.NoError(t, err)
	require.Equal(t, "created", idem)
	require.Equal(t, "redispatched", redispatched.Status)
	require.Equal(t, 2, redispatched.AttemptNo)

	replayedRedispatch, idem, err := svc.Redispatch(ctx, "tenant-1", task.ID, RedispatchRedeliveryRequest{
		RequestKey: "req-redispatch-1",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", idem)
	require.Equal(t, 2, replayedRedispatch.AttemptNo)

	closed, err := svc.Close(ctx, "tenant-1", task.ID, CloseRedeliveryRequest{Reason: "resolved"})
	require.NoError(t, err)
	require.Equal(t, "closed", closed.Status)
	require.NotNil(t, closed.ClosedAt)
}

func TestRedeliveryService_TenantIsolation(t *testing.T) {
	db := setupRedeliveryDB(t, "logistics_redelivery_tenant")
	seedRedeliveryWaybill(t, db, "tenant-1", "wb-1", "WB-1")
	seedRedeliveryWaybill(t, db, "tenant-2", "wb-2", "WB-2")
	svc := NewRedeliveryService(&app.Deps{DB: db})

	ctx1 := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	task, _, err := svc.Initiate(ctx1, "tenant-1", InitiateRedeliveryRequest{WaybillID: "wb-1", RequestKey: "t1-init"})
	require.NoError(t, err)

	ctx2 := authx.ContextWithTenantUUID(context.Background(), "tenant-2")
	_, err = svc.UpdateAddress(ctx2, "tenant-2", task.ID, UpdateRedeliveryAddressRequest{
		Address: map[string]any{"line1": "x"},
	})
	require.Error(t, err)
}

func setupRedeliveryDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
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
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_redelivery_tasks (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		request_key TEXT NOT NULL,
		status TEXT NOT NULL,
		attempt_no INTEGER NOT NULL DEFAULT 1,
		address_snapshot JSON,
		last_reason TEXT,
		operator_id TEXT,
		closed_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_redelivery_request ON logistics_redelivery_tasks(tenant_uuid, request_key)`).Error)
	return db
}

func seedRedeliveryWaybill(t *testing.T, db *gorm.DB, tenantUUID, waybillID, waybillNo string) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.Create(&logisticsmodel.Waybill{
		ID:          waybillID,
		TenantUUID:  tenantUUID,
		OrderID:     "order-" + waybillID,
		CarrierID:   "carrier-1",
		ServiceCode: "std",
		WaybillNo:   waybillNo,
		Status:      "exception",
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error)
}
