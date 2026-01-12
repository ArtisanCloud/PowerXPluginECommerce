package tests

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	spu "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	productsku "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestStandardProductOnboardingFlow(t *testing.T) {
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
	spuSvc := spu.NewService(deps)
	skuSvc := productsku.NewService(deps)

	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-standard-flow")
	testutil.SeedProductCategories(t, db, "tenant-standard-flow", "cat-std")
	spuReq := spu.UpsertSPURequest{
		Code:            "STD-PRODUCT-001",
		Name:            "常规商品",
		Type:            "one_time",
		CategoryID:      "cat-std",
		CategoryPath:    "root/cat-std",
		DefaultLocale:   "zh-CN",
		ResponsibleUser: "pm",
		Locales: []spu.LocaleContent{
			{Locale: "zh-CN", Title: "常规商品", Description: "单次购买测试"},
		},
	}

	draft, err := spuSvc.CreateDraft(tenantCtx, spuReq)
	require.NoError(t, err)
	require.Equal(t, "draft", draft.Status)

	skuResult, err := skuSvc.UpsertSkus(tenantCtx, productsku.SkuUpsertRequest{
		SKUs: []productsku.SkuUpsertPayload{
			{
				SPUID:       draft.ID,
				SKUCode:     "STD-PRODUCT-001-A",
				Status:      "online",
				MinOrderQty: 1,
				Specs: []productsku.SkuSpec{
					{SpecID: "capacity", SpecName: "Capacity", ValueID: "single", ValueName: "单件"},
				},
			},
			{
				SPUID:       draft.ID,
				SKUCode:     "STD-PRODUCT-001-B",
				Status:      "offline",
				MinOrderQty: 10,
				Specs: []productsku.SkuSpec{
					{SpecID: "capacity", SpecName: "Capacity", ValueID: "bundle", ValueName: "十件装"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, skuResult.Created)
	require.Len(t, skuResult.Summaries, 2)

	_, err = skuSvc.AdjustInventory(tenantCtx, skuResult.Summaries[0].ID, 10)
	require.NoError(t, err)

	submitted, err := spuSvc.Submit(tenantCtx, draft.ID, spu.SubmitRequest{
		Comment: "提交初版",
	})
	require.NoError(t, err)
	require.Equal(t, "reviewing", submitted.Status)

	published, err := spuSvc.Publish(tenantCtx, draft.ID, spu.PublishRequest{
		VersionID: submitted.CurrentVersion,
		Channels:  []string{"official", "reseller"},
	})
	require.NoError(t, err)
	require.Equal(t, "published", published.Status)

	listAll, err := skuSvc.ListSkus(tenantCtx, productsku.SkuListQuery{
		SPUID: published.ID,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(listAll.Items))

	onlineOnly, err := skuSvc.ListSkus(tenantCtx, productsku.SkuListQuery{
		SPUID:  published.ID,
		Status: "online",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(onlineOnly.Items))
	require.Equal(t, "STD-PRODUCT-001-A", onlineOnly.Items[0].SKUCode)
}
