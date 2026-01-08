package pricing

import (
	"context"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPricingQueryFallsBackToBaseWhenItemMissing(t *testing.T) {
	db := openPricingQueryTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewQueryService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	asOf := time.Now().UTC()
	skuID := uuid.NewString()
	channelID := "channel-1"

	basePB := uuid.NewString()
	baseV := uuid.NewString()
	scopedPB := uuid.NewString()
	scopedV := uuid.NewString()

	seedPricebook(t, db, tenantUUID, basePB, "base", "CNY")
	seedActiveVersion(t, db, tenantUUID, baseV, basePB, 1, asOf.Add(-time.Hour))
	seedItem(t, db, tenantUUID, uuid.NewString(), basePB, baseV, skuID, ptrI64(19900), nil, nil)

	seedPricebook(t, db, tenantUUID, scopedPB, "ch1", "CNY")
	seedActiveVersion(t, db, tenantUUID, scopedV, scopedPB, 1, asOf.Add(-time.Hour))
	seedScope(t, db, tenantUUID, scopedPB, "channel", channelID)

	out, err := svc.Query(ctx, QueryInput{
		SKUID:     skuID,
		Currency:  "CNY",
		ChannelID: &channelID,
		AsOf:      &asOf,
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if !out.Priced || out.AmountMinor == nil || *out.AmountMinor != 19900 {
		t.Fatalf("expected priced 19900, got %+v", out)
	}
	if out.Trace.Fallback != FallbackBaseItemFallback {
		t.Fatalf("expected fallback %s, got %s", FallbackBaseItemFallback, out.Trace.Fallback)
	}
	if out.Matched == nil || out.Matched.PricebookCode != "base" {
		t.Fatalf("expected matched base pricebook, got %+v", out.Matched)
	}
	if out.SourceField == nil || *out.SourceField != "base" {
		t.Fatalf("expected source_field=base, got %v", out.SourceField)
	}
}

func TestPricingQueryReturnsNoPriceOnScopeMismatchWhenNoOtherCandidate(t *testing.T) {
	db := openPricingQueryTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewQueryService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	asOf := time.Now().UTC()
	skuID := uuid.NewString()

	scopedPB := uuid.NewString()
	scopedV := uuid.NewString()
	seedPricebook(t, db, tenantUUID, scopedPB, "ch1", "CNY")
	seedActiveVersion(t, db, tenantUUID, scopedV, scopedPB, 1, asOf.Add(-time.Hour))
	seedScope(t, db, tenantUUID, scopedPB, "channel", "channel-1")
	seedItem(t, db, tenantUUID, uuid.NewString(), scopedPB, scopedV, skuID, ptrI64(100), nil, nil)

	channelID := "channel-2"
	out, err := svc.Query(ctx, QueryInput{
		SKUID:     skuID,
		Currency:  "CNY",
		ChannelID: &channelID,
		AsOf:      &asOf,
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if out.Priced {
		t.Fatalf("expected unpriced, got %+v", out)
	}
	if out.Trace.NoPriceReason == nil || *out.Trace.NoPriceReason != NoPriceScopeMismatch {
		t.Fatalf("expected no_price_reason=%s, got %v", NoPriceScopeMismatch, out.Trace.NoPriceReason)
	}
}

func openPricingQueryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS pricebooks (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			currency TEXT NOT NULL,
			description TEXT,
			status TEXT NOT NULL,
			current_version_id TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS pricebook_versions (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			pricebook_id TEXT NOT NULL,
			version INTEGER NOT NULL,
			state TEXT NOT NULL,
			effective_at DATETIME NOT NULL,
			expires_at DATETIME,
			published_at DATETIME,
			published_by TEXT,
			note TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS pricebook_scopes (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			pricebook_id TEXT NOT NULL,
			dimension TEXT NOT NULL,
			dimension_id TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS pricebook_items (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			pricebook_id TEXT NOT NULL,
			version_id TEXT NOT NULL,
			sku_id TEXT NOT NULL,
			base_amount_minor INTEGER,
			sale_amount_minor INTEGER,
			msrp_amount_minor INTEGER,
			cost_amount_minor INTEGER,
			min_amount_minor INTEGER,
			max_amount_minor INTEGER,
			tax_included BOOLEAN NOT NULL DEFAULT 0,
			meta TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_pricebook_item_version_sku ON pricebook_items(tenant_uuid, version_id, sku_id);`,
	}
	for _, stmt := range stmts {
		if execErr := db.Exec(stmt).Error; execErr != nil {
			t.Fatalf("init schema: %v", execErr)
		}
	}
	return db
}

func seedPricebook(t *testing.T, db *gorm.DB, tenantUUID, id, code, currency string) {
	t.Helper()
	now := time.Now().UTC()
	if err := db.Exec(
		`INSERT INTO pricebooks (id, tenant_uuid, code, name, type, currency, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, tenantUUID, code, code, "sales", currency, "active", now, now,
	).Error; err != nil {
		t.Fatalf("seed pricebook: %v", err)
	}
}

func seedActiveVersion(t *testing.T, db *gorm.DB, tenantUUID, versionID, pricebookID string, version int, effectiveAt time.Time) {
	t.Helper()
	now := time.Now().UTC()
	if err := db.Exec(
		`INSERT INTO pricebook_versions (id, tenant_uuid, pricebook_id, version, state, effective_at, published_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		versionID, tenantUUID, pricebookID, version, "active", effectiveAt, now, now, now,
	).Error; err != nil {
		t.Fatalf("seed version: %v", err)
	}
}

func seedScope(t *testing.T, db *gorm.DB, tenantUUID, pricebookID, dimension, dimensionID string) {
	t.Helper()
	now := time.Now().UTC()
	if err := db.Exec(
		`INSERT INTO pricebook_scopes (id, tenant_uuid, pricebook_id, dimension, dimension_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), tenantUUID, pricebookID, dimension, dimensionID, now, now,
	).Error; err != nil {
		t.Fatalf("seed scope: %v", err)
	}
}

func seedItem(t *testing.T, db *gorm.DB, tenantUUID, id, pricebookID, versionID, skuID string, base, sale, msrp *int64) {
	t.Helper()
	now := time.Now().UTC()
	if err := db.Exec(
		`INSERT INTO pricebook_items (id, tenant_uuid, pricebook_id, version_id, sku_id, base_amount_minor, sale_amount_minor, msrp_amount_minor, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, tenantUUID, pricebookID, versionID, skuID, base, sale, msrp, now, now,
	).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
}
