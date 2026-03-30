package main

import (
	"testing"
	"time"
)

func TestResolveFrameworkGatewayConfig_BasicFields(t *testing.T) {
	t.Setenv("PX_GATEWAY_BASE_URL", "http://127.0.0.1:8077")
	t.Setenv("PX_TOOL_TOKEN", "token-demo")
	t.Setenv("PX_TENANT_UUID", "tenant-001")
	t.Setenv("PX_GATEWAY_GRPC_TARGET", "127.0.0.1:50051")
	t.Setenv("PX_GATEWAY_TIMEOUT", "75")
	t.Setenv("PX_GATEWAY_USER_AGENT", "plugin-ecommerce/ci")
	t.Setenv("PX_GATEWAY_CONTRACT_VERSION", "v2")

	cfg := resolveFrameworkGatewayConfig()
	if cfg.BaseURL != "http://127.0.0.1:8077" {
		t.Fatalf("unexpected base url: %s", cfg.BaseURL)
	}
	if cfg.ToolToken != "token-demo" {
		t.Fatalf("unexpected tool token: %s", cfg.ToolToken)
	}
	if cfg.TenantID != "tenant-001" {
		t.Fatalf("unexpected tenant id: %s", cfg.TenantID)
	}
	if cfg.GRPCTarget != "127.0.0.1:50051" {
		t.Fatalf("unexpected grpc target: %s", cfg.GRPCTarget)
	}
	if cfg.Timeout != 75*time.Second {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
	if cfg.UserAgent != "plugin-ecommerce/ci" {
		t.Fatalf("unexpected user agent: %s", cfg.UserAgent)
	}
	if cfg.ContractVersion != "v2" {
		t.Fatalf("unexpected contract version: %s", cfg.ContractVersion)
	}
}

func TestResolveFrameworkGatewayConfig_BaseURLFallback(t *testing.T) {
	t.Setenv("PX_GATEWAY_BASE_URL", "")
	t.Setenv("POWERX_CORE_ENDPOINT", "https://gateway.example.com/")
	t.Setenv("PX_GATEWAY_TIMEOUT", "15")

	cfg := resolveFrameworkGatewayConfig()
	if cfg.BaseURL != "https://gateway.example.com" {
		t.Fatalf("unexpected base url: %s", cfg.BaseURL)
	}
	if cfg.Timeout != 15*time.Second {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
}

func TestResolveGatewayAPIPrefix(t *testing.T) {
	t.Setenv("PX_GATEWAY_API_PREFIX", "")
	if got := resolveGatewayAPIPrefix(); got != "/api/v1" {
		t.Fatalf("unexpected default prefix: %s", got)
	}

	t.Setenv("PX_GATEWAY_API_PREFIX", "api")
	if got := resolveGatewayAPIPrefix(); got != "/api" {
		t.Fatalf("unexpected normalized prefix: %s", got)
	}

	t.Setenv("PX_GATEWAY_API_PREFIX", "/")
	if got := resolveGatewayAPIPrefix(); got != "" {
		t.Fatalf("unexpected root prefix normalization: %s", got)
	}
}

func TestNormalizeGatewayAuthScheme(t *testing.T) {
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "")
	if got := normalizeGatewayAuthScheme("", "token-demo", ""); got != "bearer" {
		t.Fatalf("unexpected bearer default: %s", got)
	}

	if got := normalizeGatewayAuthScheme("apikey", "", ""); got != "apikey" {
		t.Fatalf("unexpected explicit apikey normalization: %s", got)
	}

	if got := normalizeGatewayAuthScheme("", "", "key-123"); got != "apikey" {
		t.Fatalf("unexpected api key fallback normalization: %s", got)
	}
}
