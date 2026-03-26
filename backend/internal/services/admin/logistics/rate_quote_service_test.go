package logistics

import (
	"context"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRateQuoteService_ZoneMatchOrderAndFallback(t *testing.T) {
	db := setupRateQuoteDB(t, "logistics_rate_quote_zone")
	svc := NewRateQuoteService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	resEast, err := svc.Quote(ctx, "tenant-1", QuoteRateRequest{
		TemplateID: "tpl-1",
		Region:     "华东",
		Weight:     3,
	})
	require.NoError(t, err)
	require.Equal(t, "华东", resEast.MatchedZone.Region)
	require.InDelta(t, 10, resEast.FeeAmount, 0.01)

	resFallback, err := svc.Quote(ctx, "tenant-1", QuoteRateRequest{
		TemplateID: "tpl-1",
		Region:     "西北",
		Weight:     3,
	})
	require.NoError(t, err)
	require.Equal(t, "default", resFallback.MatchedZone.Region)
	require.InDelta(t, 14, resFallback.FeeAmount, 0.01)
}

func TestRateQuoteService_BoundaryInputs(t *testing.T) {
	db := setupRateQuoteDB(t, "logistics_rate_quote_boundary")
	svc := NewRateQuoteService(&app.Deps{DB: db})
	ctx := authx.ContextWithTenantUUID(context.Background(), "tenant-1")

	resFree, err := svc.Quote(ctx, "tenant-1", QuoteRateRequest{
		TemplateID:  "tpl-2",
		Region:      "default",
		PieceCount:  5,
		OrderAmount: 120,
	})
	require.NoError(t, err)
	require.InDelta(t, 0, resFree.FeeAmount, 0.01)

	resPiece, err := svc.Quote(ctx, "tenant-1", QuoteRateRequest{
		TemplateID: "tpl-2",
		Region:     "default",
		PieceCount: 5,
	})
	require.NoError(t, err)
	// first=1 fee=4, remaining=4 with step=1 and add=1.5 => 10
	require.InDelta(t, 10, resPiece.FeeAmount, 0.01)
}

func setupRateQuoteDB(t *testing.T, name string) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
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
	require.NoError(t, db.Exec(`INSERT INTO logistics_rate_templates
		(id, tenant_uuid, name, currency, status, version, channels, rules, created_at, updated_at)
		VALUES
		(
			'tpl-1', 'tenant-1', '模板A', 'CNY', 'published', 1, '[]',
			'{
				"billing":"weight",
				"zones":[
					{"region":"华东","first_weight":1,"first_fee":6,"additional_weight":1,"additional_fee":2},
					{"region":"default","first_weight":1,"first_fee":8,"additional_weight":1,"additional_fee":3}
				]
			}',
			datetime('now'), datetime('now')
		),
		(
			'tpl-2', 'tenant-1', '模板B', 'CNY', 'published', 1, '[]',
			'{
				"billing":"piece",
				"defaultZone":{"region":"default","first_piece":1,"first_fee":4,"additional_piece":1,"additional_fee":1.5,"free_threshold":100}
			}',
			datetime('now'), datetime('now')
		),
		(
			'tpl-3', 'tenant-2', '模板C', 'CNY', 'published', 1, '[]',
			'{"billing":"weight","defaultZone":{"region":"default","first_weight":1,"first_fee":5}}',
			datetime('now'), datetime('now')
		)`).Error)
	return db
}
