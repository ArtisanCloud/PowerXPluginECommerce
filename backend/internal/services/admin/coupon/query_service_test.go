package coupon

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQueryService_ListAssetsAndUsageLogs(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_query?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	setupCouponQueryTables(t, db)

	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO coupon_assets (id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"asset-1", "tenant-1", "tpl-1", "u-1", "PROMO-001", "redeemed", "ord-1", now, now).Error)
	require.NoError(t, db.Exec(`INSERT INTO coupon_usage_logs (id, tenant_uuid, asset_id, order_id, action, action_reason, idempotency_key, request_id, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"log-1", "tenant-1", "asset-1", "ord-1", "redeem", "payment_success", "redeem:ord-1:asset-1", "req-1", "payment_callback", now).Error)

	svc := NewQueryService(&app.Deps{DB: db})
	assets, err := svc.ListAssets(ctx, "tenant-1", AssetQueryFilter{CouponCode: "PROMO"})
	require.NoError(t, err)
	require.Equal(t, int64(1), assets.Total)
	require.Len(t, assets.Items, 1)

	logs, err := svc.ListUsageLogs(ctx, "tenant-1", UsageLogQueryFilter{CouponCode: "PROMO", Action: "redeem"})
	require.NoError(t, err)
	require.Equal(t, int64(1), logs.Total)
	require.Len(t, logs.Items, 1)
	require.Equal(t, "PROMO-001", logs.Items[0].CouponCode)
	require.Equal(t, "u-1", logs.Items[0].UserID)
}

func setupCouponQueryTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
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
			created_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
