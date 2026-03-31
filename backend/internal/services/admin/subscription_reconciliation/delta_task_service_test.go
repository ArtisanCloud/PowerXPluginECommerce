package subscription_reconciliation

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceCreateDeltaTask_UniqueOpenTask(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_task")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "42b8f894-9ba5-4df2-93c8-5c460f08f7d2"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	batch, err := svc.CreateBatch(ctx, tenantUUID, CreateBatchInput{
		BillingCycle: "2026-03-31",
		RunType:      "daily",
		Samples:      []ReconciliationSample{{SubscriptionRef: "sub-1", BillRef: "bill-1", ExpectedAmountMinor: 100, ActualAmountMinor: 0}},
	})
	require.NoError(t, err)
	deltas, err := svc.ListDeltas(ctx, tenantUUID, DeltaListQuery{BatchID: batch.ID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, deltas, 1)

	task, err := svc.CreateDeltaTask(ctx, tenantUUID, CreateDeltaTaskInput{
		DeltaID:  deltas[0].ID,
		Assignee: "ops-001",
		SLALevel: "high",
		Note:     "create task",
	})
	require.NoError(t, err)
	require.Equal(t, "pending", task.Status)
	require.NotNil(t, task.SLADeadline)

	_, err = svc.CreateDeltaTask(ctx, tenantUUID, CreateDeltaTaskInput{
		DeltaID:  deltas[0].ID,
		Assignee: "ops-002",
		SLALevel: "high",
	})
	require.ErrorIs(t, err, ErrTaskAlreadyOpen)

	closed, err := svc.CloseTask(ctx, tenantUUID, CloseTaskInput{TaskID: task.ID, Resolution: "fixed", ResolutionNote: "done"})
	require.NoError(t, err)
	require.Equal(t, "closed", closed.Status)
}
