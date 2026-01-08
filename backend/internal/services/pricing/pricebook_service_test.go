package pricing

import (
	"context"
	"errors"
	"testing"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestEnsureBasePricebookCreatesActiveBase(t *testing.T) {
	db := openPricingPricebookTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewPricebookService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pb, err := svc.EnsureBasePricebook(ctx, "CNY", "tester")
	if err != nil {
		t.Fatalf("ensure base: %v", err)
	}
	if pb == nil || pb.Code != BasePricebookCode || pb.Status != "active" || pb.Currency != "CNY" {
		t.Fatalf("unexpected base pricebook: %+v", pb)
	}
	if pb.CurrentVersionID == nil || *pb.CurrentVersionID == "" {
		t.Fatalf("expected current_version_id set")
	}

	var pbCount int64
	if err := db.Model(&pricingModel.Pricebook{}).
		Where("tenant_uuid = ? AND code = ?", tenantUUID, BasePricebookCode).
		Count(&pbCount).Error; err != nil {
		t.Fatalf("count pricebooks: %v", err)
	}
	if pbCount != 1 {
		t.Fatalf("expected 1 base pricebook, got %d", pbCount)
	}

	var activeVerCount int64
	if err := db.Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND pricebook_id = ? AND state = ?", tenantUUID, pb.ID, "active").
		Count(&activeVerCount).Error; err != nil {
		t.Fatalf("count active versions: %v", err)
	}
	if activeVerCount != 1 {
		t.Fatalf("expected 1 active base version, got %d", activeVerCount)
	}
}

func TestEnsureBasePricebookIsIdempotent(t *testing.T) {
	db := openPricingPricebookTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewPricebookService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	if _, err := svc.EnsureBasePricebook(ctx, "CNY", "tester"); err != nil {
		t.Fatalf("ensure base #1: %v", err)
	}
	if _, err := svc.EnsureBasePricebook(ctx, "CNY", "tester"); err != nil {
		t.Fatalf("ensure base #2: %v", err)
	}

	var pbCount int64
	if err := db.Model(&pricingModel.Pricebook{}).
		Where("tenant_uuid = ? AND code = ?", tenantUUID, BasePricebookCode).
		Count(&pbCount).Error; err != nil {
		t.Fatalf("count pricebooks: %v", err)
	}
	if pbCount != 1 {
		t.Fatalf("expected 1 base pricebook, got %d", pbCount)
	}
}

func TestBasePricebookCannotBeArchived(t *testing.T) {
	db := openPricingPricebookTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewPricebookService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pb, err := svc.EnsureBasePricebook(ctx, "CNY", "tester")
	if err != nil {
		t.Fatalf("ensure base: %v", err)
	}

	status := "archived"
	_, err = svc.Update(ctx, UpdatePricebookInput{
		PricebookID: pb.ID,
		Status:      &status,
		Actor:       "tester",
	})
	if err == nil {
		t.Fatalf("expected forbidden archive, got nil")
	}
	var se *Error
	if !errors.As(err, &se) || se.Code != CodeForbidden {
		t.Fatalf("expected %s, got %v", CodeForbidden, err)
	}
}

func TestBasePricebookCannotBeDeleted(t *testing.T) {
	db := openPricingPricebookTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewPricebookService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pb, err := svc.EnsureBasePricebook(ctx, "CNY", "tester")
	if err != nil {
		t.Fatalf("ensure base: %v", err)
	}

	err = svc.Delete(ctx, DeletePricebookInput{
		PricebookID: pb.ID,
		Actor:       "tester",
	})
	if err == nil {
		t.Fatalf("expected forbidden delete, got nil")
	}
	var se *Error
	if !errors.As(err, &se) || se.Code != CodeForbidden {
		t.Fatalf("expected %s, got %v", CodeForbidden, err)
	}
}

func TestCreateAllowsBaseCodeWhenMissing(t *testing.T) {
	db := openPricingPricebookTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	svc := NewPricebookService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pb, err := svc.Create(ctx, CreatePricebookInput{
		Code:     BasePricebookCode,
		Name:     "Base Pricebook",
		Type:     "sales",
		Currency: "CNY",
		Actor:    "tester",
	})
	if err != nil {
		t.Fatalf("create base: %v", err)
	}
	if pb == nil || pb.Code != BasePricebookCode {
		t.Fatalf("unexpected created pricebook: %+v", pb)
	}
	if pb.CurrentVersionID == nil || *pb.CurrentVersionID == "" {
		t.Fatalf("expected current_version_id set")
	}
	var v pricingModel.PricebookVersion
	if err := db.Where("tenant_uuid = ? AND id = ?", tenantUUID, *pb.CurrentVersionID).First(&v).Error; err != nil {
		t.Fatalf("load version: %v", err)
	}
	if v.State != "draft" {
		t.Fatalf("expected draft version, got %s", v.State)
	}
}

func openPricingPricebookTestDB(t *testing.T) *gorm.DB {
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
		`CREATE TABLE IF NOT EXISTS pricebook_audit_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			action TEXT NOT NULL,
			actor TEXT,
			payload TEXT,
			created_at DATETIME
		);`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	return db
}
