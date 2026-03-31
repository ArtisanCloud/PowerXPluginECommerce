package subscription_reconciliation

import (
	"context"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceDashboard_AggregatesWithFilters(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_dashboard")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "69f7695d-38e9-4132-a188-dcb4fcb5dc9e"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_batches
		(id, tenant_uuid, billing_cycle, run_type, expected_amount_minor, actual_amount_minor, delta_amount_minor, delta_count, status, started_at, finished_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"batch-dashboard-1", tenantUUID, "2026-03-31", "daily", 1000, 900, 100, 2, "completed", now, now, now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_deltas
		(id, tenant_uuid, batch_id, subscription_ref, delta_type, risk_level, expected_amount_minor, actual_amount_minor, delta_amount_minor, reason_code, status, delta_fingerprint, detected_at, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"delta-dashboard-1", tenantUUID, "batch-dashboard-1", "sub-1", "missing_payment", "high", 500, 0, 500, "always_fail", "open", "fp-1", now, `{"channel":"app","plan":"pro","region":"CN"}`, now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_deltas
		(id, tenant_uuid, batch_id, subscription_ref, delta_type, risk_level, expected_amount_minor, actual_amount_minor, delta_amount_minor, reason_code, status, delta_fingerprint, detected_at, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"delta-dashboard-2", tenantUUID, "batch-dashboard-1", "sub-2", "amount_mismatch", "medium", 500, 400, 100, "success_at_2", "processing", "fp-2", now, `{"channel":"web","plan":"basic","region":"US"}`, now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_delta_tasks
		(id, tenant_uuid, delta_id, delta_fingerprint, assignee, priority, sla_level, status, created_at, updated_at, closed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"task-dashboard-1", tenantUUID, "delta-dashboard-1", "fp-1", "ops-1", "high", "high", "closed", now, now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_delta_tasks
		(id, tenant_uuid, delta_id, delta_fingerprint, assignee, priority, sla_level, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"task-dashboard-2", tenantUUID, "delta-dashboard-2", "fp-2", "ops-2", "medium", "medium", "pending", now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_execution_logs
		(id, tenant_uuid, subscription_ref, action_type, attempt_no, executed_at, result, operator_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"log-dashboard-1", tenantUUID, "sub-1", "retry", 1, now, "failed", "system", now, now,
	).Error)
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_execution_logs
		(id, tenant_uuid, subscription_ref, action_type, attempt_no, executed_at, result, operator_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"log-dashboard-2", tenantUUID, "sub-2", "retry", 1, now, "success", "system", now, now,
	).Error)

	resp, err := svc.Dashboard(ctx, tenantUUID, DashboardQuery{
		From:          "2026-03-30",
		To:            "2026-03-31",
		Channel:       "app",
		Plan:          "pro",
		Region:        "CN",
		FailureReason: "always_fail",
	})
	require.NoError(t, err)
	require.Equal(t, 0.5, resp["deltaRate"])
	require.Equal(t, 0.5, resp["recoveryRate"])
	require.Equal(t, int64(1), resp["pendingTasks"])
	types, ok := resp["byDeltaType"].([]DashboardDeltaTypeCount)
	require.True(t, ok)
	require.Len(t, types, 1)
	require.Equal(t, "missing_payment", types[0].Type)
}
