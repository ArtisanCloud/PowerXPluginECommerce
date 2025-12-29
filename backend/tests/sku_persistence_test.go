package tests

import (
	"context"
	"testing"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/tests/testutil"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestUpsertSkusAllowsMultipleManualEntries(t *testing.T) {
	ctx := context.Background()
	db, isPostgres := testutil.NewIsolatedDB(t)
	testutil.EnsureProductTables(t, db, isPostgres)
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	deps := &app.Deps{DB: db, Ctx: ctx}
	skuSvc := productskuservice.NewService(deps)
	require.True(t, skuSvc.Ready())

	tenantID := "tenant-specless"
	tenantCtx := middleware.ContextWithTenantUUID(ctx, tenantID)

	result, err := skuSvc.UpsertSkus(tenantCtx, productskuservice.SkuUpsertRequest{
		SKUs: []productskuservice.SkuUpsertPayload{
			{
				SPUID:       "spu-specless",
				SKUCode:     "SKU-001",
				Status:      "online",
				Specs:       nil,
				Barcode:     "BAR-001",
				MinOrderQty: 1,
			},
			{
				SPUID:       "spu-specless",
				SKUCode:     "SKU-002",
				Status:      "offline",
				Specs:       nil,
				MinOrderQty: 5,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, result.Created)
	require.Len(t, result.Skipped, 0)

	var skuCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&productskumodel.ProductSKU{}).
		Where("tenant_uuid = ? AND spu_id = ?", tenantID, "spu-specless").
		Count(&skuCount).Error)
	require.Equal(t, int64(2), skuCount, "should persist both SKU rows even without specs")
}

func TestSKULinkServiceReplaceSyncsProductSkus(t *testing.T) {
	ctx := context.Background()
	db, isPostgres := testutil.NewIsolatedDB(t)
	testutil.EnsureProductTables(t, db, isPostgres)
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	deps := &app.Deps{DB: db, Ctx: ctx}
	linkSvc := spuservice.NewSKULinkService(deps)
	require.NotNil(t, linkSvc)

	tenantID := "tenant-link"
	tenantCtx := middleware.ContextWithTenantUUID(ctx, tenantID)
	spuID := "spu-link"
	versionID := "version-link"

	seedSPU(t, db, tenantID, spuID, versionID)

	firstPayload := spuservice.ReplaceSKUsRequest{
		Items: []spuservice.SKULinkInput{
			{
				Code: "LINK-SKU-A",
				Name: "Link SKU A",
				Attributes: map[string]any{
					"specs": []map[string]any{
						{"spec_id": "tier", "spec_name": "Tier", "value_id": "a", "value_name": "Basic"},
					},
					"defaults": map[string]any{"min_order_qty": 1},
				},
			},
		},
	}

	_, err := linkSvc.Replace(tenantCtx, spuID, firstPayload)
	require.NoError(t, err)
	assertSKUAndAttributeCount(t, ctx, db, tenantID, spuID, 1, 1)

	secondPayload := spuservice.ReplaceSKUsRequest{
		Items: []spuservice.SKULinkInput{
			{
				Code: "LINK-SKU-B",
				Name: "Link SKU B",
				Attributes: map[string]any{
					"specs": []map[string]any{
						{"spec_id": "tier", "spec_name": "Tier", "value_id": "b", "value_name": "Pro"},
					},
					"defaults": map[string]any{"min_order_qty": 2},
				},
			},
			{
				Code: "LINK-SKU-C",
				Name: "Link SKU C",
				Attributes: map[string]any{
					"specs": []map[string]any{
						{"spec_id": "tier", "spec_name": "Tier", "value_id": "c", "value_name": "Enterprise"},
					},
					"defaults": map[string]any{"min_order_qty": 3},
				},
			},
		},
	}

	_, err = linkSvc.Replace(tenantCtx, spuID, secondPayload)
	require.NoError(t, err)
	assertSKUAndAttributeCount(t, ctx, db, tenantID, spuID, 2, 2)

	thirdPayload := spuservice.ReplaceSKUsRequest{
		Items: []spuservice.SKULinkInput{
			{
				Code: "LINK-SKU-B",
				Name: "Link SKU B Prime",
				Attributes: map[string]any{
					"specs": []map[string]any{
						{"spec_id": "tier", "spec_name": "Tier", "value_id": "b", "value_name": "Pro"},
					},
					"defaults": map[string]any{"min_order_qty": 5},
				},
			},
			{
				Code: "LINK-SKU-D",
				Name: "Link SKU D",
				Attributes: map[string]any{
					"specs": []map[string]any{
						{"spec_id": "tier", "spec_name": "Tier", "value_id": "d", "value_name": "VIP"},
					},
					"defaults": map[string]any{"min_order_qty": 1},
				},
			},
		},
	}

	_, err = linkSvc.Replace(tenantCtx, spuID, thirdPayload)
	require.NoError(t, err)
	assertSKUAndAttributeCount(t, ctx, db, tenantID, spuID, 2, 2)

	skuB := fetchSkuByCode(t, db, "LINK-SKU-B")
	require.NotNil(t, skuB)
	require.Equal(t, 5, skuB.MinOrderQty)
}

func seedSPU(t *testing.T, db *gorm.DB, tenantID, spuID, versionID string) {
	t.Helper()

	initialPayload := datatypes.JSON([]byte(`{"skus":[]}`))
	require.NoError(t, db.Create(&productmodel.SPUVersion{
		ID:            versionID,
		TenantUUID:    tenantID,
		SPUID:         spuID,
		VersionNumber: 1,
		Status:        "published",
		Payload:       initialPayload,
	}).Error)

	currentVersionID := versionID
	require.NoError(t, db.Create(&productmodel.SPU{
		ID:               spuID,
		TenantUUID:       tenantID,
		Code:             "SPU-LINK-CODE",
		Name:             "SPU Link",
		Type:             "one_time",
		CategoryID:       "cat-link",
		CategoryPath:     "root/cat-link",
		DefaultLocale:    "zh-CN",
		Status:           "published",
		CurrentVersionID: &currentVersionID,
	}).Error)
}

func assertSKUAndAttributeCount(t *testing.T, ctx context.Context, db *gorm.DB, tenantID, spuID string, expectedSKU, expectedAttr int64) {
	t.Helper()

	var skuCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&productskumodel.ProductSKU{}).
		Where("tenant_uuid = ? AND spu_id = ?", tenantID, spuID).
		Count(&skuCount).Error)
	require.Equal(t, expectedSKU, skuCount)

	var attrCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&productskumodel.ProductSKUAttribute{}).
		Joins("JOIN product_skus ON product_sku_attributes.sku_id = product_skus.id").
		Where("product_skus.tenant_uuid = ? AND product_skus.spu_id = ?", tenantID, spuID).
		Count(&attrCount).Error)
	require.Equal(t, expectedAttr, attrCount)
}

