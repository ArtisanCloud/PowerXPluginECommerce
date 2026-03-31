package subscription_reconciliation

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceRunGovernance_RetryAndRecovery(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_governance_retry")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "f6e384e8-59f7-42f4-a973-1667185f4ec2"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	_, err := svc.CreateBatch(ctx, tenantUUID, CreateBatchInput{
		BillingCycle: "2026-03-31",
		RunType:      "daily",
		Samples: []ReconciliationSample{
			{SubscriptionRef: "sub-recover", BillRef: "bill-1", ExpectedAmountMinor: 100, ActualAmountMinor: 0, ReasonCode: "success_at_2"},
			{SubscriptionRef: "sub-failed", BillRef: "bill-2", ExpectedAmountMinor: 200, ActualAmountMinor: 0, ReasonCode: "always_fail"},
		},
	})
	require.NoError(t, err)

	resp, err := svc.RunGovernance(ctx, tenantUUID, RunGovernanceInput{BillingCycle: "2026-03-31"})
	require.NoError(t, err)
	require.Equal(t, 2, resp["retryTotal"])
	require.Equal(t, 1, resp["retrySucceeded"])
	require.Equal(t, 5, resp["retryFailed"])
	require.Equal(t, 1, resp["escalated"])

	var total int64
	require.NoError(t, db.Table("subscription_reconciliation_execution_logs").Where("tenant_uuid = ?", tenantUUID).Count(&total).Error)
	require.Equal(t, int64(8), total)
}
