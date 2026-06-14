package promotion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testTenantUUID = "11111111-1111-1111-1111-111111111111"

func TestCampaignHandlerContract_CreateAndList(t *testing.T) {
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:promotion_campaign_handler?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	setupAdminPromotionTables(t, db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		authx.SetTenantContext(c, authx.TenantContext{TenantUUID: testTenantUUID, UserID: 1001})
		c.Set("tenant_uuid", testTenantUUID)
		c.Next()
	})
	RegisterRoutes(r.Group("/admin"), &app.Deps{DB: db})

	now := time.Now().UTC()
	body := map[string]any{
		"code": "ORDER1000", "name": "满减", "promotion_type": promotionmodel.TypeAmountOff,
		"condition_rule": map[string]any{"min_order_amount_minor": 1000},
		"scope_rule":     map[string]any{"scope_type": "all"},
		"action_rule":    map[string]any{"discount_amount_minor": 100},
		"stacking_rule":  map[string]any{"priority": 100, "stackable": true, "stackable_with_coupon": true},
		"valid_from":     now.Add(-time.Hour).Format(time.RFC3339),
		"valid_to":       now.Add(time.Hour).Format(time.RFC3339),
		"save_action":    "draft",
	}
	createResp := performAdminPromotionRequest(t, r, http.MethodPost, "/admin/promotions", body)
	if createResp.Code != http.StatusOK {
		t.Fatalf("expected create 200, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	var created struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil || created.Data.Code != "ORDER1000" {
		t.Fatalf("decode create response: %v body=%s", err, createResp.Body.String())
	}

	listResp := performAdminPromotionRequest(t, r, http.MethodGet, "/admin/promotions?status=draft", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d body=%s", listResp.Code, listResp.Body.String())
	}
}

func performAdminPromotionRequest(t *testing.T, r http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	sep := "?"
	if strings.Contains(target, "?") {
		sep = "&"
	}
	req := httptest.NewRequest(method, target+sep+"tenant_uuid="+testTenantUUID, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	ctx := authx.ContextWithTenantContext(req.Context(), authx.TenantContext{TenantUUID: testTenantUUID, UserID: 1001})
	req = req.WithContext(ctx)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func setupAdminPromotionTables(t *testing.T, db *gorm.DB) {
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
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
}
