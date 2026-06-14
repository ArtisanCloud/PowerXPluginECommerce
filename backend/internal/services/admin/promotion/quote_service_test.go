package promotion

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQuoteService_RulesAndCouponStacking(t *testing.T) {
	db := setupPromotionDB(t, "file:promotion_quote?mode=memory&cache=shared")
	now := time.Now().UTC()
	seedPromotion(t, db, promotionmodel.Campaign{
		ID: "promo-1", TenantUUID: "tenant-1", Code: "ORDER1000", Name: "order", PromotionType: promotionmodel.TypeAmountOff,
		ConditionRule: mustJSON(map[string]any{"min_order_amount_minor": 1000}),
		ScopeRule:     mustJSON(map[string]any{"scope_type": "all", "channels": []string{"miniapp"}}),
		ActionRule:    mustJSON(map[string]any{"discount_amount_minor": 100}),
		StackingRule:  mustJSON(map[string]any{"priority": 100, "stackable": true, "stackable_with_coupon": false, "exclusion_group": "order_discount"}),
		ValidFrom:     now.Add(-time.Hour), ValidTo: now, Status: promotionmodel.StatusActive,
	})
	svc := NewQuoteService(&app.Deps{DB: db})
	out, err := svc.Quote(context.Background(), QuoteInput{
		TenantUUID: "tenant-1", Channel: "miniapp", Currency: "CNY", SubmittedAt: now,
		Items: []QuoteItemInput{{LineID: "l1", SKUID: "sku-1", Qty: 1, UnitPriceMinor: 1000}},
	})
	if err != nil {
		t.Fatalf("quote failed: %v", err)
	}
	if out.PromotionDiscountMinor != 100 || out.AfterPromotionTotalMinor != 900 {
		t.Fatalf("unexpected quote totals: %+v", out)
	}
	if out.CouponStackingAllowed {
		t.Fatalf("expected coupon stacking disabled")
	}
}

func setupPromotionDB(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	setupPromotionTables(t, db)
	return db
}

func setupPromotionTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS promotion_campaigns (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			promotion_type TEXT NOT NULL,
			condition_rule TEXT NOT NULL,
			scope_rule TEXT NOT NULL,
			action_rule TEXT NOT NULL,
			stacking_rule TEXT NOT NULL,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			status TEXT NOT NULL,
			created_by TEXT,
			updated_by TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS promotion_audit_logs (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			promotion_id TEXT NOT NULL,
			order_id TEXT,
			action TEXT NOT NULL,
			action_reason TEXT,
			request_id TEXT,
			created_by TEXT,
			payload TEXT NOT NULL,
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS order_promotion_snapshots (
			id TEXT PRIMARY KEY,
			tenant_uuid TEXT NOT NULL,
			order_id TEXT NOT NULL,
			currency TEXT NOT NULL,
			base_total_minor INTEGER NOT NULL,
			promotion_discount_minor INTEGER NOT NULL,
			after_promotion_total_minor INTEGER NOT NULL,
			applied_promotions TEXT NOT NULL,
			rejected_promotions TEXT NOT NULL,
			line_allocations TEXT NOT NULL,
			priced_at DATETIME NOT NULL,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create promotion table: %v", err)
		}
	}
}

func seedPromotion(t *testing.T, db *gorm.DB, row promotionmodel.Campaign) {
	t.Helper()
	if err := db.Create(&row).Error; err != nil {
		t.Fatalf("seed promotion: %v", err)
	}
}

func mustJSON(v any) []byte {
	raw, _ := json.Marshal(v)
	return raw
}
