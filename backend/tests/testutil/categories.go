package testutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// SeedProductCategories inserts minimal category rows required by product tests.
// Note: tests use deterministic IDs like "cat-a" instead of UUIDs.
func SeedProductCategories(t *testing.T, db *gorm.DB, tenantUUID string, categoryIDs ...string) {
	t.Helper()
	if db == nil {
		return
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	require.NotEmpty(t, tenantUUID)

	for _, id := range categoryIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		// Upsert-ish: ignore duplicate insert errors by trying update if insert fails.
		res := db.Exec(
			`INSERT INTO product_categories (
				id, tenant_uuid, parent_id, code, display_name, alias_slug, path, level, sort_order, status, template_id, created_at, updated_at, deleted_at
			) VALUES (
				?, ?, NULL, ?, ?, ?, ?, 1, 0, 'enabled', NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL
			)`,
			id, tenantUUID, id, id, id, "/"+id+"/",
		)
		if res.Error == nil {
			continue
		}
		// Fallback update for Postgres unique constraints if present.
		_ = db.Exec(
			`UPDATE product_categories
			 SET code = ?, display_name = ?, alias_slug = ?, path = ?, status = 'enabled', updated_at = CURRENT_TIMESTAMP
			 WHERE tenant_uuid = ? AND id = ?`,
			id, id, id, "/"+id+"/", tenantUUID, id,
		).Error
	}
}
