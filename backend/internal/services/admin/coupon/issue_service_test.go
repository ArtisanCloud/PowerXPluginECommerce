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

func TestIssueService_IssueCoupons(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_issue?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	setupCouponIssueTables(t, db)

	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO coupon_templates (id, tenant_uuid, code, name, coupon_type, threshold_rule, scope_rule, stacking_rule, refund_rule, valid_from, valid_to, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"tpl-1", "tenant-1", "PROMO", "活动券", "fixed", `{}`, `{}`, `{}`, `{}`, now.Add(-time.Hour), now.Add(24*time.Hour), "active", now, now).Error)

	svc := NewIssueService(&app.Deps{DB: db})
	res, err := svc.Issue(ctx, IssueInput{
		TenantUUID:      "tenant-1",
		TemplateID:      "tpl-1",
		UserIDs:         []string{"u-1", "u-2"},
		QuantityPerUser: 2,
	})
	require.NoError(t, err)
	require.Equal(t, 4, res.Issued)
	require.Len(t, res.AssetIDs, 4)

	var total int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM coupon_assets WHERE tenant_uuid = ? AND template_id = ?`, "tenant-1", "tpl-1").Scan(&total).Error)
	require.Equal(t, int64(4), total)

	var issueLogs int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND action = ? AND action_reason = ?`, "tenant-1", "issue", "manual_issue").Scan(&issueLogs).Error)
	require.Equal(t, int64(4), issueLogs)
}

func setupCouponIssueTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS coupon_templates (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			coupon_type TEXT NOT NULL,
			threshold_rule TEXT,
			scope_rule TEXT,
			stacking_rule TEXT,
			refund_rule TEXT,
			valid_from DATETIME,
			valid_to DATETIME,
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
			created_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
