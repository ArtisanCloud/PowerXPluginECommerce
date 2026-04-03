package coupon

import (
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQueryIndexValidation_UsageAndSnapshot(t *testing.T) {
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_index_validation?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS coupon_usage_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			order_id TEXT,
			action TEXT NOT NULL,
			action_reason TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_coupon_snapshots (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			currency TEXT NOT NULL,
			base_total_minor INTEGER NOT NULL,
			discount_total_minor INTEGER NOT NULL,
			payable_total_minor INTEGER NOT NULL,
			line_allocations TEXT,
			applied_coupons TEXT,
			rejected_coupons TEXT,
			priced_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_tenant_order_created ON coupon_usage_logs(tenant_uuid, order_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_snapshot_tenant_order ON order_coupon_snapshots(tenant_uuid, order_id)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}

	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO coupon_usage_logs (id, tenant_uuid, asset_id, order_id, action, action_reason, idempotency_key, created_at)
		VALUES ('u1', 'tenant-1', 'asset-1', 'ord-1', 'redeem', 'payment_success', 'k1', ?)`, now).Error)
	require.NoError(t, db.Exec(`INSERT INTO order_coupon_snapshots (id, tenant_uuid, order_id, currency, base_total_minor, discount_total_minor, payable_total_minor, priced_at, created_at, updated_at)
		VALUES ('s1', 'tenant-1', 'ord-1', 'CNY', 1000, 100, 900, ?, ?, ?)`, now, now, now).Error)

	usagePlan := explainDetails(t, db, `EXPLAIN QUERY PLAN SELECT * FROM coupon_usage_logs WHERE tenant_uuid = ? AND order_id = ? ORDER BY created_at DESC`, "tenant-1", "ord-1")
	snapshotPlan := explainDetails(t, db, `EXPLAIN QUERY PLAN SELECT * FROM order_coupon_snapshots WHERE tenant_uuid = ? AND order_id = ?`, "tenant-1", "ord-1")

	require.Contains(t, strings.ToLower(usagePlan), "idx_usage_tenant_order_created")
	require.Contains(t, strings.ToLower(snapshotPlan), "idx_snapshot_tenant_order")
}

func explainDetails(t *testing.T, db *gorm.DB, sql string, args ...any) string {
	t.Helper()
	type planRow struct {
		ID     int
		Parent int
		NotUse int
		Detail string
	}
	var rows []planRow
	require.NoError(t, db.Raw(sql, args...).Scan(&rows).Error)
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		parts = append(parts, row.Detail)
	}
	return strings.Join(parts, " | ")
}
