package product_sku

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProductSkuInventoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:product_sku_inventory_tests?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS product_sku_inventories (
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
    );`).Error)

	return db
}

func TestInventoryRepository_EnsureWarehouseRowForUpdate_IsIdempotent(t *testing.T) {
	db := setupProductSkuInventoryTestDB(t)
	repo := NewInventoryRepository(db)
	ctx := context.Background()
	tenantID := "tenant-1"
	skuID := "sku-1"
	warehouseID := "default"

	require.NoError(t, repo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		row, err := repo.EnsureWarehouseRowForUpdate(ctx, tx, tenantID, skuID, warehouseID)
		require.NoError(t, err)
		require.NotEmpty(t, row.ID)
		require.Equal(t, int64(0), row.AvailableQty)
		return nil
	}))

	require.NoError(t, repo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		row, err := repo.EnsureWarehouseRowForUpdate(ctx, tx, tenantID, skuID, warehouseID)
		require.NoError(t, err)
		require.NotEmpty(t, row.ID)
		return nil
	}))

	var count int64
	require.NoError(t, db.WithContext(ctx).
		Model(&productskumodel.ProductSKUInventory{}).
		Where("tenant_uuid = ? AND sku_id = ? AND warehouse_id = ?", tenantID, skuID, warehouseID).
		Count(&count).Error)
	require.Equal(t, int64(1), count)
}

