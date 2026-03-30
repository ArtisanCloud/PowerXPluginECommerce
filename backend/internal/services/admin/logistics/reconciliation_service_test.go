package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestReconciliationService_MatchingAndCaseFlow(t *testing.T) {
	db := setupBillingDB(t, "logistics_reconciliation_flow")
	ensureReconciliationTables(t, db)
	svc := NewReconciliationService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-rcn")

	batch, err := svc.CreateBatch(ctx, "tenant-rcn", CreateReconciliationBatchRequest{
		CarrierID: "carrier-rcn",
		Records: []ReconciliationInputRecord{
			{WaybillNo: "WB-RCN-1", CarrierID: "carrier-rcn", BillAmount: 100, BankAmount: 100, InvoiceAmount: 100},
			{WaybillNo: "WB-RCN-2", CarrierID: "carrier-rcn", BillAmount: 80, BankAmount: 80, InvoiceAmount: 0},
			{WaybillNo: "WB-RCN-3", CarrierID: "carrier-rcn", BillAmount: 60, BankAmount: 0, InvoiceAmount: 60},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 3, batch.RecordCount)
	require.Equal(t, 1, batch.MatchedCount)
	require.Equal(t, 2, batch.ExceptionCount)

	records, err := svc.ListRecords(ctx, "tenant-rcn", batch.ID, "", 20)
	require.NoError(t, err)
	require.Len(t, records, 3)

	cases, err := svc.ListCases(ctx, "tenant-rcn", batch.ID, "", 20)
	require.NoError(t, err)
	require.Len(t, cases, 2)

	handled, err := svc.HandleCase(ctx, "tenant-rcn", cases[0].ID, HandleReconciliationCaseRequest{
		Action:     "confirm",
		OperatorID: "admin",
		Note:       "auto-fix",
	})
	require.NoError(t, err)
	require.Equal(t, "confirmed", handled.Status)

	reloaded, err := svc.ListRecords(ctx, "tenant-rcn", batch.ID, "", 20)
	require.NoError(t, err)
	resolvedCount := 0
	for _, row := range reloaded {
		if row.Status == "resolved" {
			resolvedCount++
		}
	}
	require.GreaterOrEqual(t, resolvedCount, 2)
}

func TestReconciliationService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_reconciliation_tenant")
	ensureReconciliationTables(t, db)
	svc := NewReconciliationService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-ra")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-rb")

	batch, err := svc.CreateBatch(ctxA, "tenant-ra", CreateReconciliationBatchRequest{
		CarrierID: "carrier-ra",
		Records: []ReconciliationInputRecord{
			{WaybillNo: "WB-A", CarrierID: "carrier-ra", BillAmount: 50, BankAmount: 48, InvoiceAmount: 50},
		},
	})
	require.NoError(t, err)

	rowsB, err := svc.ListBatches(ctxB, "tenant-rb", "", "", 20)
	require.NoError(t, err)
	require.Len(t, rowsB, 0)

	casesB, err := svc.ListCases(ctxB, "tenant-rb", batch.ID, "", 20)
	require.NoError(t, err)
	require.Len(t, casesB, 0)
}

func ensureReconciliationTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_reconciliation_batches (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		batch_no TEXT NOT NULL,
		carrier_id TEXT,
		status TEXT NOT NULL DEFAULT 'open',
		record_count INTEGER NOT NULL DEFAULT 0,
		matched_count INTEGER NOT NULL DEFAULT 0,
		exception_count INTEGER NOT NULL DEFAULT 0,
		total_bill_amount NUMERIC NOT NULL DEFAULT 0,
		total_bank_amount NUMERIC NOT NULL DEFAULT 0,
		total_invoice_amount NUMERIC NOT NULL DEFAULT 0,
		summary JSON,
		metadata JSON,
		executed_at DATETIME,
		confirmed_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_reconciliation_batch_no ON logistics_reconciliation_batches(tenant_uuid, batch_no)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_reconciliation_records (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		batch_id TEXT NOT NULL,
		waybill_id TEXT,
		waybill_no TEXT,
		carrier_id TEXT,
		bill_amount NUMERIC NOT NULL DEFAULT 0,
		bank_amount NUMERIC NOT NULL DEFAULT 0,
		invoice_amount NUMERIC NOT NULL DEFAULT 0,
		diff_amount NUMERIC NOT NULL DEFAULT 0,
		match_type TEXT NOT NULL DEFAULT 'matched',
		suggestion TEXT NOT NULL DEFAULT 'auto_archive',
		status TEXT NOT NULL DEFAULT 'resolved',
		case_id TEXT,
		handled_action TEXT,
		handled_note TEXT,
		handled_by TEXT,
		handled_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_reconciliation_record ON logistics_reconciliation_records(tenant_uuid, batch_id, waybill_no)`).Error)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_reconciliation_cases (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		case_no TEXT NOT NULL,
		batch_id TEXT NOT NULL,
		record_id TEXT NOT NULL,
		carrier_id TEXT,
		waybill_no TEXT,
		status TEXT NOT NULL DEFAULT 'open',
		reason TEXT NOT NULL DEFAULT 'amount_mismatch',
		suggestion TEXT NOT NULL DEFAULT 'manual_review',
		action_note TEXT,
		handled_by TEXT,
		handled_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_reconciliation_case_no ON logistics_reconciliation_cases(tenant_uuid, case_no)`).Error)
}
