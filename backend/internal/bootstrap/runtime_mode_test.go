package bootstrap

import (
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"os"
	"testing"
)

func TestResolveDebugTokenUsesRootDebugToken(t *testing.T) {
	t.Setenv("POWERX_AUTH_TOKEN", "legacy")

	tok, src := ResolveDebugToken()
	if tok != "legacy" || src != "env:POWERX_AUTH_TOKEN" {
		t.Fatalf("unexpected token resolution: tok=%s src=%s", tok, src)
	}
}

func TestParseTenantIDFromJWT(t *testing.T) {
	jwt := "eyJhbGciOiJub25lIn0.eyJ0aWQiOiJ0ZW5hbnQtMTIzIn0."
	tid, ok := ParseTenantIDFromJWT(jwt)
	if !ok {
		t.Fatal("expected tid from token")
	}
	if tid != "tenant-123" {
		t.Fatalf("unexpected tid: %s", tid)
	}
}

func TestResolveRuntimeModeDecision(t *testing.T) {
	t.Setenv("POWERX_PROXY", "1")

	cfg := &config.Config{GRPCUpstream: &config.GRPCUpstream{
		TenantUUID:      "tenant-abc",
		STSClientID:     "com.powerx.plugins.ecommerce.tenant-abc",
		STSClientSecret: "secret",
		STSAudience:     "powerx:api",
		STSScope:        "access",
	}}
	d := ResolveRuntimeModeDecision(cfg, "local", "config")
	if !d.PowerXProxy {
		t.Fatal("expected powerx proxy true")
	}
	if !d.EffectiveProxy {
		t.Fatal("expected effective proxy true")
	}
	if d.IAMMode != "local" {
		t.Fatalf("unexpected iam mode: %s", d.IAMMode)
	}
	if d.WSRoute != "host" || d.CapabilityRoute != "host" {
		t.Fatalf("unexpected routes: ws=%s cap=%s", d.WSRoute, d.CapabilityRoute)
	}
	if d.OutboundTokenSource != "sts:configured" {
		t.Fatalf("unexpected token source: %s", d.OutboundTokenSource)
	}
	if d.TokenTenantID != "tenant-abc" {
		t.Fatalf("unexpected tid: %s", d.TokenTenantID)
	}
}

func TestEffectiveHostMode(t *testing.T) {
	t.Setenv("POWERX_PROXY", "0")
	if EffectiveHostMode(&config.Config{Context: &config.ContextConfig{IAMMode: "delegated"}}, "") != true {
		t.Fatal("delegated IAM mode must force host mode")
	}

	t.Setenv("POWERX_PROXY", "1")
	if EffectiveHostMode(&config.Config{Context: &config.ContextConfig{IAMMode: "local"}}, "local") != true {
		t.Fatal("local + POWERX_PROXY=1 must use host route")
	}

	t.Setenv("POWERX_PROXY", "0")
	if EffectiveHostMode(&config.Config{Context: &config.ContextConfig{IAMMode: "local"}}, "local") != false {
		t.Fatal("local + POWERX_PROXY=0 must stay local")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
