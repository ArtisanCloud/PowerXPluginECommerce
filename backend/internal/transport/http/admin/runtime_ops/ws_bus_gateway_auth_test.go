package runtime_ops

import (
	"net/http/httptest"
	"testing"
	"time"

	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/iam"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

func TestResolveWSBusGatewayAuth_LocalUsesToolToken(t *testing.T) {
	t.Setenv("PX_TOOL_TOKEN", "eyJhbGciOiJub25lIn0.eyJ0aWQiOiJ0ZW5hbnQtdDEifQ.")
	t.Setenv("PX_GATEWAY_BASE_URL", "http://localhost:8077/api/v1")
	t.Setenv("PX_GATEWAY_API_PREFIX", "/api/v1")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)

	decision := resolveWSBusGatewayAuth(c, &app.Deps{IAMMode: authx.IAMModeLocal})
	if decision.Source != "env:PX_TOOL_TOKEN" {
		t.Fatalf("unexpected source: %s", decision.Source)
	}
	if decision.TenantID != "tenant-t1" {
		t.Fatalf("unexpected tid: %s", decision.TenantID)
	}
	if decision.GatewayToken == "" {
		t.Fatal("expected gateway token")
	}
	if decision.GatewayBaseURL != "http://localhost:8077" {
		t.Fatalf("unexpected base url: %s", decision.GatewayBaseURL)
	}
	if decision.GatewayAPIPrefix != "/api/v1" {
		t.Fatalf("unexpected api prefix: %s", decision.GatewayAPIPrefix)
	}
	if decision.GatewayAuthScheme != "bearer" {
		t.Fatalf("unexpected auth scheme: %s", decision.GatewayAuthScheme)
	}
}

func TestResolveWSBusGatewayAuth_DelegatedUsesToolTokenForOutbound(t *testing.T) {
	inbound := "eyJhbGciOiJub25lIn0.eyJ0aWQiOiJ0ZW5hbnQtaW5ib3VuZCJ9."
	tool := "eyJhbGciOiJub25lIn0.eyJ0aWQiOiJ0ZW5hbnQtdG9vbCJ9."
	t.Setenv("PX_TOOL_TOKEN", tool)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("Authorization", "Bearer "+inbound)
	c.Request = req

	decision := resolveWSBusGatewayAuth(c, &app.Deps{IAMMode: authx.IAMModeDelegated})
	if decision.Source != "env:PX_TOOL_TOKEN" {
		t.Fatalf("unexpected source: %s", decision.Source)
	}
	if decision.TenantID != "tenant-tool" {
		t.Fatalf("unexpected tid: %s", decision.TenantID)
	}
	if decision.Authorization != "Bearer "+tool {
		t.Fatalf("unexpected authorization: %s", decision.Authorization)
	}
}

func TestResolveWSBusGatewayAuth_ApiKey(t *testing.T) {
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "apikey")
	t.Setenv("PX_GATEWAY_API_KEY", "k_test_123")
	t.Setenv("PX_GATEWAY_BASE_URL", "http://localhost:8077")
	t.Setenv("PX_GATEWAY_API_PREFIX", "/api")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)

	decision := resolveWSBusGatewayAuth(c, &app.Deps{IAMMode: authx.IAMModeLocal})
	if decision.GatewayAuthScheme != "apikey" {
		t.Fatalf("unexpected auth scheme: %s", decision.GatewayAuthScheme)
	}
	if decision.Authorization != "ApiKey k_test_123" {
		t.Fatalf("unexpected authorization: %s", decision.Authorization)
	}
	if decision.Source != "env:PX_GATEWAY_API_KEY" {
		t.Fatalf("unexpected source: %s", decision.Source)
	}
	if decision.GatewayAPIPrefix != "/api" {
		t.Fatalf("unexpected api prefix: %s", decision.GatewayAPIPrefix)
	}
}

func TestResolveWSBusGatewayAuth_DefaultsToApiKeyWhenAPIKeyPresent(t *testing.T) {
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "")
	t.Setenv("PX_GATEWAY_API_KEY", "k_test_123")
	t.Setenv("PX_TOOL_TOKEN", "eyJhbGciOiJub25lIn0.eyJ0aWQiOiJ0ZW5hbnQtdDEifQ.")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)

	decision := resolveWSBusGatewayAuth(c, &app.Deps{IAMMode: authx.IAMModeLocal})
	if decision.GatewayAuthScheme != "apikey" {
		t.Fatalf("unexpected auth scheme: %s", decision.GatewayAuthScheme)
	}
	if decision.Authorization != "ApiKey k_test_123" {
		t.Fatalf("unexpected authorization: %s", decision.Authorization)
	}
}

func TestResolveWSBusGatewayAuth_Timeout(t *testing.T) {
	t.Setenv("PX_GATEWAY_TIMEOUT", "75")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", nil)

	decision := resolveWSBusGatewayAuth(c, &app.Deps{IAMMode: authx.IAMModeLocal})
	if decision.GatewayTimeout != 75*time.Second {
		t.Fatalf("unexpected timeout: %s", decision.GatewayTimeout)
	}
}
