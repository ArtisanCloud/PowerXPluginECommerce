package http

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	admincoupon "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/coupon"
	adminpromotion "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/transport/http/admin/promotion"
	"github.com/gin-gonic/gin"
)

func TestMarketplaceRBACEntriesContainListings(t *testing.T) {
	entries := marketplacePublicRBACEntries("/api/v1")
	perm, ok := entries["GET:/api/v1/marketplace/listings"]
	if !ok {
		t.Fatalf("expected listings GET entry in RBAC map, got %+v", entries)
	}
	if perm.Resource != "marketplace.listings" || perm.Action != "read" {
		t.Fatalf("unexpected permission %+v", perm)
	}
}

func TestRegistryIncludeCouponAdminRBAC(t *testing.T) {
	r := &Registry{rbac: map[string]Permission{}}
	r.mergeRBAC(couponRBACEntriesForTest("/api/v1"))
	perm, ok := r.rbac["GET:/api/v1/admin/coupons/templates"]
	if !ok {
		t.Fatalf("expected coupon template rbac entry in registry, got %+v", r.rbac)
	}
	if perm.Resource != "com.powerx.plugins.ecommerce:pricing.coupon.template" || perm.Action != "read" {
		t.Fatalf("unexpected permission %+v", perm)
	}
}

func TestPromotionOpenAPIAndRoutingMatrixMatchRegisteredRoutes(t *testing.T) {
	specRoot := filepath.Join("..", "..", "..", "..", "specs", "015-pricing-promotions-engine", "contracts")
	openapiRaw, err := os.ReadFile(filepath.Join(specRoot, "promotions.openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	matrixRaw, err := os.ReadFile(filepath.Join(specRoot, "routing-matrix.md"))
	if err != nil {
		t.Fatalf("read routing matrix: %v", err)
	}
	openapi := string(openapiRaw)
	matrix := string(matrixRaw)

	expectedOpenAPIPaths := []string{
		"/admin/promotions:",
		"/admin/promotions/{id}:",
		"/admin/promotions/{id}/activate:",
		"/admin/promotions/{id}/pause:",
		"/admin/promotions/{id}/clone:",
		"/admin/promotions/{id}/audit-logs:",
		"/v1/promotions/quote:",
	}
	for _, path := range expectedOpenAPIPaths {
		if !strings.Contains(openapi, path) {
			t.Fatalf("openapi missing promotion path %s", path)
		}
	}

	expectedMatrixRows := []string{
		"`GET` | `/api/v1/admin/promotions`",
		"`POST` | `/api/v1/admin/promotions`",
		"`PATCH` | `/api/v1/admin/promotions/:id`",
		"`POST` | `/api/v1/admin/promotions/:id/activate`",
		"`POST` | `/api/v1/admin/promotions/:id/pause`",
		"`POST` | `/api/v1/admin/promotions/:id/clone`",
		"`GET` | `/api/v1/admin/promotions/:id/audit-logs`",
	}
	for _, row := range expectedMatrixRows {
		if !strings.Contains(matrix, row) {
			t.Fatalf("routing matrix missing row fragment %s", row)
		}
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	adminpromotion.RegisterRoutes(engine.Group("/api/v1/admin"), &app.Deps{})
	routes := map[string]bool{}
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		"GET /api/v1/admin/promotions",
		"POST /api/v1/admin/promotions",
		"PATCH /api/v1/admin/promotions/:id",
		"POST /api/v1/admin/promotions/:id/activate",
		"POST /api/v1/admin/promotions/:id/pause",
		"POST /api/v1/admin/promotions/:id/clone",
		"GET /api/v1/admin/promotions/:id/audit-logs",
	} {
		if !routes[route] {
			t.Fatalf("registered routes missing %s", route)
		}
	}

	rbac := StaticRBACEntries("/api/v1")
	for _, key := range []string{
		"GET:/api/v1/admin/promotions",
		"POST:/api/v1/admin/promotions",
		"PATCH:/api/v1/admin/promotions/:id",
		"POST:/api/v1/admin/promotions/:id/activate",
		"POST:/api/v1/admin/promotions/:id/pause",
		"POST:/api/v1/admin/promotions/:id/clone",
		"GET:/api/v1/admin/promotions/:id/audit-logs",
	} {
		if _, ok := rbac[key]; !ok {
			t.Fatalf("rbac missing %s", key)
		}
	}
}

type Permission = authx.Permission

func couponRBACEntriesForTest(prefix string) map[string]Permission {
	return admincoupon.RBACEntries(prefix)
}
