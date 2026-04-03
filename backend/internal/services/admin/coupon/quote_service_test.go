package coupon

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQuoteService_RulesAndStacking(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_quote_service?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS coupon_templates (
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
		)`,
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
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}
	now := time.Now().UTC()
	thresholdPercent, _ := json.Marshal(map[string]any{"percent": 10.0})
	scopeItem, _ := json.Marshal(map[string]any{"level": "item", "sku_ids": []string{"sku-1"}})
	stackItem, _ := json.Marshal(map[string]any{"priority": 10, "stackable": true, "level": "item"})

	thresholdAmount, _ := json.Marshal(map[string]any{"amount_minor": 500.0, "min_amount_minor": 1000.0})
	scopeOrder, _ := json.Marshal(map[string]any{"level": "order"})
	stackOrder, _ := json.Marshal(map[string]any{"priority": 20, "stackable": true, "level": "order"})
	emptyJSON, _ := json.Marshal(map[string]any{})

	templates := []couponmodel.CouponTemplate{
		{ID: "tpl-item", TenantUUID: "tenant-1", Code: "ITEM10", Name: "item10", CouponType: "percent", ThresholdRule: thresholdPercent, ScopeRule: scopeItem, StackingRule: stackItem, RefundRule: emptyJSON, ValidFrom: now.Add(-time.Hour), ValidTo: now.Add(time.Hour), Status: "active"},
		{ID: "tpl-order", TenantUUID: "tenant-1", Code: "ORD500", Name: "ord500", CouponType: "amount", ThresholdRule: thresholdAmount, ScopeRule: scopeOrder, StackingRule: stackOrder, RefundRule: emptyJSON, ValidFrom: now.Add(-time.Hour), ValidTo: now.Add(time.Hour), Status: "active"},
	}
	for _, tpl := range templates {
		if err := db.Create(&tpl).Error; err != nil {
			t.Fatalf("seed template: %v", err)
		}
	}
	assets := []couponmodel.CouponAsset{
		{ID: "asset-1", TenantUUID: "tenant-1", TemplateID: "tpl-item", UserID: "user-1", CouponCode: "C1", Status: "available"},
		{ID: "asset-2", TenantUUID: "tenant-1", TemplateID: "tpl-order", UserID: "user-1", CouponCode: "C2", Status: "available"},
	}
	for _, asset := range assets {
		if err := db.Create(&asset).Error; err != nil {
			t.Fatalf("seed asset: %v", err)
		}
	}

	svc := NewQuoteService(&app.Deps{DB: db})
	out, err := svc.Quote(ctx, QuoteInput{
		TenantUUID: "tenant-1",
		UserID:     "user-1",
		Channel:    "official",
		CouponIDs:  []string{"asset-1", "asset-2", "missing"},
		Items: []QuoteItemInput{
			{LineID: "l1", SKUID: "sku-1", Qty: 1, UnitPriceMinor: 1000},
			{LineID: "l2", SKUID: "sku-2", Qty: 1, UnitPriceMinor: 500},
		},
		Now:      now,
		Currency: "CNY",
	})
	if err != nil {
		t.Fatalf("quote failed: %v", err)
	}
	if out.BaseTotalMinor != 1500 {
		t.Fatalf("expected base 1500, got %d", out.BaseTotalMinor)
	}
	if out.DiscountTotalMinor != 600 {
		t.Fatalf("expected discount 600, got %d", out.DiscountTotalMinor)
	}
	if out.PayableTotalMinor != 900 {
		t.Fatalf("expected payable 900, got %d", out.PayableTotalMinor)
	}
	if len(out.AppliedCoupons) != 2 {
		t.Fatalf("expected 2 applied coupons, got %d", len(out.AppliedCoupons))
	}
	if len(out.RejectedCoupons) == 0 {
		t.Fatalf("expected rejected coupon for missing asset")
	}
}
