package subscription_reconciliation

import (
	"context"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceRunGovernance_RespectEscalationThreshold(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_governance_escalation")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "25895d40-1546-4e57-964d-69628e0e2249"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO subscription_reconciliation_governance_policies
		(id, tenant_uuid, enabled, retry_windows, escalation_threshold, notify_channels, version, effective_from, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?, ?, ?, ?, ?, ?)`,
		"policy-1",
		tenantUUID,
		`["1h","24h"]`,
		2,
		`["sms"]`,
		1,
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	).Error)

	_, err := svc.CreateBatch(ctx, tenantUUID, CreateBatchInput{
		BillingCycle: "2026-03-31",
		RunType:      "daily",
		Samples: []ReconciliationSample{
			{SubscriptionRef: "sub-always-fail", BillRef: "bill-1", ExpectedAmountMinor: 100, ActualAmountMinor: 0, ReasonCode: "always_fail"},
		},
	})
	require.NoError(t, err)

	resp, err := svc.RunGovernance(ctx, tenantUUID, RunGovernanceInput{BillingCycle: "2026-03-31"})
	require.NoError(t, err)
	require.Equal(t, 1, resp["retryTotal"])
	require.Equal(t, 0, resp["retrySucceeded"])
	require.Equal(t, 2, resp["retryFailed"])
	require.Equal(t, 1, resp["escalated"])

	var retryLogs int64
	require.NoError(t, db.Table("subscription_reconciliation_execution_logs").
		Where("tenant_uuid = ? AND action_type = ?", tenantUUID, "retry").
		Count(&retryLogs).Error)
	require.Equal(t, int64(2), retryLogs)
}