func TestUpsertSkusUpdatesExistingEntries(t *testing.T) {
	ctx := context.Background()
	db, isPostgres := testutil.NewIsolatedDB(t)
	testutil.EnsureProductTables(t, db, isPostgres)
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	deps := &app.Deps{DB: db, Ctx: ctx}
	skuSvc := productskuservice.NewService(deps)
	require.True(t, skuSvc.Ready())

	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-upsert-update")
	spuID := "spu-upsert-update"

	_, err := skuSvc.UpsertSkus(tenantCtx, productskuservice.SkuUpsertRequest{
		SKUs: []productskuservice.SkuUpsertPayload{
			{
				SPUID:       spuID,
				SKUCode:     "SKU-UPDATE",
				Status:      "online",
				MinOrderQty: 1,
			},
		},
	})
	require.NoError(t, err)

	first := fetchSkuByCode(t, db, "SKU-UPDATE")
	require.NotNil(t, first)

	_, err = skuSvc.UpsertSkus(tenantCtx, productskuservice.SkuUpsertRequest{
		SKUs: []productskuservice.SkuUpsertPayload{
			{
				SPUID:       spuID,
				SKUCode:     "SKU-UPDATE",
				Status:      "offline",
				MinOrderQty: 10,
			},
		},
	})
	require.NoError(t, err)

	second := fetchSkuByCode(t, db, "SKU-UPDATE")
	require.NotNil(t, second)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, "offline", second.Status)
	require.Equal(t, 10, second.MinOrderQty)
}

func fetchSkuByCode(t *testing.T, db *gorm.DB, code string) *productskumodel.ProductSKU {
	var sku productskumodel.ProductSKU
	if err := db.Where("sku_code = ?", code).First(&sku).Error; err != nil {
		return nil
	}
	return &sku
}
