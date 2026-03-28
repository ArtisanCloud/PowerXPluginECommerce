package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestExceptionOrchestrationService_RuleAndRun(t *testing.T) {
	db := setupTrackingSyncJobDB(t, "logistics_orchestration")
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	seedTrackingSyncJobFixtures(t, db, ctx)
	svc := NewExceptionOrchestrationService(&app.Deps{DB: db})

	rule, err := svc.UpsertRule(ctx, "tenant-a", UpsertExceptionRuleRequest{
		Name:         "delay-auto",
		TriggerEvent: "delay",
		Action:       "auto_compensate",
		Priority:     120,
		Enabled:      ptrBool(true),
	})
	require.NoError(t, err)
	require.NotEmpty(t, rule.ID)

	run, err := svc.Execute(ctx, "tenant-a", ExecuteExceptionRuleRequest{
		RuleID:    rule.ID,
		WaybillNo: "WB-OK",
		Trigger:   "delay",
	})
	require.NoError(t, err)
	require.Equal(t, "success", run.Result)

	runs, err := svc.ListRuns(ctx, "tenant-a", "WB-OK", 20)
	require.NoError(t, err)
	require.NotEmpty(t, runs)

	otherTenantCtx := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	otherRuns, err := svc.ListRuns(otherTenantCtx, "tenant-b", "WB-OK", 20)
	require.NoError(t, err)
	require.Len(t, otherRuns, 0)
}

func ptrBool(v bool) *bool { return &v }
