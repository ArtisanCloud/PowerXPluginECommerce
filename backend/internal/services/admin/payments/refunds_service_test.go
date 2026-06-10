package payments

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRefundService_CreateRefundReturnsCouponWhenTemplateAllows(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db := newRefundServiceDB(t)
	createRefundServiceTables(t, db)

	const (
		tenant        = "tenant-1"
		orderID       = "ord-1"
		transactionID = 1
		assetID       = "asset-1"
	)
	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO payment_transactions (
		id, tenant_uuid, transaction_no, order_id, order_no, provider_id, pay_method, amount_total, amount_currency, fee_amount, status, metadata, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		transactionID, tenant, "TX-1", orderID, "O-1", 0, "wechat_jsapi", int64(1000), "CNY", int64(0), "paid", "{}", now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO coupon_templates (
		id, tenant_uuid, code, name, coupon_type, threshold_rule, scope_rule, stacking_rule, refund_rule, valid_from, valid_to, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"tpl-1", tenant, "TPL1", "template", "amount", "{}", "{}", "{}", `{"return_coupon":true}`, now.Add(-time.Hour), now.Add(time.Hour), "active", now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO coupon_assets (
		id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, redeemed_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assetID, tenant, "tpl-1", "user-1", "C1", "redeemed", orderID, now, now, now,
	).Error)

	svc := NewRefundService(&app.Deps{DB: db})
	resp, err := svc.CreateRefund(ctx, tenant, "admin-1", transactionID, CreateRefundRequest{AmountMinor: 100, Reason: "after sales"})
	require.NoError(t, err)
	require.NotNil(t, resp)

	var status string
	require.NoError(t, db.Raw(`SELECT status FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, tenant, assetID).Scan(&status).Error)
	require.Equal(t, "refunded", status)
	var refundLogs int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND asset_id = ? AND action = ?`, tenant, assetID, "refund").Scan(&refundLogs).Error)
	require.Equal(t, int64(1), refundLogs)
}

func newRefundServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	safeName := strings.NewReplacer("/", "_", " ", "_", ":", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func createRefundServiceTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS payment_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			transaction_no TEXT NOT NULL,
			order_id TEXT,
			order_no TEXT,
			provider_id INTEGER NOT NULL DEFAULT 0,
			pay_method TEXT NOT NULL,
			amount_total INTEGER NOT NULL DEFAULT 0,
			amount_currency TEXT NOT NULL,
			fee_amount INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			completed_at DATETIME,
			failure_reason TEXT,
			risk_flag BOOLEAN NOT NULL DEFAULT 0,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payment_refunds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			transaction_id INTEGER NOT NULL,
			refund_no TEXT NOT NULL,
			refund_amount INTEGER NOT NULL DEFAULT 0,
			refund_currency TEXT NOT NULL,
			status TEXT NOT NULL,
			reason TEXT,
			completed_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_templates (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			coupon_type TEXT NOT NULL,
			threshold_rule TEXT NOT NULL,
			scope_rule TEXT NOT NULL,
			stacking_rule TEXT NOT NULL,
			refund_rule TEXT NOT NULL,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_assets (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			template_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			coupon_code TEXT NOT NULL,
			status TEXT NOT NULL,
			reserved_order_id TEXT,
			reserved_at DATETIME,
			redeemed_at DATETIME,
			refunded_at DATETIME,
			expired_at DATETIME,
			valid_from DATETIME,
			valid_to DATETIME,
			meta TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_usage_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			order_id TEXT,
			action TEXT NOT NULL,
			action_reason TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			request_id TEXT,
			created_by TEXT,
			created_at DATETIME,
			UNIQUE(tenant_uuid, action, idempotency_key)
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
