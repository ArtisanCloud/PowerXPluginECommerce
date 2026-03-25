package logistics

import (
	"context"
	"strings"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRateTemplateService_UpsertRejectedWhenPublished(t *testing.T) {
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:logistics_rate_template_service?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS logistics_rate_templates (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		name TEXT NOT NULL,
		currency TEXT NOT NULL,
		status TEXT NOT NULL,
		version INTEGER NOT NULL,
		channels JSON,
		rules JSON,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uk_logistics_rate_template ON logistics_rate_templates(tenant_uuid, name)`).Error)

	svc := NewRateTemplateService(&app.Deps{DB: db})
	tenantUUID := "tenant-1"
	ctx := context.Background()

	created, err := svc.Upsert(ctx, tenantUUID, UpsertRateTemplateRequest{
		Name:     "全国标准模板",
		Currency: "CNY",
	})
	require.NoError(t, err)
	require.Equal(t, "draft", created.Status)

	published, err := svc.Publish(ctx, tenantUUID, created.ID)
	require.NoError(t, err)
	require.Equal(t, "published", published.Status)

	_, err = svc.Upsert(ctx, tenantUUID, UpsertRateTemplateRequest{
		ID:       created.ID,
		Name:     "全国标准模板-v2",
		Currency: "CNY",
	})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "read-only"))
}
