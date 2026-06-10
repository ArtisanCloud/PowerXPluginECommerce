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

func TestRedeemService_RedeemAndIdempotent(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_redeem?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	setupCouponAssetTables(t, db)
	if err := db.Exec(`INSERT INTO coupon_assets (id, tenant_uuid, template_id, user_id, coupon_code, status, reserved_order_id, reserved_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"asset-1", "tenant-1", "tpl-1", "user-1", "C1", "reserved", "ord-1", time.Now().UTC(), time.Now().UTC(), time.Now().UTC()).Error; err != nil {
		t.Fatalf("seed asset: %v", err)
	}

	svc := NewRedeemService(&app.Deps{DB: db})
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res, e := svc.RedeemWithTx(ctx, tx, RedeemInput{TenantUUID: "tenant-1", OrderID: "ord-1", Reason: "payment_success", Operator: "callback"})
		if e != nil {
			return e
		}
		if len(res.RedeemedAssetIDs) != 1 {
			t.Fatalf("expected 1 redeemed asset, got %d", len(res.RedeemedAssetIDs))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("redeem tx failed: %v", err)
	}

	// idempotent second call
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		_, e := svc.RedeemWithTx(ctx, tx, RedeemInput{TenantUUID: "tenant-1", OrderID: "ord-1", Reason: "payment_success", Operator: "callback"})
		return e
	})
	if err != nil {
		t.Fatalf("second redeem should be idempotent: %v", err)
	}

	var status string
	if err := db.Raw(`SELECT status FROM coupon_assets WHERE tenant_uuid = ? AND id = ?`, "tenant-1", "asset-1").Scan(&status).Error; err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != "redeemed" {
		t.Fatalf("expected redeemed status, got %s", status)
	}
	var logs int64
	if err := db.Raw(`SELECT COUNT(1) FROM coupon_usage_logs WHERE tenant_uuid = ? AND asset_id = ? AND action = ?`, "tenant-1", "asset-1", "redeem").Scan(&logs).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if logs != 1 {
		t.Fatalf("expected 1 redeem log, got %d", logs)
	}
}
