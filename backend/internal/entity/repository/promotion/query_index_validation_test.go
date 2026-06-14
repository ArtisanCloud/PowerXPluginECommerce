package promotion

import (
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPromotionQueryIndexValidation_CampaignAndSnapshot(t *testing.T) {
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:promotion_index_validation?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS promotion_campaigns (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			promotion_type TEXT NOT NULL,
			condition_rule TEXT NOT NULL,
			scope_rule TEXT NOT NULL,
			action_rule TEXT NOT NULL,
			stacking_rule TEXT NOT NULL,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_promotion_snapshots (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			currency TEXT NOT NULL,
			base_total_minor INTEGER NOT NULL,
			promotion_discount_minor INTEGER NOT NULL,
			after_promotion_total_minor INTEGER NOT NULL,
			applied_promotions TEXT NOT NULL,
			rejected_promotions TEXT NOT NULL,
			line_allocations TEXT NOT NULL,
			priced_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_promotion_campaign_active ON promotion_campaigns(tenant_uuid, status, valid_from, valid_to)`,
		`CREATE INDEX IF NOT EXISTS idx_promotion_snapshot_tenant_order ON order_promotion_snapshots(tenant_uuid, order_id)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}

	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO promotion_campaigns
		(id, tenant_uuid, code, name, promotion_type, condition_rule, scope_rule, action_rule, stacking_rule, valid_from, valid_to, status, created_at, updated_at)
		VALUES ('promo-1', 'tenant-1', 'P1', 'promo', 'amount_off', '{}', '{"scope_type":"all"}', '{"discount_amount_minor":100}', '{}', ?, ?, 'active', ?, ?)`,
		now.Add(-time.Hour), now.Add(time.Hour), now, now).Error)
	require.NoError(t, db.Exec(`INSERT INTO order_promotion_snapshots
		(id, tenant_uuid, order_id, currency, base_total_minor, promotion_discount_minor, after_promotion_total_minor, applied_promotions, rejected_promotions, line_allocations, priced_at, created_at, updated_at)
		VALUES ('snap-1', 'tenant-1', 'ord-1', 'CNY', 1000, 100, 900, '[]', '[]', '[]', ?, ?, ?)`, now, now, now).Error)

	campaignPlan := promotionExplainDetails(t, db, `EXPLAIN QUERY PLAN SELECT * FROM promotion_campaigns WHERE tenant_uuid = ? AND status = ? AND valid_from <= ? AND valid_to >= ?`, "tenant-1", "active", now, now)
	snapshotPlan := promotionExplainDetails(t, db, `EXPLAIN QUERY PLAN SELECT * FROM order_promotion_snapshots WHERE tenant_uuid = ? AND order_id = ?`, "tenant-1", "ord-1")

	require.Contains(t, strings.ToLower(campaignPlan), "idx_promotion_campaign_active")
	require.Contains(t, strings.ToLower(snapshotPlan), "idx_promotion_snapshot_tenant_order")
}

func promotionExplainDetails(t *testing.T, db *gorm.DB, sql string, args ...any) string {
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
