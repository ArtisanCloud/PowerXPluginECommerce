package integration

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseInvokePolicy_DefaultTrue(t *testing.T) {
	p := parseInvokePolicy([]byte(`{"capabilityId":"x"}`))
	if !p.AuthRequired || !p.TenantScoped {
		t.Fatalf("unexpected default policy: %+v", p)
	}
}

func TestParseInvokePolicy_Override(t *testing.T) {
	p := parseInvokePolicy([]byte(`{"auth_required":false,"tenant_scoped":false}`))
	if p.AuthRequired || p.TenantScoped {
		t.Fatalf("unexpected override policy: %+v", p)
	}
}

func TestResolveInvokeAuthContext_AuthRequiredMissingToken(t *testing.T) {
	c := newInvokeTestContext("")
	_, err := resolveInvokeAuthContext(c, "tenant-1", invokePolicy{AuthRequired: true, TenantScoped: true})
	if err == nil || err.Code != "GW_POLICY_AUTH_REQUIRED" {
		t.Fatalf("expected GW_POLICY_AUTH_REQUIRED, got %+v", err)
	}
}

func TestResolveInvokeAuthContext_TenantScopedMissingTid(t *testing.T) {
	c := newInvokeTestContext("Bearer not-a-jwt")
	_, err := resolveInvokeAuthContext(c, "tenant-1", invokePolicy{AuthRequired: true, TenantScoped: true})
	if err == nil || err.Code != "GW_POLICY_TOKEN_TENANT_MISSING" {
		t.Fatalf("expected GW_POLICY_TOKEN_TENANT_MISSING, got %+v", err)
	}
}

func TestResolveInvokeAuthContext_RejectZeroTenant(t *testing.T) {
	token := buildTestJWT(map[string]any{"tid": zeroTenantUUID})
	c := newInvokeTestContext("Bearer " + token)
	_, err := resolveInvokeAuthContext(c, zeroTenantUUID, invokePolicy{AuthRequired: true, TenantScoped: true})
	if err == nil || err.Code != "GW_POLICY_ZERO_TENANT" {
		t.Fatalf("expected GW_POLICY_ZERO_TENANT, got %+v", err)
	}
}

func TestResolveInvokeAuthContext_RejectTenantMismatch(t *testing.T) {
	token := buildTestJWT(map[string]any{"tid": "tenant-a"})
	c := newInvokeTestContext("Bearer " + token)
	_, err := resolveInvokeAuthContext(c, "tenant-b", invokePolicy{AuthRequired: true, TenantScoped: true})
	if err == nil || err.Code != "GW_POLICY_TENANT_MISMATCH" {
		t.Fatalf("expected GW_POLICY_TENANT_MISMATCH, got %+v", err)
	}
}

func TestResolveInvokeAuthContext_AnonymousNoToken(t *testing.T) {
	c := newInvokeTestContext("Bearer should-not-be-used")
	ctx, err := resolveInvokeAuthContext(c, "", invokePolicy{AuthRequired: false, TenantScoped: false})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if ctx.Token != "" || ctx.TokenSource != "none" {
		t.Fatalf("expected anonymous auth context, got %+v", ctx)
	}
}

func newInvokeTestContext(auth string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/api/v1/integration/capabilities/invoke", nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	c.Request = req
	return c
}

func buildTestJWT(claims map[string]any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	bodyJSON, _ := json.Marshal(claims)
	body := base64.RawURLEncoding.EncodeToString(bodyJSON)
	return header + "." + body + ".sig"
}
