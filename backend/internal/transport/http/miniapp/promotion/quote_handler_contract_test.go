package promotion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQuoteHandlerContract(t *testing.T) {
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:promotion_quote_handler?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
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
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	now := time.Now().UTC()
	raw, _ := json.Marshal(map[string]any{"min_order_amount_minor": 1000})
	scope, _ := json.Marshal(map[string]any{"scope_type": "all"})
	action, _ := json.Marshal(map[string]any{"discount_amount_minor": 100})
	stacking, _ := json.Marshal(map[string]any{"priority": 100, "stackable": true, "stackable_with_coupon": true})
	if err := db.Create(&promotionmodel.Campaign{ID: "promo-1", TenantUUID: "11111111-1111-1111-1111-111111111111", Code: "ORDER1000", Name: "order", PromotionType: promotionmodel.TypeAmountOff, ConditionRule: raw, ScopeRule: scope, ActionRule: action, StackingRule: stacking, ValidFrom: now.Add(-time.Hour), ValidTo: now.Add(time.Hour), Status: promotionmodel.StatusActive}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r.Group("/v1"), &app.Deps{DB: db})
	body, _ := json.Marshal(map[string]any{
		"channel": "miniapp", "currency": "CNY",
		"items": []map[string]any{{"line_id": "l1", "sku_id": "sku-1", "qty": 1, "unit_price_minor": 1000}},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/promotions/quote?tenant_uuid=11111111-1111-1111-1111-111111111111", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
}
