package coupon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTemplateHandler_CreateListUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:coupon_template_handler?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	setupCouponTemplateHandlerTables(t, db)

	handler := NewTemplateHandler(coupon.NewTemplateService(&app.Deps{DB: db}))
	tenantUUID := "11111111-1111-1111-1111-111111111111"

	now := time.Now().UTC()
	createBody := map[string]any{
		"code":        "PROMO-NEW",
		"name":        "新客券",
		"coupon_type": "fixed",
		"valid_from":  now.Format(time.RFC3339),
		"valid_to":    now.Add(24 * time.Hour).Format(time.RFC3339),
		"status":      "active",
	}
	createData, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/admin/coupons/templates", bytes.NewReader(createData))
	createReq.Header.Set("Content-Type", "application/json")
	createResp := httptest.NewRecorder()
	createCtx := gin.CreateTestContextOnly(createResp, gin.New())
	createCtx.Request = createReq
	createCtx.Set(httpmw.TenantUUIDContextKey, tenantUUID)
	authx.SetTenantContext(createCtx, authx.TenantContext{TenantUUID: tenantUUID, UserID: 1001})
	handler.Create(createCtx)
	require.Equal(t, http.StatusOK, createResp.Code)

	var createdPayload map[string]any
	require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &createdPayload))
	require.True(t, createdPayload["success"].(bool))

	var templateID string
	dataObj := createdPayload["data"].(map[string]any)
	if v, ok := dataObj["id"].(string); ok {
		templateID = v
	}
	require.NotEmpty(t, templateID)

	listReq := httptest.NewRequest(http.MethodGet, "/admin/coupons/templates?keyword=PROMO", nil)
	listResp := httptest.NewRecorder()
	listCtx := gin.CreateTestContextOnly(listResp, gin.New())
	listCtx.Request = listReq
	listCtx.Set(httpmw.TenantUUIDContextKey, tenantUUID)
	authx.SetTenantContext(listCtx, authx.TenantContext{TenantUUID: tenantUUID, UserID: 1001})
	handler.List(listCtx)
	require.Equal(t, http.StatusOK, listResp.Code)

	updateBody := map[string]any{
		"name": "新客券-改",
	}
	updateData, _ := json.Marshal(updateBody)
	updateReq := httptest.NewRequest(http.MethodPatch, "/admin/coupons/templates/"+templateID, bytes.NewReader(updateData))
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp := httptest.NewRecorder()
	updateCtx := gin.CreateTestContextOnly(updateResp, gin.New())
	updateCtx.Request = updateReq
	updateCtx.Params = gin.Params{{Key: "id", Value: templateID}}
	updateCtx.Set(httpmw.TenantUUIDContextKey, tenantUUID)
	authx.SetTenantContext(updateCtx, authx.TenantContext{TenantUUID: tenantUUID, UserID: 1001})
	handler.Update(updateCtx)
	require.Equal(t, http.StatusOK, updateResp.Code, updateResp.Body.String())
}

func setupCouponTemplateHandlerTables(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS coupon_templates (
		id TEXT PRIMARY KEY,
		tenant_uuid TEXT NOT NULL,
		code TEXT NOT NULL,
		name TEXT NOT NULL,
		coupon_type TEXT NOT NULL,
		threshold_rule TEXT,
		scope_rule TEXT,
		stacking_rule TEXT,
		refund_rule TEXT,
		valid_from DATETIME,
		valid_to DATETIME,
		status TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error)
}
