package subscription_reconciliation

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

func TestServiceCreateBatch_Idempotency(t *testing.T) {
	db := setupReconciliationDB(t, "subscription_reconciliation_idempotency")
	svc := NewService(&app.Deps{DB: db})
	tenantUUID := "3f1a98d8-7e56-4c1a-bcd4-49cb7f4e1a11"
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	input := CreateBatchInput{
		BillingCycle: "2026-03-31",
		RunType:      "daily",
		CreatedBy:    "ops-001",
		Samples: []ReconciliationSample{
			{SubscriptionRef: "sub-1", BillRef: "bill-1", ExpectedAmountMinor: 100, ActualAmountMinor: 0},
			{SubscriptionRef: "sub-2", BillRef: "bill-2", ExpectedAmountMinor: 0, ActualAmountMinor: 50},
			{SubscriptionRef: "sub-3", BillRef: "bill-3", ExpectedAmountMinor: 120, ActualAmountMinor: 100},
		},
	}

	batch1, err := svc.CreateBatch(ctx, tenantUUID, input)
	require.NoError(t, err)
	require.NotEmpty(t, batch1.ID)
	require.Equal(t, 220, int(batch1.ExpectedAmountMinor))
	require.Equal(t, 150, int(batch1.ActualAmountMinor))
	require.Equal(t, 3, batch1.DeltaCount)

	batch2, err := svc.CreateBatch(ctx, tenantUUID, input)
	require.NoError(t, err)
	require.Equal(t, batch1.ID, batch2.ID)

	batches, err := svc.ListBatches(ctx, tenantUUID, BatchListQuery{BillingCycle: "2026-03-31", Limit: 20})
	require.NoError(t, err)
	require.Len(t, batches, 1)

	deltas, err := svc.ListDeltas(ctx, tenantUUID, DeltaListQuery{BatchID: batch1.ID, Limit: 20})
	require.NoError(t, err)
	require.Len(t, deltas, 3)
}

func setupReconciliationDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_batches (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		billing_cycle TEXT NOT NULL,
		run_type TEXT NOT NULL,
		expected_amount_minor INTEGER NOT NULL DEFAULT 0,
		actual_amount_minor INTEGER NOT NULL DEFAULT 0,
		delta_amount_minor INTEGER NOT NULL DEFAULT 0,
		delta_count INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL,
		started_at DATETIME,
		finished_at DATETIME,
		created_by TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_sr_batches_cycle ON subscription_reconciliation_batches(tenant_uuid, billing_cycle, run_type)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_deltas (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		batch_id TEXT NOT NULL,
		subscription_ref TEXT,
		bill_ref TEXT,
		payment_ref TEXT,
		delta_type TEXT NOT NULL,
		risk_level TEXT NOT NULL,
		expected_amount_minor INTEGER NOT NULL DEFAULT 0,
		actual_amount_minor INTEGER NOT NULL DEFAULT 0,
		delta_amount_minor INTEGER NOT NULL DEFAULT 0,
		reason_code TEXT,
		status TEXT NOT NULL,
		delta_fingerprint TEXT NOT NULL,
		detected_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_sr_deltas_batch ON subscription_reconciliation_deltas(tenant_uuid, batch_id, status)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_delta_tasks (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		delta_id TEXT NOT NULL,
		delta_fingerprint TEXT NOT NULL,
		assignee TEXT,
		priority TEXT,
		sla_level TEXT NOT NULL,
		sla_deadline DATETIME,
		status TEXT NOT NULL,
		resolution TEXT,
		resolution_note TEXT,
		closed_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE INDEX IF NOT EXISTS idx_sr_delta_tasks_delta ON subscription_reconciliation_delta_tasks(tenant_uuid, delta_id, status)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_governance_policies (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		retry_windows JSON,
		escalation_threshold INTEGER NOT NULL DEFAULT 3,
		notify_channels JSON,
		version INTEGER NOT NULL DEFAULT 1,
		effective_from DATETIME,
		effective_to DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS subscription_reconciliation_execution_logs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		subscription_ref TEXT NOT NULL,
		action_type TEXT NOT NULL,
		attempt_no INTEGER NOT NULL DEFAULT 0,
		scheduled_at DATETIME,
		executed_at DATETIME,
		result TEXT NOT NULL,
		failure_reason TEXT,
		operator_type TEXT NOT NULL,
		operator_id TEXT,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	return db
}
