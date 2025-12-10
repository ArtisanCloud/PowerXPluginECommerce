package customer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	customermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/customer"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	httpmw "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateCustomerHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(newHandlerDeps(t))
	body := map[string]any{
		"name":           "API 客户",
		"type":           "individual",
		"phone":          "+8613800000088",
		"membershipTier": "gold",
		"source":         "website",
	}
	ctx, rec := newJSONRequest(t, http.MethodPost, "/customers", body)
	ctx.Set(httpmw.TenantUUIDContextKey, "tenant-handler")
	ctx.Request = ctx.Request.WithContext(authx.ContextWithTenantUUID(ctx.Request.Context(), "tenant-handler"))
	handler.CreateCustomer(ctx)
	require.Equal(t, http.StatusCreated, rec.Code)
	payload := parseResponse(t, rec.Body.Bytes())
	require.True(t, payload["success"].(bool))
}

func TestCreateCustomerHandlerValidation(t *testing.T) {
	handler := NewHandler(newHandlerDeps(t))
	ctx, rec := newJSONRequest(t, http.MethodPost, "/customers", map[string]any{})
	ctx.Set(httpmw.TenantUUIDContextKey, "tenant-handler")
	ctx.Request = ctx.Request.WithContext(authx.ContextWithTenantUUID(ctx.Request.Context(), "tenant-handler"))
	handler.CreateCustomer(ctx)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	payload := parseResponse(t, rec.Body.Bytes())
	require.False(t, payload["success"].(bool))
}

func TestDeleteCustomerHandlerRequiresReason(t *testing.T) {
	handler := NewHandler(newHandlerDeps(t))
	createCtx, createRec := newJSONRequest(t, http.MethodPost, "/customers", map[string]any{
		"name":           "待删除",
		"type":           "individual",
		"phone":          "+8613800000099",
		"membershipTier": "gold",
		"source":         "website",
	})
	createCtx.Set(httpmw.TenantUUIDContextKey, "tenant-handler")
	createCtx.Request = createCtx.Request.WithContext(authx.ContextWithTenantUUID(createCtx.Request.Context(), "tenant-handler"))
	handler.CreateCustomer(createCtx)
	var created map[string]any
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))
	data := created["data"].(map[string]any)
	id := data["id"].(string)
	ctx, rec := newJSONRequest(t, http.MethodDelete, "/customers/"+id, map[string]any{"reason": ""})
	ctx.Params = gin.Params{gin.Param{Key: "id", Value: id}}
	ctx.Set(httpmw.TenantUUIDContextKey, "tenant-handler")
	ctx.Request = ctx.Request.WithContext(authx.ContextWithTenantUUID(ctx.Request.Context(), "tenant-handler"))
	handler.DeleteCustomer(ctx)
	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func newJSONRequest(t *testing.T, method, path string, body map[string]any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	buf := bytes.NewBuffer(nil)
	if body != nil {
		require.NoError(t, json.NewEncoder(buf).Encode(body))
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req, err := http.NewRequest(method, path, buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

func parseResponse(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var payload map[string]any
	require.NoError(t, json.Unmarshal(raw, &payload))
	return payload
}

func newHandlerDeps(t *testing.T) *app.Deps {
	t.Helper()
	models.ForceSchemaForTests("")
	db, err := gorm.Open(sqlite.Open("file:customer_handler?mode=memory&cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&customermodel.Customer{}))
	return &app.Deps{DB: db}
}
