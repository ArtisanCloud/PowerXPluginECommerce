package coupon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQuoteHandlerContract(t *testing.T) {
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_quote_handler?mode=memory&cache=shared"), &gorm.Config{})
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
	emptyJSON, _ := json.Marshal(map[string]any{})
	thresholdPercent, _ := json.Marshal(map[string]any{"percent": 10.0})
	scopeItem, _ := json.Marshal(map[string]any{"level": "item", "sku_ids": []string{"sku-1"}})
	stackItem, _ := json.Marshal(map[string]any{"priority": 10, "stackable": true, "level": "item"})
	if err := db.Create(&couponmodel.CouponTemplate{
		ID: "tpl-item", TenantUUID: "11111111-1111-1111-1111-111111111111", Code: "ITEM10", Name: "item10", CouponType: "percent",
		ThresholdRule: thresholdPercent, ScopeRule: scopeItem, StackingRule: stackItem, RefundRule: emptyJSON,
		ValidFrom: now.Add(-time.Hour), ValidTo: now.Add(time.Hour), Status: "active",
	}).Error; err != nil {
		t.Fatalf("seed template: %v", err)
	}
	if err := db.Create(&couponmodel.CouponAsset{ID: "asset-1", TenantUUID: "11111111-1111-1111-1111-111111111111", TemplateID: "tpl-item", UserID: "user-1", CouponCode: "C1", Status: "available"}).Error; err != nil {
		t.Fatalf("seed asset: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/v1")
	RegisterRoutes(v1, &app.Deps{DB: db})

	payload := map[string]any{
		"user_id":    "user-1",
		"channel":    "official",
		"coupon_ids": []string{"asset-1"},
		"currency":   "CNY",
		"items": []map[string]any{
			{"line_id": "l1", "sku_id": "sku-1", "qty": 1, "unit_price_minor": 1000},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/coupons/quote?tenant_uuid=11111111-1111-1111-1111-111111111111", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", resp.Code, resp.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if ok, _ := out["success"].(bool); !ok {
		t.Fatalf("expected success response: %v", out)
	}
}
