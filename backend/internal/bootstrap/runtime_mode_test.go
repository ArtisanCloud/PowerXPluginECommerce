package bootstrap

import (
	"os"
	"testing"
)

func TestResolveToolTokenPriority(t *testing.T) {
	t.Setenv("PX_TOOL_TOKEN", "tool")
	t.Setenv("PX_PLUGIN_TOOL_TOKEN", "plugin")
	t.Setenv("POWERX_AUTH_TOKEN", "legacy")

	tok, src := ResolveToolToken()
	if tok != "tool" || src != "env:PX_TOOL_TOKEN" {
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
	t.Setenv("POWERX_RBAC_DELEGATE", "0")
	t.Setenv("PX_TOOL_TOKEN", "eyJhbGciOiJub25lIn0.eyJ0aWQiOiJ0ZW5hbnQtYWJjIn0.")

	d := ResolveRuntimeModeDecision(nil, "local", "config")
	if !d.PowerXProxy {
		t.Fatal("expected powerx proxy true")
	}
	if d.IAMMode != "local" {
		t.Fatalf("unexpected iam mode: %s", d.IAMMode)
	}
	if d.WSRoute != "host" || d.CapabilityRoute != "host" {
		t.Fatalf("unexpected routes: ws=%s cap=%s", d.WSRoute, d.CapabilityRoute)
	}
	if d.OutboundTokenSource != "env:PX_TOOL_TOKEN" {
		t.Fatalf("unexpected token source: %s", d.OutboundTokenSource)
	}
	if d.TokenTenantID != "tenant-abc" {
		t.Fatalf("unexpected tid: %s", d.TokenTenantID)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
