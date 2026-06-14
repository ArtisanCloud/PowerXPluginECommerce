package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func BenchmarkPromotionQuote(b *testing.B) {
	ctx := context.Background()
	svc := buildPromotionQuoteFixture(b)
	input := promotionQuoteInput()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Quote(ctx, input); err != nil {
			b.Fatalf("quote failed: %v", err)
		}
	}
}

func buildPromotionQuoteFixture(tb testing.TB) *promotionsvc.QuoteService {
	tb.Helper()
	models.ForceSchemaForTests("")
	dsn := fmt.Sprintf("file:promotion_quote_perf_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		tb.Fatalf("open db: %v", err)
	}
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
		`CREATE INDEX IF NOT EXISTS idx_promotion_campaign_active ON promotion_campaigns(tenant_uuid, status, valid_from, valid_to)`,
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
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			tb.Fatalf("create table: %v", err)
		}
	}
	now := time.Now().UTC()
	for i := 0; i < 40; i++ {
		condition, _ := json.Marshal(map[string]any{"min_order_amount_minor": 500})
		scope, _ := json.Marshal(map[string]any{"scope_type": "all"})
		action, _ := json.Marshal(map[string]any{"discount_amount_minor": int64(10 + i)})
		stacking, _ := json.Marshal(map[string]any{"priority": i + 1, "stackable": true, "stackable_with_coupon": true, "exclusion_group": fmt.Sprintf("g-%d", i%8)})
		row := promotionmodel.Campaign{
			ID: fmt.Sprintf("promo-%d", i), TenantUUID: "tenant-1", Code: fmt.Sprintf("P%03d", i), Name: "promo",
			PromotionType: promotionmodel.TypeAmountOff, ConditionRule: condition, ScopeRule: scope, ActionRule: action, StackingRule: stacking,
			ValidFrom: now.Add(-time.Hour), ValidTo: now.Add(time.Hour), Status: promotionmodel.StatusActive,
		}
		if err := db.Create(&row).Error; err != nil {
			tb.Fatalf("seed promotion: %v", err)
		}
	}
	return promotionsvc.NewQuoteService(&app.Deps{DB: db})
}

func promotionQuoteInput() promotionsvc.QuoteInput {
	return promotionsvc.QuoteInput{
		TenantUUID: "tenant-1", Channel: "official", Currency: "CNY", SubmittedAt: time.Now().UTC(),
		Items: []promotionsvc.QuoteItemInput{
			{LineID: "l1", SKUID: "sku-1", Qty: 1, UnitPriceMinor: 1200},
			{LineID: "l2", SKUID: "sku-2", Qty: 2, UnitPriceMinor: 800},
			{LineID: "l3", SKUID: "sku-3", Qty: 1, UnitPriceMinor: 500},
		},
	}
}
