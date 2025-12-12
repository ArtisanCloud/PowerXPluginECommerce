package spu

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestVersionServiceDiffAndRollback(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			NameReplacer: strings.NewReplacer("SPU", "Spu"),
		},
	})
	require.NoError(t, err)
	createTables(t, db)

	deps := &app.Deps{DB: db, Ctx: ctx}
	versionSvc := NewVersionService(deps)
	require.NotNil(t, versionSvc)

	now := time.Now().UTC()
	spu := &productmodel.SPU{
		ID:              "spu-1",
		TenantUUID:      "tenant-version",
		Code:            "SPU-1",
		Name:            "Base Name",
		Type:            "one_time",
		CategoryID:      "cat-1",
		CategoryPath:    "root/cat-1",
		DefaultLocale:   "zh-CN",
		Status:          "published",
		CreatedAt:       now,
		UpdatedAt:       now,
		ChannelsSummary: datatypes.JSON([]byte("{}")),
	}
	require.NoError(t, db.Create(spu).Error)

	basePayload := encodeVersionPayload(map[string]any{
		"input": map[string]any{
			"name":          "Base Name",
			"categoryId":    "cat-1",
			"defaultLocale": "zh-CN",
		},
	})
	publishedVersion := &productmodel.SPUVersion{
		ID:            "ver-base",
		TenantUUID:    spu.TenantUUID,
		SPUID:         spu.ID,
		VersionNumber: 1,
		Status:        "published",
		Payload:       basePayload,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, db.Create(publishedVersion).Error)
	require.NoError(t, db.Model(spu).Updates(map[string]any{
		"current_version_id": publishedVersion.ID,
	}).Error)

	reviewPayload := encodeVersionPayload(map[string]any{
		"input": map[string]any{
			"name":          "Updated Name",
			"categoryId":    "cat-1",
			"defaultLocale": "zh-CN",
			"tags":          []string{"hot"},
		},
	})
	reviewVersion := &productmodel.SPUVersion{
		ID:            "ver-review",
		TenantUUID:    spu.TenantUUID,
		SPUID:         spu.ID,
		VersionNumber: 2,
		Status:        "reviewing",
		Payload:       reviewPayload,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, db.Create(reviewVersion).Error)

	approval := &productmodel.SPUApprovalRecord{
		ID:         "apr-1",
		TenantUUID: spu.TenantUUID,
		SPUID:      spu.ID,
		VersionID:  reviewVersion.ID,
		ChainOrder: 1,
		Role:       "qc",
		Status:     "pending",
		SLADueAt:   ptrTime(now.Add(48 * time.Hour)),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, db.Create(approval).Error)

	tenantCtx := middleware.ContextWithTenantUUID(ctx, spu.TenantUUID)

	detail, err := versionSvc.Get(tenantCtx, spu.ID, reviewVersion.ID)
	require.NoError(t, err)
	require.Equal(t, reviewVersion.ID, detail.ID)
	require.Equal(t, reviewVersion.VersionNumber, detail.VersionNumber)
	require.Equal(t, "reviewing", detail.Status)
	require.NotEmpty(t, detail.Diff)
	require.Equal(t, "input.name", detail.Diff[0].Field)
	require.Equal(t, "Base Name", detail.Diff[0].Before)
	require.Equal(t, "Updated Name", detail.Diff[0].After)
	require.Len(t, detail.Approvals, 1)

	list, err := versionSvc.List(tenantCtx, spu.ID, VersionListFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(2), list.Total)
	require.Equal(t, 2, len(list.Items))
	require.Equal(t, 2, list.Items[0].VersionNumber)

	rollbackDetail, err := versionSvc.Rollback(tenantCtx, spu.ID, RollbackRequest{
		TargetVersionID: publishedVersion.ID,
		Reason:          "rollback test",
	})
	require.NoError(t, err)
	require.Equal(t, "draft", rollbackDetail.Status)
	require.Equal(t, publishedVersion.ID, rollbackDetail.RollbackSourceID)

	var count int64
	require.NoError(t, db.Model(&productmodel.SPUVersion{}).Where("spu_id = ?", spu.ID).Count(&count).Error)
	require.Equal(t, int64(3), count)

	var updatedSPU productmodel.SPU
	require.NoError(t, db.Where("id = ?", spu.ID).First(&updatedSPU).Error)
	require.Equal(t, "draft", updatedSPU.Status)
	require.Equal(t, rollbackDetail.ID, derefString(updatedSPU.CurrentVersionID))
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
