package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRiskService_FalsePositiveReleaseFlow(t *testing.T) {
	db := setupRiskDB(t, "logistics_risk_release")
	svc := NewRiskService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	rule, err := svc.UpsertRule(ctx, "tenant-1", UpsertRiskRuleRequest{
		Name:       "疑似高风险地址",
		MatchField: "address",
		MatchMode:  "contains",
		Pattern:    "测试路",
		Decision:   "block",
		RiskLevel:  "high",
		Priority:   300,
	})
	require.NoError(t, err)
	require.NotEmpty(t, rule.ID)

	first, err := svc.Evaluate(ctx, "tenant-1", EvaluateRiskRequest{
		WaybillID:       "wb-1",
		WaybillNo:       "WB-1",
		RecipientName:   "张三",
		RecipientPhone:  "13800138000",
		DestinationLine: "上海市浦东新区测试路88号",
	})
	require.NoError(t, err)
	require.True(t, first.Blocked)
	require.Equal(t, "block", first.Decision)
	require.NotEmpty(t, first.Hits)

	released, err := svc.Release(ctx, "tenant-1", first.Hits[0].ID, ReleaseRiskHitRequest{
		OperatorID: "qa",
		Reason:     "误拦截豁免",
	})
	require.NoError(t, err)
	require.Equal(t, "released", released.Status)

	second, err := svc.Evaluate(ctx, "tenant-1", EvaluateRiskRequest{
		WaybillID:       "wb-2",
		WaybillNo:       "WB-2",
		RecipientName:   "张三",
		RecipientPhone:  "13800138000",
		DestinationLine: "上海市浦东新区测试路88号",
	})
	require.NoError(t, err)
	require.False(t, second.Blocked)
	require.Equal(t, "allow", second.Decision)
	require.Equal(t, "qa", second.ReleasedBy)
}

func TestRiskService_TenantIsolation(t *testing.T) {
	db := setupRiskDB(t, "logistics_risk_tenant")
	svc := NewRiskService(&app.Deps{DB: db})
	ctx1 := authx.ContextWithTenantUUID(context.Background(), "tenant-1")
	ctx2 := authx.ContextWithTenantUUID(context.Background(), "tenant-2")

	_, err := svc.UpsertRule(ctx1, "tenant-1", UpsertRiskRuleRequest{
		Name:       "tenant1-rule",
		MatchField: "phone",
		MatchMode:  "exact",
		Pattern:    "13800138000",
		Decision:   "review",
	})
	require.NoError(t, err)

	rows2, err := svc.ListRules(ctx2, "tenant-2")
	require.NoError(t, err)
	require.Len(t, rows2, 0)

	result1, err := svc.Evaluate(ctx1, "tenant-1", EvaluateRiskRequest{
		WaybillID:      "wb-1",
		RecipientPhone: "13800138000",
	})
	require.NoError(t, err)
	require.NotEmpty(t, result1.Hits)

	_, err = svc.Release(ctx2, "tenant-2", result1.Hits[0].ID, ReleaseRiskHitRequest{OperatorID: "tenant2"})
	require.Error(t, err)
}

func setupRiskDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_risk_rules (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		match_field TEXT NOT NULL,
		match_mode TEXT NOT NULL,
		pattern TEXT NOT NULL,
		decision TEXT NOT NULL,
		risk_level TEXT NOT NULL,
		priority INTEGER NOT NULL,
		enabled BOOLEAN NOT NULL,
		description TEXT,
		rule_config JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_blacklists (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		entry_type TEXT NOT NULL,
		recipient_name TEXT,
		recipient_phone TEXT,
		address_line TEXT,
		reason TEXT,
		status TEXT NOT NULL,
		expires_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_risk_hits (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT NOT NULL,
		rule_id TEXT NOT NULL,
		blacklist_id TEXT NOT NULL,
		source TEXT NOT NULL,
		decision TEXT NOT NULL,
		risk_level TEXT NOT NULL,
		fingerprint TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL,
		released_by TEXT,
		release_reason TEXT,
		released_at DATETIME,
		payload JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	return db
}
