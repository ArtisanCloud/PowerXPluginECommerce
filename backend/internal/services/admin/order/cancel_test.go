package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceCancelOrder_PendingPayment_Succeeds(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-1"
	const skuID = "sku-1"

	seedCustomer(t, db, tenant, customerID)

	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-1", tenant, skuID, "default", int64(10), int64(2)).Error)
	require.NoError(t, db.Exec(`INSERT INTO orders (id, tenant_uuid, order_no, customer_id, channel, status, currency, subtotal_amount, total_amount, created_by_type, created_by, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		"ord-1", tenant, "O20260101010101ABCDEFGH", customerID, "official", "pending_payment", "CNY", int64(100), int64(100), "admin", adminID, time.Now().UTC(), time.Now().UTC()).Error)
	require.NoError(t, db.Exec(`INSERT INTO order_items (id, tenant_uuid, order_id, sku_id, qty, unit_price, line_amount, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"oi-1", tenant, "ord-1", skuID, int64(2), int64(50), int64(100), time.Now().UTC()).Error)

	svc := NewService(&app.Deps{DB: db})
	resp, err := svc.CancelOrder(ctx, tenant, adminID, "ord-1", "user cancel")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "cancelled", resp.Status)

	var status string
	require.NoError(t, db.Raw(`SELECT status FROM orders WHERE tenant_uuid = ? AND id = ?`, tenant, "ord-1").Scan(&status).Error)
	require.Equal(t, "cancelled", status)

	var locked int64
	require.NoError(t, db.Raw(`SELECT locked_qty FROM product_sku_inventories WHERE tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?`, tenant, skuID, "default").Scan(&locked).Error)
	require.Equal(t, int64(0), locked)

	var eventCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM order_events WHERE tenant_uuid = ? AND order_id = ? AND event_type = ?`, tenant, "ord-1", "order.cancelled").Scan(&eventCount).Error)
	require.Equal(t, int64(1), eventCount)
}

func TestServiceCancelOrder_StatusNotAllowed_NoSideEffects(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-1"
	const skuID = "sku-1"

	seedCustomer(t, db, tenant, customerID)

	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-1", tenant, skuID, "default", int64(10), int64(2)).Error)
	require.NoError(t, db.Exec(`INSERT INTO orders (id, tenant_uuid, order_no, customer_id, channel, status, currency, subtotal_amount, total_amount, created_by_type, created_by, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		"ord-1", tenant, "O20260101010101ABCDEFGH", customerID, "official", "paid", "CNY", int64(100), int64(100), "admin", adminID, time.Now().UTC(), time.Now().UTC()).Error)
	require.NoError(t, db.Exec(`INSERT INTO order_items (id, tenant_uuid, order_id, sku_id, qty, unit_price, line_amount, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"oi-1", tenant, "ord-1", skuID, int64(2), int64(50), int64(100), time.Now().UTC()).Error)

	svc := NewService(&app.Deps{DB: db})
	_, err := svc.CancelOrder(ctx, tenant, adminID, "ord-1", "should fail")
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrOrderNotCancellable))

	var status string
	require.NoError(t, db.Raw(`SELECT status FROM orders WHERE tenant_uuid = ? AND id = ?`, tenant, "ord-1").Scan(&status).Error)
	require.Equal(t, "paid", status)

	var locked int64
	require.NoError(t, db.Raw(`SELECT locked_qty FROM product_sku_inventories WHERE tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?`, tenant, skuID, "default").Scan(&locked).Error)
	require.Equal(t, int64(2), locked)

	var eventCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(1) FROM order_events WHERE tenant_uuid = ? AND order_id = ? AND event_type = ?`, tenant, "ord-1", "order.cancelled").Scan(&eventCount).Error)
	require.Equal(t, int64(0), eventCount)
}

func TestServiceCancelOrder_UnlockFloorZero(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-1"
	const skuID = "sku-1"

	seedCustomer(t, db, tenant, customerID)

	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-1", tenant, skuID, "default", int64(10), int64(1)).Error)
	require.NoError(t, db.Exec(`INSERT INTO orders (id, tenant_uuid, order_no, customer_id, channel, status, currency, subtotal_amount, total_amount, created_by_type, created_by, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		"ord-1", tenant, "O20260101010101ABCDEFGH", customerID, "official", "pending_payment", "CNY", int64(100), int64(100), "admin", adminID, time.Now().UTC(), time.Now().UTC()).Error)
	require.NoError(t, db.Exec(`INSERT INTO order_items (id, tenant_uuid, order_id, sku_id, qty, unit_price, line_amount, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"oi-1", tenant, "ord-1", skuID, int64(5), int64(50), int64(250), time.Now().UTC()).Error)

	svc := NewService(&app.Deps{DB: db})
	_, err := svc.CancelOrder(ctx, tenant, adminID, "ord-1", "")
	require.NoError(t, err)

	var locked int64
	require.NoError(t, db.Raw(`SELECT locked_qty FROM product_sku_inventories WHERE tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?`, tenant, skuID, "default").Scan(&locked).Error)
	require.Equal(t, int64(0), locked)
}
