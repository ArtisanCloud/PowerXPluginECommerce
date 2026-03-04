package main

import (
	"testing"
	"time"
)

func TestResolveFrameworkGatewayConfig_BearerDefault(t *testing.T) {
	t.Setenv("PX_GATEWAY_BASE_URL", "http://127.0.0.1:8077")
	t.Setenv("PX_GATEWAY_API_PREFIX", "/api/v1")
	t.Setenv("PX_TOOL_TOKEN", "token-demo")
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "")
	t.Setenv("PX_GATEWAY_API_KEY", "")
	t.Setenv("PX_GATEWAY_TIMEOUT", "75")

	cfg := resolveFrameworkGatewayConfig()
	if cfg.BaseURL != "http://127.0.0.1:8077" {
		t.Fatalf("unexpected base url: %s", cfg.BaseURL)
	}
	if cfg.APIPrefix != "/api/v1" {
		t.Fatalf("unexpected api prefix: %s", cfg.APIPrefix)
	}
	if cfg.AuthScheme != "bearer" {
		t.Fatalf("unexpected auth scheme: %s", cfg.AuthScheme)
	}
	if cfg.ToolToken != "token-demo" {
		t.Fatalf("unexpected tool token: %s", cfg.ToolToken)
	}
	if cfg.Timeout != 75*time.Second {
		t.Fatalf("unexpected timeout: %s", cfg.Timeout)
	}
}

func TestResolveFrameworkGatewayConfig_ApiKey(t *testing.T) {
	t.Setenv("PX_GATEWAY_BASE_URL", "http://127.0.0.1:8077")
	t.Setenv("PX_GATEWAY_API_PREFIX", "/api")
	t.Setenv("PX_TOOL_TOKEN", "")
	t.Setenv("PX_PLUGIN_TOOL_TOKEN", "")
	t.Setenv("PX_GATEWAY_AUTH_SCHEME", "apikey")
	t.Setenv("PX_GATEWAY_API_KEY", "key-123")

	cfg := resolveFrameworkGatewayConfig()
	if cfg.AuthScheme != "apikey" {
		t.Fatalf("unexpected auth scheme: %s", cfg.AuthScheme)
	}
	if cfg.APIKey != "key-123" {
		t.Fatalf("unexpected api key: %s", cfg.APIKey)
	}
	if cfg.APIPrefix != "/api" {
		t.Fatalf("unexpected api prefix: %s", cfg.APIPrefix)
	}
}
