package spu

import (
	context "context"
	"strings"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestSPULifecycleCreateSubmitPublish(t *testing.T) {
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
	svc := NewService(deps)
	tenantCtx := middleware.ContextWithTenantUUID(ctx, "tenant-test")
	req := UpsertSPURequest{
		Code:            "SPU-001",
		Name:            "测试商品",
		Type:            "one_time",
		CategoryID:      "cat-1",
		CategoryPath:    "root/cat-1",
		DefaultLocale:   "zh-CN",
		ResponsibleUser: "ops-1",
		Locales: []LocaleContent{
			{Locale: "zh-CN", Title: "默认标题", Description: "描述"},
		},
	}
	draft, err := svc.CreateDraft(tenantCtx, req)
	require.NoError(t, err)
	require.Equal(t, "draft", draft.Status)
	require.NotEmpty(t, draft.CurrentVersion)

	updateReq := req
	updateReq.Name = "更新后的商品"
	updated, err := svc.UpdateDraft(tenantCtx, draft.ID, updateReq)
	require.NoError(t, err)
	require.Equal(t, "更新后的商品", updated.Name)

	submitted, err := svc.Submit(tenantCtx, draft.ID, SubmitRequest{})
	require.NoError(t, err)
	require.Equal(t, "reviewing", submitted.Status)
	require.NotEmpty(t, submitted.CurrentVersion)

	published, err := svc.Publish(tenantCtx, draft.ID, PublishRequest{VersionID: submitted.CurrentVersion, Channels: []string{"official"}})
	require.NoError(t, err)
	require.Equal(t, "published", published.Status)

	var entity productmodel.SPU
	require.NoError(t, db.WithContext(ctx).Where("id = ?", draft.ID).First(&entity).Error)
	require.Equal(t, "published", entity.Status)
}

func createTables(t *testing.T, db *gorm.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS product_spus (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			category_id TEXT NOT NULL,
			category_path TEXT NOT NULL,
			brand_id TEXT,
			default_locale TEXT NOT NULL,
			status TEXT NOT NULL,
			current_version_id TEXT,
			tags TEXT,
			responsible_user TEXT,
			channels_summary TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_versions (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			version_number INTEGER NOT NULL,
			status TEXT NOT NULL,
			payload TEXT,
			diff_summary TEXT,
			submitted_by TEXT,
			submitted_at DATETIME,
			approved_by TEXT,
			approved_at DATETIME,
			rollback_source_version TEXT,
			audit_log_id TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_approvals (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			version_id TEXT NOT NULL,
			chain_order INTEGER NOT NULL,
			role TEXT NOT NULL,
			status TEXT NOT NULL,
			comment TEXT,
			sla_due_at DATETIME,
			acted_by TEXT,
			acted_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_channels (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			channel TEXT NOT NULL,
			availability TEXT NOT NULL,
			publish_at DATETIME,
			withdraw_at DATETIME,
			content_override TEXT,
			audit_state TEXT,
			last_feedback TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS product_spu_audit_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			spu_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload TEXT,
			operator TEXT,
			created_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}
