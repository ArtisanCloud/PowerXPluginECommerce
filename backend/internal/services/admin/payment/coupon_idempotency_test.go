package payment

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	agentpayments "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/agent/payments"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestCouponCallbackIdempotency_RedeemOnlyOnce(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newCouponIdempotencyDB(t)
	createCouponIdempotencyTables(t, db)

	const (
		tenant        = "tenant-test"
		orderID       = "ord-1"
		orderNo       = "O20260403001"
		transactionNo = "TX20260403001"
		assetID       = "asset-1"
	)

	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO orders (
		id, tenant_uuid, order_no, customer_id, channel, status, currency, subtotal_amount, total_amount, created_by_type, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, tenant, orderNo, "customer-1", "miniapp", "pending_payment", "CNY", int64(100), int64(100), "customer", now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO payment_transactions (
		tenant_uuid, transaction_no, order_id, order_no, provider_id, pay_method, amount_total, amount_currency, fee_amount, status, metadata, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenant, transactionNo, orderID, orderNo, int64(0), "wechat_jsapi", int64(100), "CNY", int64(0), "pending_payment", "{}", now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO coupon_assets (
		id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, reserved_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		assetID, tenant, "tpl-1", "customer-1", "CODE-1", "reserved", orderID, now, now, now,
	).Error)

	svc := agentpayments.NewTransactionService(&app.Deps{DB: db})
	require.True(t, svc.Ready())

	payload := map[string]any{
		"transaction_no": transactionNo,
		"trade_state":    "SUCCESS",
	}
	require.NoError(t, svc.HandleProviderCallback(ctx, tenant, 0, payload))
	require.NoError(t, svc.HandleProviderCallback(ctx, tenant, 0, payload))

	var orderStatus string
	require.NoError(t, db.Raw(`SELECT status FROM orders WHERE tenant_uuid = ? AND id = ?`, tenant, orderID).Scan(&orderStatus).Error)
	require.Equal(t, "paid", orderStatus)

	var couponStatus string
	require.NoError(t, db.Raw(`SELECT status FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, tenant, assetID).Scan(&couponStatus).Error)
	require.Equal(t, "redeemed", couponStatus)

	var redeemLogs int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND asset_id = ? AND action = ?`, tenant, assetID, "redeem").Scan(&redeemLogs).Error)
	require.Equal(t, int64(1), redeemLogs)
}

func newCouponIdempotencyDB(t *testing.T) *gorm.DB {
	t.Helper()
	safeName := strings.NewReplacer("/", "_", " ", "_", ":", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	return db
}

func createCouponIdempotencyTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_no TEXT NOT NULL,
			customer_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			status TEXT NOT NULL,
			currency TEXT NOT NULL,
			subtotal_amount INTEGER NOT NULL DEFAULT 0,
			total_amount INTEGER NOT NULL DEFAULT 0,
			created_by_type TEXT NOT NULL DEFAULT 'system',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			sku_id TEXT NOT NULL,
			qty INTEGER NOT NULL DEFAULT 1,
			unit_price INTEGER NOT NULL DEFAULT 0,
			line_amount INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME
		)`,
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
			metadata TEXT,
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
