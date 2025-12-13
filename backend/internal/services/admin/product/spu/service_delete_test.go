package spu

import (
	"context"
	"strings"
	"testing"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestSPUSoftDelete(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{NameReplacer: strings.NewReplacer("SPU", "Spu")},
	})
	require.NoError(t, err)
	createTables(t, db)
	deps := &app.Deps{DB: db, Ctx: ctx}
	svc := NewService(deps)
	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-delete")

	spu := &productmodel.SPU{
		ID:            "spu-delete",
		TenantUUID:    "tenant-delete",
		Code:          "SPU-DEL",
		Name:          "Delete Me",
		Type:          "one_time",
		CategoryID:    "cat",
		CategoryPath:  "root/cat",
		DefaultLocale: "zh-CN",
		Status:        "draft",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, db.Create(spu).Error)

	_, err = svc.Delete(tenantCtx, spu.ID, DeleteRequest{Reason: "不再需要"})
	require.NoError(t, err)

	var deleted productmodel.SPU
	require.NoError(t, db.Unscoped().Where("id = ?", spu.ID).First(&deleted).Error)
	require.True(t, deleted.DeletedAt.Valid)

	spuPublished := *spu
	spuPublished.ID = "spu-published"
	spuPublished.Status = "published"
	spuPublished.DeletedAt.Valid = false
	require.NoError(t, db.Create(&spuPublished).Error)
	_, err = svc.Delete(tenantCtx, spuPublished.ID, DeleteRequest{Reason: "禁删"})
	require.Error(t, err)
}
