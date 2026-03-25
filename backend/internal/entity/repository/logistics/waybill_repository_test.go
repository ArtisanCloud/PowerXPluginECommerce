package logistics

import (
	"context"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	logisticsmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTrackingEventRepository_IdempotencyKey(t *testing.T) {
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:logistics_repo_tracking?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_waybills (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		order_id TEXT NOT NULL,
		carrier_id TEXT NOT NULL,
		service_code TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		status TEXT NOT NULL,
		fee_amount NUMERIC NOT NULL DEFAULT 0,
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

	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	waybillRepo := NewWaybillRepository(db)
	trackRepo := NewTrackingEventRepository(db)

	wb := &logisticsmodel.Waybill{
		ID:          "wb-1",
		OrderID:     "order-1",
		CarrierID:   "carrier-1",
		ServiceCode: "std",
		WaybillNo:   "WB-0001",
		Status:      "created",
	}
	require.NoError(t, waybillRepo.Create(ctx, wb))

	event := &logisticsmodel.TrackingEvent{
		ID:        "evt-1",
		WaybillID: wb.ID,
		WaybillNo: wb.WaybillNo,
		EventID:   "event-1",
		Status:    "in_transit",
		Source:    "provider",
	}
	require.NoError(t, trackRepo.Create(ctx, event))

	exists, err := trackRepo.ExistsByEventKey(ctx, wb.WaybillNo, "event-1")
	require.NoError(t, err)
	require.True(t, exists)

	dup := &logisticsmodel.TrackingEvent{
		ID:        "evt-2",
		WaybillID: wb.ID,
		WaybillNo: wb.WaybillNo,
		EventID:   "event-1",
		Status:    "delivered",
		Source:    "provider",
	}
	err = trackRepo.Create(ctx, dup)
	require.Error(t, err)
	require.True(t, IsDuplicateConstraintError(err))
}
