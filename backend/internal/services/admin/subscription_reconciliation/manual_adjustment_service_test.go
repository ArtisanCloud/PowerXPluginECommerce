package subscription_reconciliation

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceAdjustDelta_UpdatesAmountsAndStatus(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_adjust")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "c3218a03-fcc4-4c7c-9eeb-bf5ad17e3f21"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	batch, err := svc.CreateBatch(ctx, tenantUUID, CreateBatchInput{
		BillingCycle: "2026-03-31",
		RunType:      "daily",
		Samples: []ReconciliationSample{
			{SubscriptionRef: "sub-adjust", BillRef: "bill-adjust", ExpectedAmountMinor: 100, ActualAmountMinor: 80, ReasonCode: "init"},
		},
	})
	require.NoError(t, err)

	deltas, err := svc.ListDeltas(ctx, tenantUUID, DeltaListQuery{BatchID: batch.ID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, deltas, 1)

	updated, err := svc.AdjustDelta(ctx, tenantUUID, AdjustDeltaInput{
		DeltaID:             deltas[0].ID,
		ExpectedAmountMinor: 120,
		ActualAmountMinor:   100,
		ReasonCode:          "manual_adjust",
		Note:                "operator fixed amount",
	})
	require.NoError(t, err)
	require.Equal(t, int64(120), updated.ExpectedAmountMinor)
	require.Equal(t, int64(100), updated.ActualAmountMinor)
	require.Equal(t, int64(20), updated.DeltaAmountMinor)
	require.Equal(t, "manual_adjust", updated.ReasonCode)
	require.Equal(t, "processing", updated.Status)
}
