package product_sku

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProductSkuServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:product_sku_service_tests?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS product_skus (
        id TEXT PRIMARY KEY,
        tenant_uuid TEXT NOT NULL,
        created_at DATETIME,
        updated_at DATETIME,
        deleted_at DATETIME
    );`).Error)

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

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS product_sku_audit_logs (
        id TEXT PRIMARY KEY,
        tenant_uuid TEXT NOT NULL,
        sku_id TEXT NOT NULL,
        warehouse_id TEXT NOT NULL,
        action TEXT NOT NULL,
        actor TEXT,
        delta INTEGER NOT NULL,
        before_available_qty INTEGER NOT NULL,
        after_available_qty INTEGER NOT NULL,
        request_id TEXT,
        created_at DATETIME,
        deleted_at DATETIME
    );`).Error)

	return db
}

func TestService_AdjustInventory_DeltaAndAudit(t *testing.T) {
	db := setupProductSkuServiceTestDB(t)

	deps := &app.Deps{DB: db}
	service := NewService(deps)
	require.True(t, service.Ready())

	tenantID := "tenant-1"
	skuID := "sku-1"
	require.NoError(t, db.Select("id", "tenant_uuid").Create(&productskumodel.ProductSKU{ID: skuID, TenantUUID: tenantID}).Error)

	ctx := context.Background()
	ctx = authx.ContextWithTenantUUID(ctx, tenantID)
	ctx = authx.ContextWithTenantContext(ctx, authx.TenantContext{TenantUUID: tenantID, UserID: 123})

	snap, err := service.AdjustInventory(ctx, skuID, 10)
	require.NoError(t, err)
	require.NotNil(t, snap)
	require.Equal(t, int64(10), snap.Total.AvailableQty)

	_, err = service.AdjustInventory(ctx, skuID, -20)
	require.Error(t, err)

	var auditCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&productskumodel.ProductSKUAuditLog{}).
		Where("tenant_uuid = ? AND sku_id = ?", tenantID, skuID).
		Count(&auditCount).Error)
	require.Equal(t, int64(1), auditCount)

	require.NoError(t, repo.NewInventoryRepository(db).WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		row, err := repo.NewInventoryRepository(db).EnsureWarehouseRowForUpdate(ctx, tx, tenantID, skuID, "default")
		require.NoError(t, err)
		require.Equal(t, int64(10), row.AvailableQty)
		return nil
	}))
}
