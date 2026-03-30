package logistics

import (
	"context"
	"strings"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBillingCaseService_TransitionFlow(t *testing.T) {
	db := setupBillingDB(t, "logistics_billing_case_transition")
	ctx := context.Background()

	billingSvc := NewBillingService(&app.Deps{DB: db})
	caseSvc := NewBillingCaseService(&app.Deps{DB: db})
	waybillSvc := NewWaybillService(&app.Deps{DB: db})
	tenant := authx.ContextWithTenantUUID(ctx, "tenant-case-1")

	seedCarrier(t, db, "tenant-case-1", "carrier-case-1", "CaseCarrier")
	wb, _, err := waybillSvc.Create(tenant, "tenant-case-1", CreateWaybillRequest{
		OrderID:        "order-case-1",
		CarrierID:      "carrier-case-1",
		ServiceCode:    "std",
		WaybillNo:      "WB-CASE-1",
		PackageKey:     "order-case-1-pkg-1",
		ShipmentItems:  []string{"sku-1"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wb.FeeAmount = 10
	require.NoError(t, db.WithContext(tenant).Save(wb).Error)
	_, err = billingSvc.UpdateWaybillCost(tenant, "tenant-case-1", wb.ID, UpdateWaybillCostRequest{ActualFeeAmount: 13})
	require.NoError(t, err)

	caseRow, idem, err := caseSvc.Create(tenant, "tenant-case-1", CreateBillingCaseRequest{
		WaybillID: wb.ID,
		Reason:    "carrier bill mismatch",
	})
	require.NoError(t, err)
	require.Equal(t, "created", idem)
	require.Equal(t, "open", caseRow.Status)

	caseRow2, idem2, err := caseSvc.Create(tenant, "tenant-case-1", CreateBillingCaseRequest{
		WaybillID: wb.ID,
		Reason:    "duplicate open",
	})
	require.NoError(t, err)
	require.Equal(t, "replayed", idem2)
	require.Equal(t, caseRow.ID, caseRow2.ID)

	confirmed, err := caseSvc.Transition(tenant, "tenant-case-1", caseRow.ID, TransitionBillingCaseRequest{
		Action: "confirm",
		Note:   "confirm diff",
	})
	require.NoError(t, err)
	require.Equal(t, "confirmed", confirmed.Status)

	appealed, err := caseSvc.Transition(tenant, "tenant-case-1", caseRow.ID, TransitionBillingCaseRequest{
		Action: "appeal",
		Note:   "appeal by carrier",
	})
	require.NoError(t, err)
	require.Equal(t, "appealed", appealed.Status)

	closed, err := caseSvc.Transition(tenant, "tenant-case-1", caseRow.ID, TransitionBillingCaseRequest{
		Action: "writeoff",
		Note:   "writeoff settled",
	})
	require.NoError(t, err)
	require.Equal(t, "written_off", closed.Status)
	require.NotNil(t, closed.ClosedAt)

	_, err = caseSvc.Transition(tenant, "tenant-case-1", caseRow.ID, TransitionBillingCaseRequest{
		Action: "appeal",
	})
	require.Error(t, err)
}

func TestBillingCaseService_TenantIsolation(t *testing.T) {
	db := setupBillingDB(t, "logistics_billing_case_tenant")
	ctx := context.Background()
	caseSvc := NewBillingCaseService(&app.Deps{DB: db})
	billingSvc := NewBillingService(&app.Deps{DB: db})
	waybillSvc := NewWaybillService(&app.Deps{DB: db})

	seedCarrier(t, db, "tenant-case-a", "carrier-case-a", "Carrier-A")
	seedCarrier(t, db, "tenant-case-b", "carrier-case-b", "Carrier-B")
	ctxA := authx.ContextWithTenantUUID(ctx, "tenant-case-a")
	ctxB := authx.ContextWithTenantUUID(ctx, "tenant-case-b")
	wbA, _, err := waybillSvc.Create(ctxA, "tenant-case-a", CreateWaybillRequest{
		OrderID:        "order-a",
		CarrierID:      "carrier-case-a",
		ServiceCode:    "std",
		WaybillNo:      "WB-A",
		PackageKey:     "order-a-pkg-1",
		ShipmentItems:  []string{"sku-a"},
		OrderItemCount: 1,
	})
	require.NoError(t, err)
	wbA.FeeAmount = 20
	require.NoError(t, db.WithContext(ctxA).Save(wbA).Error)
	_, err = billingSvc.UpdateWaybillCost(ctxA, "tenant-case-a", wbA.ID, UpdateWaybillCostRequest{ActualFeeAmount: 25})
	require.NoError(t, err)
	caseA, _, err := caseSvc.Create(ctxA, "tenant-case-a", CreateBillingCaseRequest{WaybillID: wbA.ID})
	require.NoError(t, err)

	rowsA, err := caseSvc.List(ctxA, "tenant-case-a", BillingCaseQuery{})
	require.NoError(t, err)
	require.Len(t, rowsA, 1)
	require.Equal(t, caseA.ID, rowsA[0].ID)

	rowsB, err := caseSvc.List(ctxB, "tenant-case-b", BillingCaseQuery{})
	require.NoError(t, err)
	require.Len(t, rowsB, 0)

	_, err = caseSvc.Transition(ctxB, "tenant-case-b", caseA.ID, TransitionBillingCaseRequest{Action: "confirm"})
	require.Error(t, err)
}

func seedCarrier(t *testing.T, db *gorm.DB, tenantUUID, carrierID, name string) {
	t.Helper()
	require.NoError(t, db.WithContext(authx.ContextWithTenantUUID(context.Background(), tenantUUID)).
		Exec(
			`INSERT INTO logistics_carriers (id, tenant_uuid, name, code, type, status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, 'self', 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			carrierID,
			tenantUUID,
			name,
			strings.ToLower(strings.ReplaceAll(name, " ", "_")),
		).Error)
}
