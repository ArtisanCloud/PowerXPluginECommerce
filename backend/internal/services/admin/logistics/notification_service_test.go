package logistics

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNotificationService_RenderTemplateAndIdempotency(t *testing.T) {
	db := setupNotificationDB(t, "logistics_notification_render_idem")
	svc := NewNotificationService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-notify-1")

	_, err := svc.UpsertTemplate(ctx, "tenant-notify-1", UpsertNotificationTemplateRequest{
		Name:    "签收通知",
		Event:   "signed",
		Channel: "sms",
		Title:   "订单 {{order_id}} 已签收",
		Body:    "运单 {{waybill_no}} 已签收",
	})
	require.NoError(t, err)

	seedWaybill(t, db, "tenant-notify-1", "wb-n1", "order-n1", "WB-NO-1")
	result1, err := svc.Send(ctx, "tenant-notify-1", SendNotificationRequest{
		WaybillID:      "wb-n1",
		Event:          "signed",
		IdempotencyKey: "idem-001",
	})
	require.NoError(t, err)
	require.Equal(t, "created", result1.IdempotencyStatus)
	require.Equal(t, "sent", result1.Record.Status)
	require.Contains(t, result1.Record.RenderedTitle, "order-n1")
	require.Contains(t, result1.Record.RenderedBody, "WB-NO-1")

	result2, err := svc.Send(ctx, "tenant-notify-1", SendNotificationRequest{
		WaybillID:      "wb-n1",
		Event:          "signed",
		IdempotencyKey: "idem-001",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", result2.IdempotencyStatus)
	require.Equal(t, result1.Record.ID, result2.Record.ID)
}

func TestNotificationService_RetryFailedRecord(t *testing.T) {
	db := setupNotificationDB(t, "logistics_notification_retry")
	svc := NewNotificationService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-notify-2")

	_, err := svc.UpsertTemplate(ctx, "tenant-notify-2", UpsertNotificationTemplateRequest{
		Name:    "异常通知",
		Event:   "exception",
		Channel: "sms",
		Body:    "运单 {{waybill_no}} 异常，请处理",
	})
	require.NoError(t, err)
	seedWaybill(t, db, "tenant-notify-2", "wb-n2", "order-n2", "WB-NO-2")

	sendResult, err := svc.Send(ctx, "tenant-notify-2", SendNotificationRequest{
		WaybillID: "wb-n2",
		Event:     "exception",
		Payload: map[string]any{
			"simulate_fail": true,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "failed", sendResult.Record.Status)

	retried, err := svc.Retry(ctx, "tenant-notify-2", RetryNotificationRequest{
		RecordID: sendResult.Record.ID,
	})
	require.NoError(t, err)
	require.Equal(t, "sent", retried.Status)
	require.Equal(t, 2, retried.AttemptCount)
	require.NotNil(t, retried.SentAt)
}

func setupNotificationDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
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
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_notification_templates (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		event TEXT NOT NULL,
		channel TEXT NOT NULL DEFAULT 'sms',
		title TEXT,
		body TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_notify_tpl ON logistics_notification_templates(tenant_uuid, event, channel)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_notification_records (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		template_id TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		event TEXT NOT NULL,
		channel TEXT NOT NULL,
		status TEXT NOT NULL,
		attempt_count INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 3,
		idempotency_key TEXT NOT NULL,
		last_error TEXT,
		rendered_title TEXT,
		rendered_body TEXT,
		payload JSON,
		sent_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_notify_idem ON logistics_notification_records(tenant_uuid, idempotency_key)`).Error)
	return db
}

func seedWaybill(t *testing.T, db *gorm.DB, tenantUUID, waybillID, orderID, waybillNo string) {
	t.Helper()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)
	require.NoError(t, db.WithContext(ctx).Exec(
		`INSERT INTO logistics_waybills (
			id, tenant_uuid, order_id, carrier_id, service_code, waybill_no, package_no, package_key,
			status, created_at, updated_at
		) VALUES (?, ?, ?, 'carrier-seed', 'std', ?, 1, ?, 'created', ?, ?)`,
		waybillID,
		tenantUUID,
		orderID,
		waybillNo,
		orderID+"#1",
		time.Now().UTC(),
		time.Now().UTC(),
	).Error)
}
