package tests

import (
	"context"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	spu "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestSPUListSupportsCategoryAndTagFilters(t *testing.T) {
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
	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-spu-list-filter")
	testutil.SeedProductCategories(t, db, "tenant-spu-list-filter", "cat-a", "cat-b")

	spuA := createPublishedSPU(t, spuSvc, tenantCtx, createPublishedSPUInput{
		Code:         "TAG-CAT-001",
		Name:         "带标签与类目商品A",
		CategoryID:   "cat-a",
		CategoryPath: "root/cat-a",
		Tags:         []string{"bestsellers", "imported"},
	})
	spuB := createPublishedSPU(t, spuSvc, tenantCtx, createPublishedSPUInput{
		Code:         "TAG-CAT-002",
		Name:         "带标签商品B",
		CategoryID:   "cat-b",
		CategoryPath: "root/cat-b",
		Tags:         []string{"bestsellers"},
	})
	require.NotEmpty(t, spuA)
	require.NotEmpty(t, spuB)

	list, err := spuSvc.List(tenantCtx, spu.ListFilters{
		Status:             "published",
		CategoryPathPrefix: "root/cat-a",
		Tags:               []string{"bestsellers"},
		Page:               1,
		PageSize:           50,
	})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	require.Equal(t, "TAG-CAT-001", list.Items[0].Code)

	list, err = spuSvc.List(tenantCtx, spu.ListFilters{
		Status: "published",
		Tags:   []string{"imported"},
		Page:   1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Len(t, list.Items, 1)
	require.Equal(t, "TAG-CAT-001", list.Items[0].Code)

	list, err = spuSvc.List(tenantCtx, spu.ListFilters{
		Status:             "published",
		CategoryPathPrefix: "root/cat-b",
		Tags:               []string{"imported"},
		Page:               1,
		PageSize:           50,
	})
	require.NoError(t, err)
	require.Len(t, list.Items, 0)
}

type createPublishedSPUInput struct {
	Code         string
	Name         string
	CategoryID   string
	CategoryPath string
	Tags         []string
}

func createPublishedSPU(t *testing.T, svc *spu.Service, tenantCtx context.Context, input createPublishedSPUInput) string {
	t.Helper()

	req := spu.UpsertSPURequest{
		Code:            input.Code,
		Name:            input.Name,
		Type:            "one_time",
		CategoryID:      input.CategoryID,
		CategoryPath:    input.CategoryPath,
		DefaultLocale:   "zh-CN",
		Tags:            input.Tags,
		ResponsibleUser: "pm",
		Locales: []spu.LocaleContent{
			{Locale: "zh-CN", Title: input.Name, Description: "测试"},
		},
	}
	draft, err := svc.CreateDraft(tenantCtx, req)
	require.NoError(t, err)

	submitted, err := svc.Submit(tenantCtx, draft.ID, spu.SubmitRequest{Comment: "提交"})
	require.NoError(t, err)

	published, err := svc.Publish(tenantCtx, draft.ID, spu.PublishRequest{
		VersionID: submitted.CurrentVersion,
		Channels:  []string{"official"},
	})
	require.NoError(t, err)
	require.Equal(t, "published", published.Status)
	return published.ID
}
