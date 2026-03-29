package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCustomsRuleService_VersionSwitching(t *testing.T) {
	db := setupBillingDB(t, "logistics_customs_rule_version")
	ensureCustomsRuleTables(t, db)
	svc := NewCustomsRuleService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-customs")

	pack, err := svc.UpsertPack(ctx, "tenant-customs", UpsertCustomsRulePackRequest{
		Name:             "US 清关规则",
		CountryCode:      "US",
		Status:           "active",
		DefaultRiskLevel: "low",
	})
	require.NoError(t, err)

	_, err = svc.PublishVersion(ctx, "tenant-customs", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "DECLARED_GT_100", Name: "申报金额过高", Field: "declared_value", Operator: "gt", Value: 100, RiskLevel: "high", Suggestion: "manual_review", Enabled: true},
		},
		Status:      "published",
		HitStrategy: "first_hit",
	})
	require.NoError(t, err)

	v2, err := svc.PublishVersion(ctx, "tenant-customs", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "DECLARED_GT_1000", Name: "超高申报金额", Field: "declared_value", Operator: "gt", Value: 1000, RiskLevel: "high", Suggestion: "manual_review", Enabled: true},
		},
		Status:      "published",
		HitStrategy: "first_hit",
	})
	require.NoError(t, err)
	require.Equal(t, 2, v2.VersionNo)

	result, err := svc.Precheck(ctx, "tenant-customs", CustomsPrecheckRequest{
		PackID:        pack.ID,
		CountryCode:   "US",
		WaybillNo:     "WB-CUSTOMS-1",
		DeclaredValue: 200,
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.VersionNo)
	require.Equal(t, "pass", result.Decision)
}

func TestCustomsRuleService_HitExplanationComplete(t *testing.T) {
	db := setupBillingDB(t, "logistics_customs_rule_explain")
	ensureCustomsRuleTables(t, db)
	svc := NewCustomsRuleService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-customs-exp")

	pack, err := svc.UpsertPack(ctx, "tenant-customs-exp", UpsertCustomsRulePackRequest{
		Name:        "EU 清关规则",
		CountryCode: "DE",
		Status:      "active",
	})
	require.NoError(t, err)

	_, err = svc.PublishVersion(ctx, "tenant-customs-exp", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "TAX_NO_REQUIRED", Name: "税号必填", Field: "tax_no", Operator: "missing", Value: "", RiskLevel: "high", Suggestion: "request_tax_no", Enabled: true},
		},
		Status:      "published",
		HitStrategy: "all_hit",
	})
	require.NoError(t, err)

	result, err := svc.Precheck(ctx, "tenant-customs-exp", CustomsPrecheckRequest{
		PackID:        pack.ID,
		CountryCode:   "DE",
		WaybillNo:     "WB-CUSTOMS-2",
		DeclaredValue: 80,
		TaxNo:         "",
	})
	require.NoError(t, err)
	require.Equal(t, "block", result.Decision)
	require.NotEmpty(t, result.MatchedRules)
	require.NotEmpty(t, result.MatchedRules[0].Code)
	require.NotEmpty(t, result.MatchedRules[0].Reason)
}

func TestCustomsRuleService_ManualReleaseForFalseIntercept(t *testing.T) {
	db := setupBillingDB(t, "logistics_customs_rule_release")
	ensureCustomsRuleTables(t, db)
	svc := NewCustomsRuleService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-customs-release")

	pack, err := svc.UpsertPack(ctx, "tenant-customs-release", UpsertCustomsRulePackRequest{
		Name:        "JP 清关规则",
		CountryCode: "JP",
		Status:      "active",
	})
	require.NoError(t, err)

	_, err = svc.PublishVersion(ctx, "tenant-customs-release", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "DOC_COUNT_ZERO", Name: "单据数量异常", Field: "document_count", Operator: "lte", Value: 0, RiskLevel: "high", Suggestion: "manual_review", Enabled: true},
		},
		Status:      "published",
		HitStrategy: "first_hit",
	})
	require.NoError(t, err)

	blocked, err := svc.Precheck(ctx, "tenant-customs-release", CustomsPrecheckRequest{
		PackID:        pack.ID,
		CountryCode:   "JP",
		WaybillNo:     "WB-CUSTOMS-3",
		DocumentCount: 0,
	})
	require.NoError(t, err)
	require.Equal(t, "block", blocked.Decision)

	released, err := svc.Precheck(ctx, "tenant-customs-release", CustomsPrecheckRequest{
		PackID:        pack.ID,
		CountryCode:   "JP",
		WaybillNo:     "WB-CUSTOMS-3",
		DocumentCount: 0,
		ManualRelease: true,
	})
	require.NoError(t, err)
	require.Equal(t, "manual_release", released.Decision)
}

func ensureCustomsRuleTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_customs_rule_packs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		country_code TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		strategy TEXT NOT NULL DEFAULT 'first_hit',
		default_risk_level TEXT NOT NULL DEFAULT 'low',
		description TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_customs_rule_pack ON logistics_customs_rule_packs(tenant_uuid, name)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_customs_rule_versions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		pack_id TEXT NOT NULL,
		version_no INTEGER NOT NULL DEFAULT 1,
		status TEXT NOT NULL DEFAULT 'draft',
		rules JSON NOT NULL,
		hit_strategy TEXT NOT NULL DEFAULT 'first_hit',
		risk_snapshot JSON,
		published_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_customs_rule_version ON logistics_customs_rule_versions(tenant_uuid, pack_id, version_no)`).Error)
}
