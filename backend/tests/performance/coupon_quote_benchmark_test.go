package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func BenchmarkCouponQuote(b *testing.B) {
	ctx := context.Background()
	svc := buildQuoteServiceFixture(b)
	now := time.Now().UTC()
	input := couponsvc.QuoteInput{
		TenantUUID: "tenant-1",
		UserID:     "user-1",
		Channel:    "official",
		CouponIDs:  []string{"asset-1", "asset-2"},
		Items: []couponsvc.QuoteItemInput{
			{LineID: "l1", SKUID: "sku-1", Qty: 1, UnitPriceMinor: 1200},
			{LineID: "l2", SKUID: "sku-2", Qty: 1, UnitPriceMinor: 800},
		},
		Now:      now,
		Currency: "CNY",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Quote(ctx, input); err != nil {
			b.Fatalf("quote failed: %v", err)
		}
	}
}

func buildQuoteServiceFixture(tb testing.TB) *couponsvc.QuoteService {
	tb.Helper()
	models.ForceSchemaForTests("")
	dsn := fmt.Sprintf("file:coupon_quote_bench_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		tb.Fatalf("open db: %v", err)
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
			tb.Fatalf("create table failed: %v", err)
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
			tb.Fatalf("seed template: %v", err)
		}
	}
	assets := []couponmodel.CouponAsset{
		{ID: "asset-1", TenantUUID: "tenant-1", TemplateID: "tpl-item", UserID: "user-1", CouponCode: "C1", Status: "available"},
		{ID: "asset-2", TenantUUID: "tenant-1", TemplateID: "tpl-order", UserID: "user-1", CouponCode: "C2", Status: "available"},
	}
	for _, asset := range assets {
		if err := db.Create(&asset).Error; err != nil {
			tb.Fatalf("seed asset: %v", err)
		}
	}
	return couponsvc.NewQuoteService(&app.Deps{DB: db})
}
