package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	spu "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestSPUListRejectsInvalidSortAndOrder(t *testing.T) {
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
	svc := spu.NewService(deps)
	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-spu-list-sort-validate")

	_, err := svc.List(tenantCtx, spu.ListFilters{Status: "published", Sort: "not-a-sort", Page: 1, PageSize: 10})
	require.Error(t, err)
	require.True(t, errors.Is(err, spu.ErrInvalidSPUListSort))

	_, err = svc.List(tenantCtx, spu.ListFilters{Status: "published", Order: "up", Page: 1, PageSize: 10})
	require.Error(t, err)
	require.True(t, errors.Is(err, spu.ErrInvalidSPUListOrder))
}

func TestSPUListPriceSortRequiresPostgres(t *testing.T) {
	ctx := context.Background()
	db, isPostgres := testutil.NewIsolatedDB(t)
	testutil.EnsureProductTables(t, db, isPostgres)
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	if isPostgres {
		t.Skip("该用例仅验证非 postgres 方言下的保护行为")
	}

	deps := &app.Deps{DB: db, Ctx: ctx}
	svc := spu.NewService(deps)
	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-spu-list-sort-price")

	_, err := svc.List(tenantCtx, spu.ListFilters{Status: "published", Sort: "price", Page: 1, PageSize: 10})
	require.Error(t, err)
	require.True(t, errors.Is(err, spu.ErrSPUListRequiresPostgres))
}
