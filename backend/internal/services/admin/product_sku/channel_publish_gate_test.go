package product_sku

import (
	"context"
	"testing"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestService_PublishChannelMapping_RequiresInventory(t *testing.T) {
	db := setupProductSkuServiceTestDB(t)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS product_sku_channels (
        id TEXT PRIMARY KEY,
        tenant_uuid TEXT NOT NULL,
        sku_id TEXT NOT NULL,
        channel_code TEXT NOT NULL,
        channel_sku_id TEXT NOT NULL,
        status TEXT NOT NULL,
        sync_mode TEXT,
        publish_time DATETIME,
        publish_task_id TEXT,
        last_error TEXT,
        price_override TEXT,
        media_override TEXT,
        metadata TEXT,
        created_at DATETIME,
        updated_at DATETIME
    );`).Error)

	deps := &app.Deps{DB: db}
	service := NewService(deps)
	require.True(t, service.Ready())

	tenantID := "tenant-2"
	skuID := "sku-2"
	require.NoError(t, db.Select("id", "tenant_uuid").Create(&productskumodel.ProductSKU{ID: skuID, TenantUUID: tenantID}).Error)
	require.NoError(t, db.Exec(`INSERT INTO product_sku_channels (id, tenant_uuid, sku_id, channel_code, channel_sku_id, status) VALUES (?, ?, ?, ?, ?, ?)`,
		"ch-2", tenantID, skuID, "official", "OFF-2", "pending").Error)

	ctx := context.Background()
	ctx = authx.ContextWithTenantUUID(ctx, tenantID)

	_, err := service.PublishChannelMapping(ctx, skuID, ChannelPublishRequest{
		ChannelCode: "official",
		Force:       false,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInventoryRequiredForPublish)
}
