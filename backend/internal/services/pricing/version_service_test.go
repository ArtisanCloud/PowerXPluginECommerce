package pricing

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	coremodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	pricingModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/pricing"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPublishNewVersionTerminatesOldVersion(t *testing.T) {
	db := openPricingTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	pbSvc := NewPricebookService(deps)
	verSvc := NewVersionService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pb, err := pbSvc.Create(ctx, CreatePricebookInput{
		Code:     "base",
		Name:     "Base",
		Type:     "sales",
		Currency: "CNY",
		Actor:    "tester",
	})
	if err != nil {
		t.Fatalf("create pricebook: %v", err)
	}
	if pb.CurrentVersionID == nil || *pb.CurrentVersionID == "" {
		t.Fatalf("expected v1 draft created")
	}

	v1, err := verSvc.Publish(ctx, PublishVersionInput{
		PricebookID: pb.ID,
		VersionID:   *pb.CurrentVersionID,
		Actor:       "tester",
	})
	if err != nil {
		t.Fatalf("publish v1: %v", err)
	}
	if v1.State != "active" {
		t.Fatalf("expected v1 active, got %s", v1.State)
	}

	v2Draft, err := verSvc.CreateDraft(ctx, CreateVersionInput{PricebookID: pb.ID, Actor: "tester"})
	if err != nil {
		t.Fatalf("create v2 draft: %v", err)
	}

	eff := time.Now().UTC()
	v2, err := verSvc.Publish(ctx, PublishVersionInput{
		PricebookID: pb.ID,
		VersionID:   v2Draft.ID,
		EffectiveAt: &eff,
		Actor:       "tester",
	})
	if err != nil {
		t.Fatalf("publish v2: %v", err)
	}
	if v2.State != "active" {
		t.Fatalf("expected v2 active, got %s", v2.State)
	}

	var v1Reload pricingModel.PricebookVersion
	if err := db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, v1.ID).First(&v1Reload).Error; err != nil {
		t.Fatalf("reload v1: %v", err)
	}
	if v1Reload.State != "expired" {
		t.Fatalf("expected v1 expired after publishing v2, got %s", v1Reload.State)
	}
	if v1Reload.ExpiresAt == nil {
		t.Fatalf("expected v1 expires_at set")
	}
	if v1Reload.ExpiresAt.After(eff) {
		t.Fatalf("expected v1 expires_at <= v2 effective_at")
	}

	var pbReload pricingModel.Pricebook
	if err := db.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, pb.ID).First(&pbReload).Error; err != nil {
		t.Fatalf("reload pricebook: %v", err)
	}
	if pbReload.CurrentVersionID == nil || *pbReload.CurrentVersionID != v2.ID {
		t.Fatalf("expected current_version_id updated to v2")
	}
}

func TestPublishConcurrentKeepsSingleActiveVersion(t *testing.T) {
	db := openPricingTestDB(t)
	deps := &app.Deps{Ctx: context.Background(), DB: db}
	pbSvc := NewPricebookService(deps)
	verSvc := NewVersionService(deps)

	tenantUUID := uuid.NewString()
	ctx := authx.ContextWithTenantUUID(context.Background(), tenantUUID)

	pb, err := pbSvc.Create(ctx, CreatePricebookInput{
		Code:     "base",
		Name:     "Base",
		Type:     "sales",
		Currency: "CNY",
		Actor:    "tester",
	})
	if err != nil {
		t.Fatalf("create pricebook: %v", err)
	}

	// Publish v1 so later publishes have something to terminate.
	if _, err := verSvc.Publish(ctx, PublishVersionInput{PricebookID: pb.ID, VersionID: *pb.CurrentVersionID, Actor: "tester"}); err != nil {
		t.Fatalf("publish v1: %v", err)
	}

	v2, err := verSvc.CreateDraft(ctx, CreateVersionInput{PricebookID: pb.ID, Actor: "tester"})
	if err != nil {
		t.Fatalf("create v2: %v", err)
	}
	v3, err := verSvc.CreateDraft(ctx, CreateVersionInput{PricebookID: pb.ID, Actor: "tester"})
	if err != nil {
		t.Fatalf("create v3: %v", err)
	}

	eff := time.Now().UTC()
	var wg sync.WaitGroup
	wg.Add(2)

	var err1, err2 error
	go func() {
		defer wg.Done()
		_, err1 = verSvc.Publish(ctx, PublishVersionInput{PricebookID: pb.ID, VersionID: v2.ID, EffectiveAt: &eff, Actor: "tester"})
	}()
	go func() {
		defer wg.Done()
		_, err2 = verSvc.Publish(ctx, PublishVersionInput{PricebookID: pb.ID, VersionID: v3.ID, EffectiveAt: &eff, Actor: "tester"})
	}()
	wg.Wait()

	okCount := 0
	conflictCount := 0
	for _, err := range []error{err1, err2} {
		if err == nil {
			okCount++
			continue
		}
		var se *Error
		if errors.As(err, &se) && se.Code == CodePublishConflict {
			conflictCount++
			continue
		}
		t.Fatalf("unexpected publish error: %v", err)
	}
	if okCount != 1 || conflictCount != 1 {
		t.Fatalf("expected 1 success + 1 conflict, got success=%d conflict=%d (err1=%v err2=%v)", okCount, conflictCount, err1, err2)
	}

	var activeCount int64
	if err := db.WithContext(ctx).Model(&pricingModel.PricebookVersion{}).
		Where("tenant_uuid = ? AND pricebook_id = ? AND state = ?", tenantUUID, pb.ID, "active").
		Count(&activeCount).Error; err != nil {
		t.Fatalf("count active: %v", err)
	}
	if activeCount != 1 {
		t.Fatalf("expected exactly 1 active version, got %d", activeCount)
	}
}

func openPricingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	coremodels.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// NOTE: Pricing models use Postgres defaults like gen_random_uuid().
	// For sqlite-based unit tests we create a minimal compatible schema manually.
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
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_pricebook_tenant_code ON pricebooks(tenant_uuid, code);`,
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
		`CREATE UNIQUE INDEX IF NOT EXISTS uk_pricebook_version ON pricebook_versions(tenant_uuid, pricebook_id, version);`,
		`CREATE INDEX IF NOT EXISTS idx_pricebook_versions_pb_state ON pricebook_versions(tenant_uuid, pricebook_id, state);`,
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
		if execErr := db.Exec(stmt).Error; execErr != nil {
			t.Fatalf("init schema: %v", execErr)
		}
	}
	return db
}
