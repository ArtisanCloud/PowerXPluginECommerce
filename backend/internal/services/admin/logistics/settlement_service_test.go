package logistics

import (
	"context"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSettlementService_AttributionAndTransition(t *testing.T) {
	db := setupBillingDB(t, "logistics_settlement_transition")
	ensureSettlementTables(t, db)
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-settlement")
	seedCarrier(t, db, "tenant-settlement", "carrier-s1", "Carrier-S1")
	waybillSvc := NewWaybillService(&app.Deps{DB: db})
	billingSvc := NewBillingService(&app.Deps{DB: db})
	svc := NewSettlementService(&app.Deps{DB: db})

	wb, _, err := waybillSvc.Create(ctx, "tenant-settlement", CreateWaybillRequest{
		OrderID:        "order-set-1",
		CarrierID:      "carrier-s1",
		ServiceCode:    "std",
		WaybillNo:      "WB-SET-1",
		PackageKey:     "order-set-1#1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb.FeeAmount = 10
	require.NoError(t, db.WithContext(ctx).Save(wb).Error)
	_, err = billingSvc.UpdateWaybillCost(ctx, "tenant-settlement", wb.ID, UpdateWaybillCostRequest{ActualFeeAmount: 14})
	require.NoError(t, err)

	batch, err := svc.CreateBatch(ctx, "tenant-settlement", CreateSettlementBatchRequest{CarrierID: "carrier-s1"})
	require.NoError(t, err)
	require.Equal(t, "open", batch.Status)

	diffs, err := svc.ListDiffs(ctx, "tenant-settlement", batch.ID, "", 20)
	require.NoError(t, err)
	require.NotEmpty(t, diffs)
	require.Equal(t, "weight_mismatch", diffs[0].Attribution)

	_, err = svc.ConfirmBatch(ctx, "tenant-settlement", batch.ID)
	require.Error(t, err)

	_, err = svc.HandleDiff(ctx, "tenant-settlement", diffs[0].ID, HandleSettlementDiffRequest{Action: "dispute", OperatorID: "admin"})
	require.NoError(t, err)
	confirmed, err := svc.ConfirmBatch(ctx, "tenant-settlement", batch.ID)
	require.NoError(t, err)
	require.Equal(t, "confirmed", confirmed.Status)
}

func TestSettlementService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_settlement_tenant")
	ensureSettlementTables(t, db)
	svc := NewSettlementService(&app.Deps{DB: db})

	ctxA := authx.ContextWithTenantUUID(context.Background(), "tenant-a")
	ctxB := authx.ContextWithTenantUUID(context.Background(), "tenant-b")
	seedCarrier(t, db, "tenant-a", "carrier-a", "Carrier-A")

	batch, err := svc.CreateBatch(ctxA, "tenant-a", CreateSettlementBatchRequest{CarrierID: "carrier-a"})
	require.NoError(t, err)
	rowsB, err := svc.ListBatches(ctxB, "tenant-b", "", "", 20)
	require.NoError(t, err)
	require.Len(t, rowsB, 0)

	_, err = svc.ConfirmBatch(ctxB, "tenant-b", batch.ID)
	require.Error(t, err)
}

func ensureSettlementTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_settlement_batches (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		batch_no TEXT NOT NULL,
		carrier_id TEXT,
		status TEXT NOT NULL DEFAULT 'open',
		waybill_count INTEGER NOT NULL DEFAULT 0,
		diff_count INTEGER NOT NULL DEFAULT 0,
		total_expected_fee NUMERIC NOT NULL DEFAULT 0,
		total_actual_fee NUMERIC NOT NULL DEFAULT 0,
		total_diff_amount NUMERIC NOT NULL DEFAULT 0,
		suggestion_summary JSON,
		metadata JSON,
		confirmed_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_settlement_batch_no ON logistics_settlement_batches(tenant_uuid, batch_no)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_settlement_diffs (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		batch_id TEXT NOT NULL,
		waybill_id TEXT NOT NULL,
		waybill_no TEXT,
		carrier_id TEXT,
		expected_fee NUMERIC NOT NULL DEFAULT 0,
		actual_fee NUMERIC NOT NULL DEFAULT 0,
		diff_amount NUMERIC NOT NULL DEFAULT 0,
		attribution TEXT NOT NULL DEFAULT 'matched',
		suggestion TEXT NOT NULL DEFAULT 'accept',
		status TEXT NOT NULL DEFAULT 'pending',
		handled_action TEXT,
		handled_note TEXT,
		handled_by TEXT,
		handled_at DATETIME,
		metadata JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_settlement_diff ON logistics_settlement_diffs(tenant_uuid, batch_id, waybill_id)`).Error)
}
