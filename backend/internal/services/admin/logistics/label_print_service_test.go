package logistics

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLabelPrintService_BatchFailThenRetrySuccess(t *testing.T) {
	db := setupLabelPrintDB(t, "logistics_label_print_retry")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	seedLabelPrintFixtures(t, db, ctx)

	svc := NewLabelPrintService(&app.Deps{DB: db})
	resp, err := svc.BatchPrint(ctx, "tenant-1", BatchPrintLabelsRequest{
		WaybillIDs: []string{"wb-id-1", "wb-id-2"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, resp.Success)
	require.Equal(t, 1, resp.Failed)

	var failedTaskID string
	for _, item := range resp.Results {
		if item.Task != nil && item.Task.Status == "failed" {
			failedTaskID = item.Task.ID
		}
	}
	require.NotEmpty(t, failedTaskID)

	waybillRepo := LogisticsRepo.NewWaybillRepository(db)
	wb, err := waybillRepo.GetByID(ctx, "wb-id-2")
	require.NoError(t, err)
	wb.LabelURL = "https://cdn.powerx.dev/label-2.pdf"
	require.NoError(t, waybillRepo.Save(ctx, wb))

	retryResp, err := svc.RetryFailed(ctx, "tenant-1", RetryLabelPrintRequest{TaskIDs: []string{failedTaskID}})
	require.NoError(t, err)
	require.Equal(t, 1, retryResp.Success)
	require.Equal(t, 0, retryResp.Failed)
	require.Len(t, retryResp.Results, 1)
	require.NotNil(t, retryResp.Results[0].Task)
	require.Equal(t, "success", retryResp.Results[0].Task.Status)
	require.NotNil(t, retryResp.Results[0].Task.PrintedAt)
}

func TestLabelPrintService_IdempotentReprint(t *testing.T) {
	db := setupLabelPrintDB(t, "logistics_label_print_idempotent")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	seedLabelPrintFixtures(t, db, ctx)

	svc := NewLabelPrintService(&app.Deps{DB: db})
	first, err := svc.BatchPrint(ctx, "tenant-1", BatchPrintLabelsRequest{
		WaybillIDs:     []string{"wb-id-1"},
		IdempotencyKey: "reprint-order-1",
		ReprintReason:  "template-updated",
	})
	require.NoError(t, err)
	require.Equal(t, 1, first.Success)
	require.Equal(t, 0, first.Failed)
	require.Len(t, first.Results, 1)
	require.Equal(t, "created", first.Results[0].IdempotencyStatus)

	second, err := svc.BatchPrint(ctx, "tenant-1", BatchPrintLabelsRequest{
		WaybillIDs:     []string{"wb-id-1"},
		IdempotencyKey: "reprint-order-1",
		ReprintReason:  "template-updated",
	})
	require.NoError(t, err)
	require.Len(t, second.Results, 1)
	require.Equal(t, "replayed", second.Results[0].IdempotencyStatus)

	taskRepo := LogisticsRepo.NewLabelPrintRepository(db)
	rows, err := taskRepo.List(ctx, "", 20)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}

func seedLabelPrintFixtures(t *testing.T, db *gorm.DB, ctx context.Context) {
	t.Helper()
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-1",
		TenantUUID: "tenant-1",
		Name:       "自营",
		Code:       "self",
		Type:       "self",
		Status:     "active",
	}).Error)
	now := time.Now().UTC()
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:                     "wb-id-1",
		TenantUUID:             "tenant-1",
		OrderID:                "order-1",
		CarrierID:              "carrier-1",
		ServiceCode:            "std",
		WaybillNo:              "WB-1",
		PackageNo:              1,
		PackageKey:             "order-1#1",
		Status:                 "created",
		OrderFulfillmentStatus: "partial_shipped",
		LabelURL:               "https://cdn.powerx.dev/label-1.pdf",
		CreatedAt:              now,
		UpdatedAt:              now,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:                     "wb-id-2",
		TenantUUID:             "tenant-1",
		OrderID:                "order-2",
		CarrierID:              "carrier-1",
		ServiceCode:            "std",
		WaybillNo:              "WB-2",
		PackageNo:              1,
		PackageKey:             "order-2#1",
		Status:                 "created",
		OrderFulfillmentStatus: "partial_shipped",
		LabelURL:               "",
		CreatedAt:              now,
		UpdatedAt:              now,
	}).Error)
}

func setupLabelPrintDB(t *testing.T, name string) *gorm.DB {
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
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_label_print_tasks (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		status TEXT NOT NULL,
		attempt_count INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 3,
		last_error TEXT,
		retry_queued_at DATETIME,
		printed_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_label_print_request ON logistics_label_print_tasks(tenant_uuid, request_key)`).Error)
	return db
}
