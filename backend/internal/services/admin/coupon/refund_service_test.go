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

func TestRefundService_RefundsRedeemedCouponWhenRuleAllows(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_refund_allowed?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	setupCouponRefundTables(t, db)
	now := time.Now().UTC()
	if err := db.Exec(`INSERT INTO coupon_templates (
		id, tenant_uuid, code, name, coupon_type, threshold_rule, scope_rule, stacking_rule, refund_rule, valid_from, valid_to, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"tpl-1", "tenant-1", "TPL1", "template", "amount", "{}", "{}", "{}", `{"return_coupon":true}`, now.Add(-24*time.Hour), now.Add(-time.Hour), "inactive", now, now,
	).Error; err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if err := db.Exec(`INSERT INTO coupon_assets (
		id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, redeemed_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"asset-1", "tenant-1", "tpl-1", "user-1", "C1", "redeemed", "ord-1", now, now, now,
	).Error; err != nil {
		t.Fatalf("seed asset: %v", err)
	}

	svc := NewRefundService(&app.Deps{DB: db})
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res, e := svc.RefundWithTx(ctx, tx, RefundInput{TenantUUID: "tenant-1", OrderID: "ord-1", RefundNo: "RF-1", Operator: "admin"})
		if e != nil {
			return e
		}
		if len(res.RefundedAssetIDs) != 1 {
			t.Fatalf("expected 1 refunded asset, got %d", len(res.RefundedAssetIDs))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("refund tx failed: %v", err)
	}

	var status string
	if err := db.Raw(`SELECT status FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, "tenant-1", "asset-1").Scan(&status).Error; err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != "refunded" {
		t.Fatalf("expected refunded status, got %s", status)
	}
	var logs int64
	if err := db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND asset_id = ? AND action = ?`, "tenant-1", "asset-1", "refund").Scan(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if logs != 1 {
		t.Fatalf("expected 1 refund log, got %d", logs)
	}
}

func TestRefundService_SkipsRedeemedCouponByDefault(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_refund_default_skip?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	setupCouponRefundTables(t, db)
	now := time.Now().UTC()
	if err := db.Exec(`INSERT INTO coupon_templates (
		id, tenant_uuid, code, name, coupon_type, threshold_rule, scope_rule, stacking_rule, refund_rule, valid_from, valid_to, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"tpl-1", "tenant-1", "TPL1", "template", "amount", "{}", "{}", "{}", `{}`, now.Add(-time.Hour), now.Add(time.Hour), "active", now, now,
	).Error; err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if err := db.Exec(`INSERT INTO coupon_assets (
		id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, redeemed_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"asset-1", "tenant-1", "tpl-1", "user-1", "C1", "redeemed", "ord-1", now, now, now,
	).Error; err != nil {
		t.Fatalf("seed asset: %v", err)
	}

	svc := NewRefundService(&app.Deps{DB: db})
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res, e := svc.RefundWithTx(ctx, tx, RefundInput{TenantUUID: "tenant-1", OrderID: "ord-1", RefundNo: "RF-1", Operator: "admin"})
		if e != nil {
			return e
		}
		if len(res.RefundedAssetIDs) != 0 || len(res.SkippedAssetIDs) != 1 {
			t.Fatalf("expected skip only, got refunded=%d skipped=%d", len(res.RefundedAssetIDs), len(res.SkippedAssetIDs))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("refund tx failed: %v", err)
	}

	var status string
	if err := db.Raw(`SELECT status FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, "tenant-1", "asset-1").Scan(&status).Error; err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != "redeemed" {
		t.Fatalf("expected redeemed status, got %s", status)
	}
	var logs int64
	if err := db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND asset_id = ? AND action = ?`, "tenant-1", "asset-1", "refund").Scan(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if logs != 0 {
		t.Fatalf("expected 0 refund logs, got %d", logs)
	}
}

func setupCouponRefundTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	setupCouponAssetTables(t, db)
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS coupon_templates (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		code TEXT NOT NULL,
		name TEXT NOT NULL,
		coupon_type TEXT NOT NULL,
		threshold_rule TEXT NOT NULL,
		scope_rule TEXT NOT NULL,
		stacking_rule TEXT NOT NULL,
		refund_rule TEXT NOT NULL,
		valid_from DATETIME NOT NULL,
		valid_to DATETIME NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create coupon_templates failed: %v", err)
	}
}
