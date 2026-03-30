package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestComplianceKBService_DiffAccuracy(t *testing.T) {
	db := setupBillingDB(t, "logistics_compliance_kb_diff")
	ensureComplianceKBTables(t, db)
	svc := NewComplianceKBService(&app.Deps{DB: db})
	customsSvc := NewCustomsRuleService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-compliance-a")

	pack, err := customsSvc.UpsertPack(ctx, "tenant-compliance-a", UpsertCustomsRulePackRequest{
		Name:        "US 合规规则",
		CountryCode: "US",
		Status:      "active",
	})
	require.NoError(t, err)

	v1, err := customsSvc.PublishVersion(ctx, "tenant-compliance-a", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "A", Name: "A", Field: "declared_value", Operator: "gt", Value: 100, RiskLevel: "high", Suggestion: "review", Enabled: true},
			{Code: "B", Name: "B", Field: "tax_no", Operator: "missing", Value: "", RiskLevel: "medium", Suggestion: "fill_tax_no", Enabled: true},
		},
		Status: "published",
	})
	require.NoError(t, err)
	v2, err := customsSvc.PublishVersion(ctx, "tenant-compliance-a", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "A", Name: "A", Field: "declared_value", Operator: "gt", Value: 200, RiskLevel: "high", Suggestion: "review", Enabled: true},       // changed
			{Code: "C", Name: "C", Field: "document_count", Operator: "lte", Value: 0, RiskLevel: "high", Suggestion: "manual_review", Enabled: true}, // added
		},
		Status: "published",
	})
	require.NoError(t, err)

	p1, err := svc.SyncPolicy(ctx, "tenant-compliance-a", SyncComplianceKBPolicyRequest{
		CountryCode:     "US",
		PackID:          pack.ID,
		SourceVersionID: v1.ID,
		PolicyVersion:   "v1",
	})
	require.NoError(t, err)
	p2, err := svc.SyncPolicy(ctx, "tenant-compliance-a", SyncComplianceKBPolicyRequest{
		CountryCode:     "US",
		PackID:          pack.ID,
		SourceVersionID: v2.ID,
		PolicyVersion:   "v2",
	})
	require.NoError(t, err)

	diff, err := svc.DiffPolicies(ctx, "tenant-compliance-a", DiffComplianceKBPolicyRequest{
		BasePolicyID:   p1.ID,
		TargetPolicyID: p2.ID,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"C"}, diff.AddedRuleCodes)
	require.ElementsMatch(t, []string{"B"}, diff.RemovedRuleCodes)
	require.ElementsMatch(t, []string{"A"}, diff.ChangedRuleCodes)
}

func TestComplianceKBService_GrayscaleControl(t *testing.T) {
	db := setupBillingDB(t, "logistics_compliance_kb_gray")
	ensureComplianceKBTables(t, db)
	svc := NewComplianceKBService(&app.Deps{DB: db})
	customsSvc := NewCustomsRuleService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-compliance-b")

	pack, err := customsSvc.UpsertPack(ctx, "tenant-compliance-b", UpsertCustomsRulePackRequest{
		Name:        "DE 合规规则",
		CountryCode: "DE",
		Status:      "active",
	})
	require.NoError(t, err)
	_, err = customsSvc.PublishVersion(ctx, "tenant-compliance-b", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "D1", Name: "D1", Field: "document_count", Operator: "lte", Value: 0, RiskLevel: "high", Suggestion: "manual_review", Enabled: true},
		},
		Status: "published",
	})
	require.NoError(t, err)

	row, err := svc.SyncPolicy(ctx, "tenant-compliance-b", SyncComplianceKBPolicyRequest{
		CountryCode:   "DE",
		PackID:        pack.ID,
		PolicyVersion: "v1",
	})
	require.NoError(t, err)

	published, err := svc.PublishPolicy(ctx, "tenant-compliance-b", row.ID, PublishComplianceKBPolicyRequest{
		RolloutPercent: 30,
		RolloutTenants: []string{"tenant-a", "tenant-b"},
		OperatorID:     "admin",
	})
	require.NoError(t, err)
	require.Equal(t, "grayscale", published.Status)
	require.Contains(t, string(published.RolloutScope), "rollout_percent")
	require.Contains(t, string(published.RolloutScope), "30")
}

