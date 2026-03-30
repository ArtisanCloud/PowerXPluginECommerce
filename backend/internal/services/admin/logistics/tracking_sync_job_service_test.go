package logistics

import (
	"context"
	"strconv"
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

func TestTrackingSyncJobService_CreateAndRun(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "tracking_sync_job_create_run")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	seedTrackingSyncJobFixtures(t, db, ctx)

	svc := NewTrackingSyncJobService(&app.Deps{DB: db})
	job, err := svc.CreateAndRun(ctx, "tenant-a", CreateTrackingSyncJobRequest{
		CarrierID:     "carrier-a",
		WaybillStatus: "in_transit",
		BatchLimit:    20,
		EventLimit:    10,
	})
	require.NoError(t, err)
	require.NotNil(t, job)
	require.Equal(t, "success", job.Status)
	require.Equal(t, 1, job.TotalWaybills)
	require.Equal(t, 1, job.SuccessCount)
	require.Equal(t, 0, job.FailedCount)

	otherTenantCtx := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	rows, err := svc.List(otherTenantCtx, "tenant-b", TrackingSyncJobQuery{Limit: 50})
	require.NoError(t, err)
	require.Len(t, rows, 0)
}

func TestTrackingSyncJobService_PartialFailedAndRetry(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "tracking_sync_job_retry")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	seedTrackingSyncJobFixtures(t, db, ctx)
	now := time.Now().UTC()
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:                     "wb-bad",
		TenantUUID:             "tenant-a",
		OrderID:                "order-bad",
		CarrierID:              "carrier-missing",
		ServiceCode:            "std",
		WaybillNo:              "WB-BAD",
		Status:                 "in_transit",
		PackageNo:              1,
		PackageKey:             "order-bad#1",
		OrderFulfillmentStatus: "partial_shipped",
		CreatedAt:              now,
		UpdatedAt:              now,
	}).Error)

	svc := NewTrackingSyncJobService(&app.Deps{DB: db})
	job, err := svc.CreateAndRun(ctx, "tenant-a", CreateTrackingSyncJobRequest{
		WaybillStatus: "in_transit",
		BatchLimit:    10,
		EventLimit:    5,
	})
	require.NoError(t, err)
	require.Equal(t, "partial_failed", job.Status)
	require.Equal(t, 2, job.TotalWaybills)
	require.Equal(t, 1, job.SuccessCount)
	require.Equal(t, 1, job.FailedCount)

	retry, err := svc.Retry(ctx, "tenant-a", job.ID)
	require.NoError(t, err)
	require.NotEqual(t, job.ID, retry.ID)
}

func TestTrackingSyncJobService_BulkSyncPressureSmoke(t *testing.T) {
	cases := []struct {
		name  string
		count int
	}{
		{name: "batch_100", count: 100},
		{name: "batch_500", count: 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTrackingSyncJobDB(t, "tracking_sync_job_pressure_"+tc.name)
			ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
			seedTrackingSyncJobFixtures(t, db, ctx)
			now := time.Now().UTC()
			for i := 0; i < tc.count; i++ {
				require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
					ID:                     "wb-pressure-" + tc.name + "-" + strconv.Itoa(i),
					TenantUUID:             "tenant-a",
					OrderID:                "order-pressure-" + strconv.Itoa(i),
					CarrierID:              "carrier-a",
					ServiceCode:            "std",
					WaybillNo:              "WB-PRESSURE-" + tc.name + "-" + strconv.Itoa(i),
					Status:                 "in_transit",
					PackageNo:              1,
					PackageKey:             "order-pressure-" + strconv.Itoa(i) + "#1",
					OrderFulfillmentStatus: "partial_shipped",
					CreatedAt:              now,
					UpdatedAt:              now,
				}).Error)
			}
			svc := NewTrackingSyncJobService(&app.Deps{DB: db})
			started := time.Now()
			job, err := svc.CreateAndRun(ctx, "tenant-a", CreateTrackingSyncJobRequest{
				CarrierID:     "carrier-a",
				WaybillStatus: "in_transit",
				BatchLimit:    tc.count + 10,
				EventLimit:    10,
			})
			elapsed := time.Since(started)
			require.NoError(t, err)
			require.NotNil(t, job)
			require.Equal(t, "success", job.Status)
			expectedTotal := tc.count + 1 // seeded wb-ok + pressure dataset
			if expectedTotal > 500 {
				expectedTotal = 500 // repository默认单次列表上限
			}
			require.Equal(t, expectedTotal, job.TotalWaybills)
			require.Equal(t, expectedTotal, job.SuccessCount)
			require.Equal(t, 0, job.FailedCount)
			t.Logf("pressure case=%s waybills=%d elapsed_ms=%d", tc.name, job.TotalWaybills, elapsed.Milliseconds())
		})
	}
}

