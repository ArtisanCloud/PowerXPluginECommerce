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

func TestSPUWithdrawAllChannels(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	createTables(t, db)
	deps := &app.Deps{DB: db, Ctx: ctx}
	svc := NewService(deps)
	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-withdraw")

	spu := &productmodel.SPU{
		ID:            "spu-withdraw",
		TenantUUID:    "tenant-withdraw",
		Code:          "SPU-W",
		Name:          "Withdraw Test",
		Type:          "one_time",
		CategoryID:    "cat",
		CategoryPath:  "root/cat",
		DefaultLocale: "zh-CN",
		Status:        "published",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	require.NoError(t, db.Create(spu).Error)

	channels := []productmodel.ChannelVisibility{
		{
			ID:           "ch-official",
			TenantUUID:   spu.TenantUUID,
			SPUID:        spu.ID,
			Channel:      "official",
			Availability: "published",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "ch-douyin",
			TenantUUID:   spu.TenantUUID,
			SPUID:        spu.ID,
			Channel:      "douyin",
			Availability: "published",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	require.NoError(t, db.Create(&channels).Error)

	detail, err := svc.Withdraw(tenantCtx, spu.ID, WithdrawRequest{Reason: "库存消耗"})
	require.NoError(t, err)
	require.Equal(t, "offboarded", detail.Status)

	var remain int64
	require.NoError(t, db.Model(&productmodel.ChannelVisibility{}).
		Where("tenant_uuid = ? AND spu_id = ? AND availability <> ?", spu.TenantUUID, spu.ID, "offboarded").
		Count(&remain).Error)
	require.Equal(t, int64(0), remain)

	var updated productmodel.SPU
	require.NoError(t, db.Where("id = ?", spu.ID).First(&updated).Error)
	require.Equal(t, "offboarded", updated.Status)
}
