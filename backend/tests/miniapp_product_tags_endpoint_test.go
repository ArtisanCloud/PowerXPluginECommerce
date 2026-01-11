package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	spu "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/miniapp"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMiniAppProductTagsEndpointReturnsDistinctTags(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	tenantUUID := "00000000-0000-0000-0000-000000000111"
	tenantCtx := middleware.ContextWithTenantUUID(ctx, tenantUUID)
	testutil.SeedProductCategories(t, db, tenantUUID, "cat-a", "cat-b")

	_ = createPublishedSPU(t, spuSvc, tenantCtx, createPublishedSPUInput{
		Code:         "TAGS-001",
		Name:         "标签商品A",
		CategoryID:   "cat-a",
		CategoryPath: "root/cat-a",
		Tags:         []string{"bestsellers", "imported"},
	})
	_ = createPublishedSPU(t, spuSvc, tenantCtx, createPublishedSPUInput{
		Code:         "TAGS-002",
		Name:         "标签商品B",
		CategoryID:   "cat-b",
		CategoryPath: "root/cat-b",
		Tags:         []string{"bestsellers"},
	})

	engine := gin.New()
	root := engine.Group("/api/v1")
	miniapp.RegisterRoutes(root, deps)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/mini-app/products/tags?limit=50", nil)
	req.Header.Set("X-Tenant-UUID", tenantUUID)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Items []struct {
				Tag   string `json:"tag"`
				Count int    `json:"count"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Success)

	tags := map[string]int{}
	for _, it := range body.Data.Items {
		tags[it.Tag] = it.Count
	}
	require.Equal(t, 2, tags["bestsellers"])
	require.Equal(t, 1, tags["imported"])
}
