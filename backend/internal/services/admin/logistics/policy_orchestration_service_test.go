package logistics

import (
	"context"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPolicyOrchestrationService_ConflictDetectionAccuracy(t *testing.T) {
	db := setupPolicyOrchestrationDB(t, "logistics_policy_orch_conflict")
	svc := NewPolicyOrchestrationService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-orch-a")

	_, err := svc.UpsertFlow(ctx, "tenant-orch-a", UpsertPolicyOrchestrationFlowRequest{
		Name:              "主链路发布策略",
		Status:            "active",
		ConflictRelations: []string{"carrier:sf", "zone:cn-east"},
		FlowDefinition: map[string]any{
			"nodes": []any{"check_stock", "route_plan", "publish"},
		},
	})
	require.NoError(t, err)

	preview, err := svc.PreviewConflicts(ctx, "tenant-orch-a", PolicyOrchestrationConflictPreviewRequest{
		Name:              "灰度策略",
		ConflictRelations: []string{"zone:cn-east", "channel:vip"},
	})
	require.NoError(t, err)
	require.True(t, preview.HasConflict)
	require.NotEmpty(t, preview.Items)
	require.Equal(t, "zone:cn-east", preview.Items[0].ConflictKeys[0])
}

func TestPolicyOrchestrationService_RollbackToTargetVersion(t *testing.T) {
	db := setupPolicyOrchestrationDB(t, "logistics_policy_orch_rollback")
	svc := NewPolicyOrchestrationService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-orch-b")

	flow, err := svc.UpsertFlow(ctx, "tenant-orch-b", UpsertPolicyOrchestrationFlowRequest{
		Name:              "履约策略编排",
		ConflictRelations: []string{"carrier:jd"},
		FlowDefinition:    map[string]any{"steps": []any{"risk_check", "dispatch"}},
		OperatorID:        "qa",
	})
	require.NoError(t, err)

	v1, err := svc.Publish(ctx, "tenant-orch-b", PublishPolicyOrchestrationRequest{
		FlowID:        flow.ID,
		RequestKey:    "publish-v1",
		ChangeSummary: "初版发布",
		OperatorID:    "qa",
	})
	require.NoError(t, err)
	require.Equal(t, "created", v1.Idempotency)

	_, err = svc.UpsertFlow(ctx, "tenant-orch-b", UpsertPolicyOrchestrationFlowRequest{
		ID:                flow.ID,
		Name:              flow.Name,
		ConflictRelations: []string{"carrier:jd", "warehouse:hz"},
		FlowDefinition:    map[string]any{"steps": []any{"risk_check", "route_opt", "dispatch"}},
		OperatorID:        "qa2",
	})
	require.NoError(t, err)

	v2, err := svc.Publish(ctx, "tenant-orch-b", PublishPolicyOrchestrationRequest{
		FlowID:        flow.ID,
		RequestKey:    "publish-v2",
		ChangeSummary: "增加路由优化",
		OperatorID:    "qa2",
		Force:         true,
	})
	require.NoError(t, err)
	require.NotEqual(t, v1.Version.ID, v2.Version.ID)

	rollback, err := svc.Rollback(ctx, "tenant-orch-b", RollbackPolicyOrchestrationRequest{
		FlowID:          flow.ID,
		TargetVersionID: v1.Version.ID,
		Reason:          "回滚验证",
		OperatorID:      "qa3",
	})
	require.NoError(t, err)
	require.Equal(t, "rolled_back", rollback.Version.Status)
	require.Equal(t, v1.Version.ID, rollback.Version.RolledBackFromID)
	require.Equal(t, v1.Version.ID, rollback.Flow.PublishedVersionID)
}

func TestPolicyOrchestrationService_ConcurrentPublishProtection(t *testing.T) {
	db := setupPolicyOrchestrationDB(t, "logistics_policy_orch_concurrent")
	svc := NewPolicyOrchestrationService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-orch-c")

	flow, err := svc.UpsertFlow(ctx, "tenant-orch-c", UpsertPolicyOrchestrationFlowRequest{
		Name:              "并发发布保护策略",
		ConflictRelations: []string{"channel:web"},
		FlowDefinition:    map[string]any{"steps": []any{"prepare", "publish"}},
		OperatorID:        "qa",
	})
	require.NoError(t, err)

	const requestKey = "concurrent-publish-key"
	first, err := svc.Publish(ctx, "tenant-orch-c", PublishPolicyOrchestrationRequest{
		FlowID:        flow.ID,
		RequestKey:    requestKey,
		ChangeSummary: "并发发布压测",
		OperatorID:    "qa",
	})
	require.NoError(t, err)
	require.Equal(t, "created", first.Idempotency)

	second, err := svc.Publish(ctx, "tenant-orch-c", PublishPolicyOrchestrationRequest{
		FlowID:        flow.ID,
		RequestKey:    requestKey,
		ChangeSummary: "并发发布压测",
		OperatorID:    "qa",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", second.Idempotency)
	require.Equal(t, first.Version.ID, second.Version.ID)
}

func setupPolicyOrchestrationDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_policy_orchestration_flows (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 100,
		status TEXT NOT NULL DEFAULT 'draft',
		flow_definition JSON,
		conflict_relations JSON,
		gray_release_config JSON,
		published_version_id TEXT,
		description TEXT,
		created_by TEXT,
		updated_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_policy_orchestration_flow_name ON logistics_policy_orchestration_flows(tenant_uuid, name)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_policy_orchestration_versions (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		flow_id TEXT NOT NULL,
		version_no INTEGER NOT NULL DEFAULT 1,
		request_key TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'published',
		change_summary TEXT,
		snapshot JSON,
		conflict_report JSON,
		gray_release_plan JSON,
		rolled_back_from_id TEXT,
		published_at DATETIME,
		created_by TEXT,
		updated_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_policy_orchestration_version_no ON logistics_policy_orchestration_versions(tenant_uuid, flow_id, version_no)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_policy_orchestration_request_key ON logistics_policy_orchestration_versions(tenant_uuid, flow_id, request_key)`).Error)
	return db
}
