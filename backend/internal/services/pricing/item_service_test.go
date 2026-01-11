package pricing

import (
	"context"
	"errors"
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

func TestUpsertItemsRejectsNonDraftVersion(t *testing.T) {
	db := openPricingItemsTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewItemService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pricebookID := uuid.NewString()
	versionID := uuid.NewString()

	now := time.Now().UTC()
	if err := db.WithContext(ctx).Exec(
		`INSERT INTO pricebook_versions (id, tenant_uuid, pricebook_id, version, state, effective_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		versionID, tenantUUID, pricebookID, 1, "active", now, now, now,
	).Error; err != nil {
		t.Fatalf("seed version: %v", err)
	}

	_, err := svc.UpsertItems(ctx, UpsertItemsInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		Items: []ItemInput{
			{SKUID: uuid.NewString(), BaseAmountMinor: ptrI64(100)},
		},
		Actor: "tester",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	var se *Error
	if !errors.As(err, &se) || se.Code != CodeVersionNotEditable {
		t.Fatalf("expected %s, got %v", CodeVersionNotEditable, err)
	}
}

func TestUpsertItemsValidatesNegativeAmount(t *testing.T) {
	db := openPricingItemsTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewItemService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pricebookID := uuid.NewString()
	versionID := uuid.NewString()

	now := time.Now().UTC()
	if err := db.WithContext(ctx).Exec(
		`INSERT INTO pricebook_versions (id, tenant_uuid, pricebook_id, version, state, effective_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		versionID, tenantUUID, pricebookID, 1, "draft", now, now, now,
	).Error; err != nil {
		t.Fatalf("seed version: %v", err)
	}

	_, err := svc.UpsertItems(ctx, UpsertItemsInput{
		PricebookID: pricebookID,
		VersionID:   versionID,
		Items: []ItemInput{
			{SKUID: uuid.NewString(), BaseAmountMinor: ptrI64(-1)},
		},
		Actor: "tester",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	var se *Error
	if !errors.As(err, &se) || se.Code != CodeInvalidArgument {
		t.Fatalf("expected %s, got %v", CodeInvalidArgument, err)
	}
}

func openPricingItemsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
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

func ptrI64(v int64) *int64 { return &v }
