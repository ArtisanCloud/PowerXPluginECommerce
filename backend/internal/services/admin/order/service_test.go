package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestServiceCreateOrder_Idempotency(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-1"
	const spuID = "spu-1"
	const skuID = "sku-1"

	seedCustomer(t, db, tenant, customerID)
	seedSellabilityFixtures(t, db, tenant, spuID, skuID, 10, 0, true)

	svc := NewService(&app.Deps{DB: db})
	req := CreateOrderRequest{
		CustomerID: customerID,
		Channel:    "official",
		ShippingAddress: &ShippingAddress{
			RecipientName:  "张三",
			RecipientPhone: "13800138000",
			Address1:       "北京市朝阳区",
		},
		Items: []CreateOrderItemInput{{SKUID: skuID, Qty: 1}},
		Note:  "manual order",
	}

	first, err := svc.CreateOrder(ctx, tenant, adminID, "idem-1", req)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := svc.CreateOrder(ctx, tenant, adminID, "idem-1", req)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, first.OrderID, second.OrderID)
	require.Equal(t, first.OrderNo, second.OrderNo)

	var orderCount int64
	require.NoError(t, db.Table("orders").Count(&orderCount).Error)
	require.Equal(t, int64(1), orderCount)

	var locked int64
	require.NoError(t, db.Raw(`SELECT locked_qty FROM product_sku_inventories WHERE tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?`, tenant, skuID, "default").Scan(&locked).Error)
	require.Equal(t, int64(1), locked)
}

func TestServiceCreateOrder_OutOfStock_IsAtomic(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-1"
	const spuID = "spu-1"
	const skuID = "sku-1"

	seedCustomer(t, db, tenant, customerID)
	seedSellabilityFixtures(t, db, tenant, spuID, skuID, 0, 0, true)

	svc := NewService(&app.Deps{DB: db})
	req := CreateOrderRequest{
		CustomerID: customerID,
		Channel:    "official",
		ShippingAddress: &ShippingAddress{
			RecipientName:  "张三",
			RecipientPhone: "13800138000",
			Address1:       "北京市朝阳区",
		},
		Items: []CreateOrderItemInput{{SKUID: skuID, Qty: 1}},
	}

	_, err := svc.CreateOrder(ctx, tenant, adminID, "idem-2", req)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrOutOfStock) || errors.Is(err, ErrSellabilityFailed))

	var orderCount int64
	require.NoError(t, db.Table("orders").Count(&orderCount).Error)
	require.Equal(t, int64(0), orderCount)

	var locked int64
	require.NoError(t, db.Raw(`SELECT locked_qty FROM product_sku_inventories WHERE tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?`, tenant, skuID, "default").Scan(&locked).Error)
	require.Equal(t, int64(0), locked)

	var idemCount int64
	require.NoError(t, db.Table("integration_idempotency_records").Count(&idemCount).Error)
	require.Equal(t, int64(0), idemCount)
}

func TestServiceCreateOrder_NotSellable_NoSideEffects(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-1"
	const spuID = "spu-1"
	const skuID = "sku-1"

	seedCustomer(t, db, tenant, customerID)
	seedSellabilityFixtures(t, db, tenant, spuID, skuID, 10, 0, false)

	svc := NewService(&app.Deps{DB: db})
	req := CreateOrderRequest{
		CustomerID: customerID,
		Channel:    "official",
		ShippingAddress: &ShippingAddress{
			RecipientName:  "张三",
			RecipientPhone: "13800138000",
			Address1:       "北京市朝阳区",
		},
		Items: []CreateOrderItemInput{{SKUID: skuID, Qty: 1}},
	}

	_, err := svc.CreateOrder(ctx, tenant, adminID, "idem-3", req)
	require.Error(t, err)

	var orderCount int64
	require.NoError(t, db.Table("orders").Count(&orderCount).Error)
	require.Equal(t, int64(0), orderCount)

	var locked int64
	require.NoError(t, db.Raw(`SELECT locked_qty FROM product_sku_inventories WHERE tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?`, tenant, skuID, "default").Scan(&locked).Error)
	require.Equal(t, int64(0), locked)
}

func TestServiceCreateOrder_CustomerNotFound(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createOrderTables(t, db)

	const tenant = "tenant-test"
	const adminID = "1001"
	const customerID = "cust-missing"
	const spuID = "spu-1"
	const skuID = "sku-1"

	seedSellabilityFixtures(t, db, tenant, spuID, skuID, 10, 0, true)

	svc := NewService(&app.Deps{DB: db})
	req := CreateOrderRequest{
		CustomerID: customerID,
		Channel:    "official",
		Items:      []CreateOrderItemInput{{SKUID: skuID, Qty: 1}},
	}

	_, err := svc.CreateOrder(ctx, tenant, adminID, "idem-4", req)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrCustomerNotFound))

	var idemCount int64
	require.NoError(t, db.Table("integration_idempotency_records").Count(&idemCount).Error)
	require.Equal(t, int64(0), idemCount)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	safeName := strings.NewReplacer("/", "_", " ", "_", ":", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	return db
}

func createOrderTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS customers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_uuid TEXT NOT NULL,
			customer_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_spus (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			status TEXT NOT NULL,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_channels (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			availability TEXT NOT NULL,
			publish_at DATETIME,
			withdraw_at DATETIME,
			audit_state TEXT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_skus (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			status TEXT NOT NULL,
			sku_code TEXT,
			default_values TEXT,
			created_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_sku_inventories (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			sku_id TEXT NOT NULL,
			warehouse_id TEXT NOT NULL,
			available_qty INTEGER NOT NULL DEFAULT 0,
			locked_qty INTEGER NOT NULL DEFAULT 0,
			in_transit_qty INTEGER NOT NULL DEFAULT 0,
			safety_stock INTEGER NOT NULL DEFAULT 0,
			alert_threshold INTEGER NOT NULL DEFAULT 0,
			last_synced_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			UNIQUE(tenant_uuid, sku_id, warehouse_id)
		)`,
		`CREATE TABLE IF NOT EXISTS pricebooks (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			current_version_id TEXT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS pricebook_items (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			pricebook_id TEXT NOT NULL,
			version_id TEXT NOT NULL,
			sku_id TEXT NOT NULL,
			base_amount_minor BIGINT,
			sale_amount_minor BIGINT,
			msrp_amount_minor BIGINT,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS integration_idempotency_records (
			key TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			scope TEXT,
			operation TEXT,
			payload_hash TEXT,
			response_data TEXT,
			metadata TEXT,
			expires_at DATETIME,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_no TEXT NOT NULL,
			customer_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			status TEXT NOT NULL,
			currency TEXT NOT NULL,
			subtotal_amount BIGINT NOT NULL DEFAULT 0,
			total_amount BIGINT NOT NULL DEFAULT 0,
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
			qty BIGINT NOT NULL,
			unit_price BIGINT NOT NULL,
			line_amount BIGINT NOT NULL,
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
		`CREATE TABLE IF NOT EXISTS coupon_assets (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			template_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			coupon_code TEXT NOT NULL,
			status TEXT NOT NULL,
			reserved_order_id TEXT,
			reserved_at DATETIME,
			redeemed_at DATETIME,
			refunded_at DATETIME,
			expired_at DATETIME,
			valid_from DATETIME,
			valid_to DATETIME,
			meta TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_usage_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			order_id TEXT,
			action TEXT NOT NULL,
			action_reason TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			request_id TEXT,
			created_by TEXT,
			created_at DATETIME,
			UNIQUE(tenant_uuid, action, idempotency_key)
		)`,
		`CREATE TABLE IF NOT EXISTS order_coupon_snapshots (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			currency TEXT NOT NULL,
			base_total_minor BIGINT NOT NULL DEFAULT 0,
			discount_total_minor BIGINT NOT NULL DEFAULT 0,
			payable_total_minor BIGINT NOT NULL DEFAULT 0,
			line_allocations TEXT NOT NULL DEFAULT '[]',
			applied_coupons TEXT NOT NULL DEFAULT '[]',
			rejected_coupons TEXT NOT NULL DEFAULT '[]',
			priced_at DATETIME NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			UNIQUE(tenant_uuid, order_id)
		)`,
		`CREATE TABLE IF NOT EXISTS promotion_campaigns (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			promotion_type TEXT NOT NULL,
			condition_rule TEXT NOT NULL,
			scope_rule TEXT NOT NULL,
			action_rule TEXT NOT NULL,
			stacking_rule TEXT NOT NULL,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			status TEXT NOT NULL,
			created_by TEXT,
			updated_by TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_promotion_snapshots (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			currency TEXT NOT NULL,
			base_total_minor BIGINT NOT NULL,
			promotion_discount_minor BIGINT NOT NULL,
			after_promotion_total_minor BIGINT NOT NULL,
			applied_promotions TEXT NOT NULL,
			rejected_promotions TEXT NOT NULL,
			line_allocations TEXT NOT NULL,
			priced_at DATETIME NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			UNIQUE(tenant_uuid, order_id)
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}

func seedCustomer(t *testing.T, db *gorm.DB, tenant, customerID string) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO customers (tenant_uuid, customer_id, name, type, deleted_at) VALUES (?, ?, ?, ?, NULL)`, tenant, customerID, "Demo", "individual").Error)
}

func seedSellabilityFixtures(t *testing.T, db *gorm.DB, tenant, spuID, skuID string, available, locked int64, withPrice bool) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO product_spus (id, tenant_uuid, status, deleted_at) VALUES (?, ?, ?, NULL)`, spuID, tenant, "published").Error)
	require.NoError(t, db.Exec(`INSERT INTO product_spu_channels (id, tenant_uuid, spu_id, channel, availability, audit_state, publish_at, withdraw_at) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL)`,
		"ch-1", tenant, spuID, "official", "published", "approved").Error)
	require.NoError(t, db.Exec(`INSERT INTO product_skus (id, tenant_uuid, spu_id, status, sku_code, default_values, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		skuID, tenant, spuID, "online", "SKU-1", `{}`).Error)
	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-1", tenant, skuID, "default", available, locked).Error)

	if !withPrice {
		return
	}
	require.NoError(t, db.Exec(`INSERT INTO pricebooks (id, tenant_uuid, code, currency, status, current_version_id, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"pb-1", tenant, "base", "CNY", "active", "ver-1").Error)
	require.NoError(t, db.Exec(`INSERT INTO pricebook_items (id, tenant_uuid, pricebook_id, version_id, sku_id, sale_amount_minor, base_amount_minor, msrp_amount_minor, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		"pbi-1", tenant, "pb-1", "ver-1", skuID, 19900, nil, nil).Error)
}
