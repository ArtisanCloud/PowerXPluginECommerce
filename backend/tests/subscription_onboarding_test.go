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

func TestSubscriptionProductOnboardingFlow(t *testing.T) {
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

	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-subscription-flow")
	testutil.SeedProductCategories(t, db, "tenant-subscription-flow", "cat-sub")
	spuReq := spu.UpsertSPURequest{
		Code:            "SUB-PLAN-TEST",
		Name:            "订阅型商品",
		Type:            "subscription",
		CategoryID:      "cat-sub",
		CategoryPath:    "root/cat-sub",
		DefaultLocale:   "zh-CN",
		ResponsibleUser: "qa",
		Locales: []spu.LocaleContent{
			{Locale: "zh-CN", Title: "订阅型商品", Description: "E2E 测试用"},
		},
	}

	draft, err := spuSvc.CreateDraft(tenantCtx, spuReq)
	require.NoError(t, err)
	require.Equal(t, "draft", draft.Status)

	skuResult, err := skuSvc.UpsertSkus(tenantCtx, productsku.SkuUpsertRequest{
		SKUs: []productsku.SkuUpsertPayload{
			{
				SPUID:       draft.ID,
				SKUCode:     "SUB-PLAN-TEST-A",
				Status:      "online",
				MinOrderQty: 1,
				Specs: []productsku.SkuSpec{
					{SpecID: "billing", SpecName: "Billing Cycle", ValueID: "annual", ValueName: "年度"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, skuResult.Created)
	require.Len(t, skuResult.Summaries, 1)

	_, err = skuSvc.AdjustInventory(tenantCtx, skuResult.Summaries[0].ID, 1)
	require.NoError(t, err)

	submitted, err := spuSvc.Submit(tenantCtx, draft.ID, spu.SubmitRequest{})
	require.NoError(t, err)
	require.Equal(t, "reviewing", submitted.Status)
	require.NotEmpty(t, submitted.CurrentVersion)

	published, err := spuSvc.Publish(tenantCtx, draft.ID, spu.PublishRequest{
		VersionID: submitted.CurrentVersion,
		Channels:  []string{"official"},
	})
	require.NoError(t, err)
	require.Equal(t, "published", published.Status)

	list, err := skuSvc.ListSkus(tenantCtx, productsku.SkuListQuery{
		SPUID:  published.ID,
		Status: "online",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(list.Items))
	require.Equal(t, "SUB-PLAN-TEST-A", list.Items[0].SKUCode)
	require.Equal(t, "online", list.Items[0].Status)
}
