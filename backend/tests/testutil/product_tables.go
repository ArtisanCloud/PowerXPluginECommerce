package testutil

import (
	"testing"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// EnsureProductTables recreates the SPU/SKU tables required for product onboarding tests.
func EnsureProductTables(t *testing.T, db *gorm.DB, isPostgres bool) {
	t.Helper()

	jsonType := "TEXT"
	if isPostgres {
		jsonType = "JSONB"
	}

	tableNames := []string{
		basemodels.TableProductSpus,
		basemodels.TableProductSpuVersions,
		basemodels.TableProductSpuApprovals,
		basemodels.TableProductSpuChannels,
		basemodels.TableProductSpuAuditLogs,
		basemodels.TableProductSkus,
		basemodels.TableProductSkuAttributes,
	}
	for _, table := range tableNames {
		_ = db.Exec("DROP TABLE IF EXISTS " + table).Error
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS product_spus (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			category_id TEXT NOT NULL,
			category_path TEXT NOT NULL,
			brand_id TEXT,
			default_locale TEXT NOT NULL,
			status TEXT NOT NULL,
			current_version_id TEXT,
			tags TEXT,
			responsible_user TEXT,
			channels_summary ` + jsonType + `,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_versions (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			version_number INTEGER NOT NULL,
			status TEXT NOT NULL,
			payload ` + jsonType + `,
			diff_summary ` + jsonType + `,
			submitted_by TEXT,
			submitted_at TIMESTAMP,
			approved_by TEXT,
			approved_at TIMESTAMP,
			rollback_source_version TEXT,
			audit_log_id TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_approvals (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			version_id TEXT NOT NULL,
			chain_order INTEGER NOT NULL,
			role TEXT NOT NULL,
			status TEXT NOT NULL,
			comment TEXT,
			sla_due_at TIMESTAMP,
			acted_by TEXT,
			acted_at TIMESTAMP,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_channels (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			availability TEXT NOT NULL,
			publish_at TIMESTAMP,
			withdraw_at TIMESTAMP,
			content_override ` + jsonType + `,
			audit_state TEXT,
			last_feedback ` + jsonType + `,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_audit_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload ` + jsonType + `,
			operator TEXT,
			created_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS product_skus (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			sku_code TEXT NOT NULL,
			barcode TEXT,
			status TEXT NOT NULL,
			lifecycle_phase TEXT,
			min_order_qty INTEGER,
			spec_values ` + jsonType + `,
			spec_signature TEXT,
			default_values ` + jsonType + `,
			logistics ` + jsonType + `,
			price_refs ` + jsonType + `,
			tags TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS product_sku_attributes (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			sku_id TEXT NOT NULL,
			spec_id TEXT NOT NULL,
			spec_value_id TEXT NOT NULL,
			spec_name TEXT,
			value_name TEXT,
			display_order INTEGER,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
	}

	for _, stmt := range statements {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
