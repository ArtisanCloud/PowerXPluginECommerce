package after_sales_test

import (
	"context"
	"testing"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	orderModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	reverseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/reverse"
	adminAfterSales "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	miniappAfterSales "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/miniapp/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAfterSalesUS123Smoke(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:after_sales_us123_smoke?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	models.ForceSchemaForTests("")
	require.NoError(t, createSmokeTables(db))

	deps := &app.Deps{DB: db, Ctx: context.Background()}
	tenantUUID := "11111111-1111-1111-1111-111111111111"
	customerID := "22222222-2222-2222-2222-222222222222"
	orderID := "33333333-3333-3333-3333-333333333333"
	orderItemID := "44444444-4444-4444-4444-444444444444"

	require.NoError(t, db.Create(&orderModel.Order{
		ID:         orderID,
		TenantUUID: tenantUUID,
		OrderNo:    "ORD-SMOKE-001",
		CustomerID: customerID,
		Channel:    "miniapp",
		Status:     "paid",
		Currency:   "CNY",
		CreatedBy:  customerID,
	}).Error)
	require.NoError(t, db.Create(&orderModel.OrderItem{
		ID:         orderItemID,
		TenantUUID: tenantUUID,
		OrderID:    orderID,
		SKUID:      "55555555-5555-5555-5555-555555555555",
		Qty:        1,
		UnitPrice:  1000,
		LineAmount: 1000,
	}).Error)

	miniSvc := miniappAfterSales.NewCaseService(deps)
	adminCaseSvc := adminAfterSales.NewCaseService(deps)
	decisionSvc := adminAfterSales.NewDecisionService(deps)
	reverseSvc := adminAfterSales.NewReverseLinkService(deps)

	us1Start := time.Now()
	created, err := miniSvc.Create(context.Background(), tenantUUID, customerID, miniappAfterSales.CreateCaseRequest{
		OrderID:              orderID,
		OrderItemID:          orderItemID,
		CaseType:             "return_refund",
		ReasonCode:           "damaged",
		ReasonDetail:         "box damaged",
		RequestedQty:         1,
		RequestedAmountMinor: 1000,
		Currency:             "CNY",
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, adminAfterSales.CaseStatusPending, created.Case.Status)
	us1Duration := time.Since(us1Start)

	us2Start := time.Now()
	_, err = adminCaseSvc.Transition(context.Background(), tenantUUID, created.Case.ID, "accept", "9001", "")
	require.NoError(t, err)
	_, err = adminCaseSvc.Transition(context.Background(), tenantUUID, created.Case.ID, "review", "9001", "")
	require.NoError(t, err)
	approved, err := decisionSvc.Approve(context.Background(), tenantUUID, created.Case.ID, "9001", "approved by smoke")
	require.NoError(t, err)
	require.Equal(t, adminAfterSales.CaseStatusApproved, approved.Status)
	us2Duration := time.Since(us2Start)

	waybillID := utils.NewUUID()
	require.NoError(t, db.Create(&reverseModel.Waybill{
		ID:          waybillID,
		TenantUUID:  tenantUUID,
		OrderID:     orderID,
		AfterSaleID: created.Case.ID,
		WaybillNo:   "RWB-SMOKE-001",
		Status:      "created",
	}).Error)

	us3Start := time.Now()
	link, err := reverseSvc.Link(context.Background(), tenantUUID, created.Case.ID, "9001", adminAfterSales.ReverseLogisticsLinkInput{
		ReverseWaybillID: waybillID,
		CarrierCode:      "SF",
		Note:             "smoke link",
	})
	require.NoError(t, err)
	require.Equal(t, "RWB-SMOKE-001", link.ReverseWaybillNo)
	require.Equal(t, "pending", link.ReceiveStatus)

	list, err := adminCaseSvc.List(context.Background(), tenantUUID, adminAfterSales.CaseQuery{OrderID: orderID, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.NotEmpty(t, list.Items)
	require.Equal(t, "RWB-SMOKE-001", list.Items[0].ReverseWaybillNo)

	var afterSalesEventCount int64
	require.NoError(t, db.Model(&orderModel.OrderEvent{}).
		Where("tenant_uuid = ? AND order_id = ? AND event_type = ?", tenantUUID, orderID, "after_sales.approved").
		Count(&afterSalesEventCount).Error)
	require.GreaterOrEqual(t, afterSalesEventCount, int64(1))
	us3Duration := time.Since(us3Start)

	t.Logf("US1 duration=%s", us1Duration)
	t.Logf("US2 duration=%s", us2Duration)
	t.Logf("US3 duration=%s", us3Duration)
}

func createSmokeTables(db *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_no TEXT,
			customer_id TEXT NOT NULL,
			channel TEXT,
			status TEXT NOT NULL,
			currency TEXT,
			subtotal_amount INTEGER DEFAULT 0,
			total_amount INTEGER DEFAULT 0,
			shipping_address_id TEXT,
			shipping_address_snapshot TEXT,
			price_snapshot TEXT,
			sellability_snapshot TEXT,
			created_by_type TEXT,
			created_by TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			sku_id TEXT NOT NULL,
			qty INTEGER NOT NULL,
			unit_price INTEGER NOT NULL,
			line_amount INTEGER NOT NULL,
			price_source TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_events (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			operator_type TEXT NOT NULL,
			operator TEXT,
			payload TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payment_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT,
			order_no TEXT,
			amount_currency TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS payment_refunds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			transaction_id INTEGER NOT NULL,
			status TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS after_sales_cases (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			case_no TEXT NOT NULL,
			order_id TEXT NOT NULL,
			order_item_id TEXT NOT NULL,
			customer_id TEXT NOT NULL,
			case_type TEXT NOT NULL,
			status TEXT NOT NULL,
			reason_code TEXT,
			reason_detail TEXT,
			requested_qty INTEGER NOT NULL,
			requested_amount_minor INTEGER NOT NULL,
			currency TEXT NOT NULL,
			source_channel TEXT,
			closed_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS after_sales_timelines (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			case_id TEXT NOT NULL,
			action TEXT NOT NULL,
			from_status TEXT,
			to_status TEXT NOT NULL,
			operator_type TEXT NOT NULL,
			operator_id TEXT,
			note TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS after_sales_evidences (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			case_id TEXT NOT NULL,
			uploader_type TEXT NOT NULL,
			uploader_id TEXT,
			evidence_type TEXT NOT NULL,
			content_ref TEXT NOT NULL,
			description TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS after_sales_decisions (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			case_id TEXT NOT NULL,
			decision TEXT NOT NULL,
			reject_reason_code TEXT,
			decision_note TEXT,
			decided_by TEXT,
			decided_at DATETIME,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS after_sales_reverse_logistics_links (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			case_id TEXT NOT NULL,
			reverse_waybill_id TEXT,
			reverse_waybill_no TEXT,
			carrier_code TEXT,
			receive_status TEXT NOT NULL,
			received_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS reverse_waybills (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			after_sale_id TEXT NOT NULL,
			waybill_no TEXT NOT NULL,
			status TEXT NOT NULL,
			inspection_result TEXT,
			disposition TEXT,
			metadata TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}
