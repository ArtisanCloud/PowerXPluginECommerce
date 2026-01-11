package migrate

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

func ensurePricingRLSPolicies(ctx context.Context, db *gorm.DB) error {
	if db == nil || db.Dialector == nil || !strings.EqualFold(db.Dialector.Name(), "postgres") {
		return nil
	}
	type policy struct {
		table string
		name  string
	}
	targets := []policy{
		{models.S(models.TablePricebooks), "pricebook_tenant_rls"},
		{models.S(models.TablePricebookVersions), "pricebook_version_tenant_rls"},
		{models.S(models.TablePricebookScopes), "pricebook_scope_tenant_rls"},
		{models.S(models.TablePricebookItems), "pricebook_item_tenant_rls"},
		{models.S(models.TablePricebookAuditLogs), "pricebook_audit_log_tenant_rls"},
	}
	for _, p := range targets {
		if err := db.WithContext(ctx).Exec(fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", p.table)).Error; err != nil {
			return err
		}
		exists, err := rlsPolicyExists(ctx, db, p.table, p.name)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		using := "tenant_uuid::text = current_setting('app.tenant_uuid', true)"
		stmt := fmt.Sprintf(`CREATE POLICY %s ON %s USING (%s) WITH CHECK (%s)`, p.name, p.table, using, using)
		if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}
