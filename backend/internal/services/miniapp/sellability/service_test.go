package sellability

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

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

func TestServiceEvaluate(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createSellabilityTables(t, db)

	const tenant = "tenant-test"
	const spuID = "spu-1"

	require.NoError(t, db.Exec(`INSERT INTO product_spus (id, tenant_uuid, status, deleted_at) VALUES (?, ?, ?, NULL)`, spuID, tenant, "published").Error)
	require.NoError(t, db.Exec(`INSERT INTO product_skus (id, tenant_uuid, spu_id, status, sku_code, default_values, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"sku-1", tenant, spuID, "online", "SKU-1", `{}`).Error)
	require.NoError(t, db.Exec(`INSERT INTO product_skus (id, tenant_uuid, spu_id, status, sku_code, default_values, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"sku-2", tenant, spuID, "online", "SKU-2", `{}`).Error)

	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-1", tenant, "sku-1", "default", 10, 2).Error)
	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-2", tenant, "sku-2", "default", 0, 0).Error)

	require.NoError(t, db.Exec(`INSERT INTO pricebooks (id, tenant_uuid, code, currency, status, current_version_id, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"pb-1", tenant, "base", "CNY", "active", "ver-1").Error)
	require.NoError(t, db.Exec(`INSERT INTO pricebook_items (id, tenant_uuid, pricebook_id, version_id, sku_id, sale_amount_minor, base_amount_minor, msrp_amount_minor, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		"pbi-1", tenant, "pb-1", "ver-1", "sku-1", 19900, nil, nil).Error)

	t.Run("sellable and unsellable mix", func(t *testing.T) {
		require.NoError(t, db.Exec(`INSERT INTO product_spu_channels (id, tenant_uuid, spu_id, channel, availability, audit_state, publish_at, withdraw_at) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL)`,
			"ch-1", tenant, spuID, "official", "published", "approved").Error)

		svc := NewService(db).WithNow(func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
		got, err := svc.Evaluate(ctx, tenant, spuID, "official", "zh-CN")
		require.NoError(t, err)
		require.Equal(t, spuID, got.SPUID)
		require.Equal(t, "official", got.Channel)
		require.Len(t, got.Items, 2)

		var sku1, sku2 ItemDTO
		for _, it := range got.Items {
			if it.SKUID == "sku-1" {
				sku1 = it
			}
			if it.SKUID == "sku-2" {
				sku2 = it
			}
		}

		require.True(t, sku1.Sellable)
		require.Empty(t, sku1.Reasons)
		require.NotNil(t, sku1.Price)
		require.Equal(t, 199.0, sku1.Price.Amount)
		require.Equal(t, "CNY", sku1.Price.Currency)
		require.Equal(t, 8, sku1.AvailableQty)

		require.False(t, sku2.Sellable)
		require.Contains(t, sku2.Reasons, string(ReasonNoPublicPrice))
		require.Contains(t, sku2.Reasons, string(ReasonOutOfStock))
		require.Nil(t, sku2.Price)
		require.Equal(t, 0, sku2.AvailableQty)
	})

	t.Run("channel disabled", func(t *testing.T) {
		require.NoError(t, db.Exec(`DELETE FROM product_spu_channels`).Error)
		require.NoError(t, db.Exec(`INSERT INTO product_spu_channels (id, tenant_uuid, spu_id, channel, availability, audit_state, publish_at, withdraw_at) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL)`,
			"ch-2", tenant, spuID, "official", "unlisted", "approved").Error)

		svc := NewService(db).WithNow(func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
		got, err := svc.Evaluate(ctx, tenant, spuID, "official", "")
		require.NoError(t, err)
		for _, it := range got.Items {
			require.Contains(t, it.Reasons, string(ReasonChannelDisabled))
			require.False(t, it.Sellable)
		}
	})

	t.Run("not in availability window", func(t *testing.T) {
		require.NoError(t, db.Exec(`DELETE FROM product_spu_channels`).Error)
		publishAt := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, db.Exec(`INSERT INTO product_spu_channels (id, tenant_uuid, spu_id, channel, availability, audit_state, publish_at, withdraw_at) VALUES (?, ?, ?, ?, ?, ?, ?, NULL)`,
			"ch-3", tenant, spuID, "official", "scheduled", "approved", publishAt).Error)

		svc := NewService(db).WithNow(func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
		got, err := svc.Evaluate(ctx, tenant, spuID, "official", "")
		require.NoError(t, err)
		for _, it := range got.Items {
			require.Contains(t, it.Reasons, string(ReasonNotInAvailabilityWindow))
			require.False(t, it.Sellable)
		}
	})

	t.Run("channel audit blocked", func(t *testing.T) {
		require.NoError(t, db.Exec(`DELETE FROM product_spu_channels`).Error)
		require.NoError(t, db.Exec(`INSERT INTO product_spu_channels (id, tenant_uuid, spu_id, channel, availability, audit_state, publish_at, withdraw_at) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL)`,
			"ch-4", tenant, spuID, "official", "published", "pending").Error)

		svc := NewService(db).WithNow(func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
		got, err := svc.Evaluate(ctx, tenant, spuID, "official", "")
		require.NoError(t, err)
		for _, it := range got.Items {
			require.Contains(t, it.Reasons, string(ReasonChannelStatusBlocked))
			require.False(t, it.Sellable)
		}
	})
}

func TestServiceEvaluateSKUs(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")

	db := newTestDB(t)
	createSellabilityTables(t, db)

	const tenant = "tenant-test"
	const spuID = "spu-1"

	require.NoError(t, db.Exec(`INSERT INTO product_spus (id, tenant_uuid, status, deleted_at) VALUES (?, ?, ?, NULL)`, spuID, tenant, "published").Error)
	require.NoError(t, db.Exec(`INSERT INTO product_spu_channels (id, tenant_uuid, spu_id, channel, availability, audit_state, publish_at, withdraw_at) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL)`,
		"ch-1", tenant, spuID, "official", "published", "approved").Error)

	require.NoError(t, db.Exec(`INSERT INTO product_skus (id, tenant_uuid, spu_id, status, sku_code, default_values, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"sku-1", tenant, spuID, "online", "SKU-1", `{}`).Error)
	require.NoError(t, db.Exec(`INSERT INTO product_sku_inventories (id, tenant_uuid, sku_id, warehouse_id, available_qty, locked_qty, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"inv-1", tenant, "sku-1", "default", 10, 2).Error)

	require.NoError(t, db.Exec(`INSERT INTO pricebooks (id, tenant_uuid, code, currency, status, current_version_id, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NULL)`,
		"pb-1", tenant, "base", "CNY", "active", "ver-1").Error)
	require.NoError(t, db.Exec(`INSERT INTO pricebook_items (id, tenant_uuid, pricebook_id, version_id, sku_id, sale_amount_minor, base_amount_minor, msrp_amount_minor, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL)`,
		"pbi-1", tenant, "pb-1", "ver-1", "sku-1", 19900, nil, nil).Error)

	svc := NewService(db)
	got, err := svc.EvaluateSKUs(ctx, tenant, []string{"sku-1"}, "official", "zh-CN")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "official", got.Channel)
	require.Len(t, got.Items, 1)
	require.Equal(t, "sku-1", got.Items[0].SKUID)
	require.True(t, got.Items[0].Sellable)
	require.NotNil(t, got.Items[0].Price)
	require.Equal(t, 199.0, got.Items[0].Price.Amount)
	require.Equal(t, "CNY", got.Items[0].Price.Currency)
	require.Equal(t, 8, got.Items[0].AvailableQty)
}

func createSellabilityTables(t *testing.T, db *gorm.DB) {
	stmts := []string{
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
			deleted_at DATETIME
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
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
