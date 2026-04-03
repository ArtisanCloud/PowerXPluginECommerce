package coupon

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReleaseService_ReleaseAndIdempotent(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_release?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	setupCouponAssetTables(t, db)
	if err := db.Exec(`INSERT INTO coupon_assets (id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, reserved_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"asset-1", "tenant-1", "tpl-1", "user-1", "C1", "reserved", "ord-1", time.Now().UTC(), time.Now().UTC(), time.Now().UTC()).Error; err != nil {
		t.Fatalf("seed asset: %v", err)
	}

	svc := NewReleaseService(&app.Deps{DB: db})
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res, e := svc.ReleaseWithTx(ctx, tx, ReleaseInput{TenantUUID: "tenant-1", OrderID: "ord-1", Reason: "order_cancelled", Operator: "admin"})
		if e != nil {
			return e
		}
		if len(res.ReleasedAssetIDs) != 1 {
			t.Fatalf("expected 1 released asset, got %d", len(res.ReleasedAssetIDs))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("release tx failed: %v", err)
	}

	// idempotent second call
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_, e := svc.ReleaseWithTx(ctx, tx, ReleaseInput{TenantUUID: "tenant-1", OrderID: "ord-1", Reason: "order_cancelled", Operator: "admin"})
		return e
	})
	if err != nil {
		t.Fatalf("second release should be idempotent: %v", err)
	}

	var status string
	if err := db.Raw(`SELECT status FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, "tenant-1", "asset-1").Scan(&status).Error; err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != "available" {
		t.Fatalf("expected available status, got %s", status)
	}
	var reservedOrder string
	if err := db.Raw(`SELECT COALESCE(reserved_order_id, '') FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, "tenant-1", "asset-1").Scan(&reservedOrder).Error; err != nil {
		t.Fatalf("query reserved order: %v", err)
	}
	if reservedOrder != "" {
		t.Fatalf("expected empty reserved_order_id, got %s", reservedOrder)
	}
	var logs int64
	if err := db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND asset_id = ? AND action = ?`, "tenant-1", "asset-1", "release").Scan(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if logs != 1 {
		t.Fatalf("expected 1 release log, got %d", logs)
	}
}

func setupCouponAssetTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS coupon_assets (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			template_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			coupon_code TEXT NOT NULL,
			status TEXT NOT NULL,
			reserved_order_id TEXT,
			reserved_at DATETIME,
			redeemed_at DATETIME,
			refunded_at DATETIME,
			expired_at DATETIME,
			valid_from DATETIME,
			valid_to DATETIME,
			meta TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_usage_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			order_id TEXT,
			action TEXT NOT NULL,
			action_reason TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			request_id TEXT,
			created_by TEXT,
			created_at DATETIME,
			UNIQUE(tenant_uuid, action, idempotency_key)
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}
}