func TestComplianceKBService_RollbackSafety(t *testing.T) {
	db := setupBillingDB(t, "logistics_compliance_kb_rollback")
	ensureComplianceKBTables(t, db)
	svc := NewComplianceKBService(&app.Deps{DB: db})
	customsSvc := NewCustomsRuleService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-compliance-c")

	pack, err := customsSvc.UpsertPack(ctx, "tenant-compliance-c", UpsertCustomsRulePackRequest{
		Name:        "JP 合规规则",
		CountryCode: "JP",
		Status:      "active",
	})
	require.NoError(t, err)
	v1, err := customsSvc.PublishVersion(ctx, "tenant-compliance-c", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "J1", Name: "J1", Field: "declared_value", Operator: "gt", Value: 200, RiskLevel: "high", Suggestion: "review", Enabled: true},
		},
		Status: "published",
	})
	require.NoError(t, err)
	v2, err := customsSvc.PublishVersion(ctx, "tenant-compliance-c", PublishCustomsRuleVersionRequest{
		PackID: pack.ID,
		Rules: []CustomsRuleDefinition{
			{Code: "J2", Name: "J2", Field: "hs_code", Operator: "missing", Value: "", RiskLevel: "medium", Suggestion: "fill_hs", Enabled: true},
		},
		Status: "published",
	})
	require.NoError(t, err)

	policyV1, err := svc.SyncPolicy(ctx, "tenant-compliance-c", SyncComplianceKBPolicyRequest{
		CountryCode:     "JP",
		PackID:          pack.ID,
		SourceVersionID: v1.ID,
		PolicyVersion:   "v1",
	})
	require.NoError(t, err)
	policyV2, err := svc.SyncPolicy(ctx, "tenant-compliance-c", SyncComplianceKBPolicyRequest{
		CountryCode:     "JP",
		PackID:          pack.ID,
		SourceVersionID: v2.ID,
		PolicyVersion:   "v2",
	})
	require.NoError(t, err)

	_, err = svc.PublishPolicy(ctx, "tenant-compliance-c", policyV2.ID, PublishComplianceKBPolicyRequest{RolloutPercent: 100})
	require.NoError(t, err)
	_, err = svc.PublishPolicy(ctx, "tenant-compliance-c", policyV1.ID, PublishComplianceKBPolicyRequest{RolloutPercent: 100})
	require.NoError(t, err)

	rows, err := svc.ListPolicies(ctx, "tenant-compliance-c", ComplianceKBQuery{CountryCode: "JP", Limit: 20})
	require.NoError(t, err)
	activeCount := 0
	for _, item := range rows {
		if item.Status == "active" {
			activeCount++
			require.Equal(t, policyV1.ID, item.ID)
		}
	}
	require.Equal(t, 1, activeCount)
}

func ensureComplianceKBTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	ensureCustomsRuleTables(t, db)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_compliance_kb_versions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		country_code TEXT NOT NULL,
		policy_version TEXT NOT NULL,
		source_pack_id TEXT,
		source_version_id TEXT,
		source_version_no INTEGER NOT NULL DEFAULT 0,
		country_rule_mapping JSON NOT NULL,
		effective_from DATETIME,
		effective_to DATETIME,
		rollout_scope JSON,
		status TEXT NOT NULL DEFAULT 'draft',
		notes TEXT,
		published_by TEXT,
		published_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_compliance_kb_version ON logistics_compliance_kb_versions(tenant_uuid, country_code, policy_version)`).Error)
}