func setupTrackingSyncJobDB(t *testing.T, name string) *gorm.DB {
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
		description TEXT,
		source TEXT NOT NULL,
		payload JSON,
		occurred_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_event_key ON logistics_tracking_events(tenant_uuid, waybill_no, event_id)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_tracking_sync_jobs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		carrier_id TEXT,
		provider TEXT,
		waybill_status TEXT,
		status TEXT NOT NULL,
		batch_limit INTEGER NOT NULL DEFAULT 20,
		event_limit INTEGER NOT NULL DEFAULT 20,
		total_waybills INTEGER NOT NULL DEFAULT 0,
		success_count INTEGER NOT NULL DEFAULT 0,
		failed_count INTEGER NOT NULL DEFAULT 0,
		appended_count INTEGER NOT NULL DEFAULT 0,
		replayed_count INTEGER NOT NULL DEFAULT 0,
		p95_latency_ms INTEGER NOT NULL DEFAULT 0,
		last_error TEXT,
		cancelled_reason TEXT,
		started_at DATETIME,
		finished_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_tracking_sync_schedules (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		cron_expr TEXT NOT NULL,
		carrier_id TEXT,
		waybill_status TEXT,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		max_concurrency INTEGER NOT NULL DEFAULT 1,
		dedupe_window_sec INTEGER NOT NULL DEFAULT 300,
		batch_limit INTEGER NOT NULL DEFAULT 20,
		event_limit INTEGER NOT NULL DEFAULT 20,
		last_triggered_at DATETIME,
		next_trigger_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_gateway_failure_events (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		carrier_id TEXT,
		waybill_id TEXT,
		waybill_no TEXT,
		provider TEXT,
		source_job_id TEXT,
		error_class TEXT NOT NULL,
		error_code TEXT,
		error_message TEXT,
		status TEXT NOT NULL,
		retry_count INTEGER NOT NULL DEFAULT 0,
		next_retry_at DATETIME,
		circuit_open_till DATETIME,
		recovered_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_gateway_usages (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		carrier_id TEXT,
		provider TEXT,
		source_job_id TEXT,
		request_count INTEGER NOT NULL DEFAULT 0,
		success_count INTEGER NOT NULL DEFAULT 0,
		failed_count INTEGER NOT NULL DEFAULT 0,
		unit_price NUMERIC NOT NULL DEFAULT 0,
		cost_amount NUMERIC NOT NULL DEFAULT 0,
		quota_consumed INTEGER NOT NULL DEFAULT 0,
		window_start_at DATETIME,
		window_end_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_gateway_usage_job ON logistics_gateway_usages(tenant_uuid, source_job_id)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_exception_orchestration_rules (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		trigger_event TEXT NOT NULL,
		action TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 100,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_orchestration_rule ON logistics_exception_orchestration_rules(tenant_uuid, name)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_exception_orchestration_runs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		rule_id TEXT NOT NULL,
		waybill_id TEXT,
		waybill_no TEXT,
		trigger TEXT NOT NULL,
		result TEXT NOT NULL,
		message TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_address_validations (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		request_key TEXT NOT NULL,
		waybill_id TEXT,
		waybill_no TEXT,
		raw_address TEXT NOT NULL,
		normalized TEXT,
		reachable BOOLEAN NOT NULL DEFAULT 1,
		risk_level TEXT NOT NULL DEFAULT 'low',
		suggestion TEXT,
		need_manual_review BOOLEAN NOT NULL DEFAULT 0,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_address_validation ON logistics_address_validations(tenant_uuid, request_key)`).Error)
	return db
}

func seedTrackingSyncJobFixtures(t *testing.T, db *gorm.DB, ctx context.Context) {
	t.Helper()
	now := time.Now().UTC()
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Carrier{
		ID:         "carrier-a",
		TenantUUID: "tenant-a",
		Name:       "自营",
		Code:       "self",
		Type:       "self",
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:                     "wb-ok",
		TenantUUID:             "tenant-a",
		OrderID:                "order-ok",
		CarrierID:              "carrier-a",
		ServiceCode:            "std",
		WaybillNo:              "WB-OK",
		Status:                 "in_transit",
		PackageNo:              1,
		PackageKey:             "order-ok#1",
		OrderFulfillmentStatus: "partial_shipped",
		CreatedAt:              now,
		UpdatedAt:              now,
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&LogisticsModel.Waybill{
		ID:                     "wb-skip",
		TenantUUID:             "tenant-a",
		OrderID:                "order-skip",
		CarrierID:              "carrier-a",
		ServiceCode:            "std",
		WaybillNo:              "WB-SKIP",
		Status:                 "created",
		PackageNo:              1,
		PackageKey:             "order-skip#1",
		OrderFulfillmentStatus: "partial_shipped",
		CreatedAt:              now,
		UpdatedAt:              now,
	}).Error)
}
