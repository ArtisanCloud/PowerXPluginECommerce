package router

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

func TestPricingQueryRoutesExposeBothPrefixes(t *testing.T) {
	t.Setenv("POWERX_AUTH_OPTIONAL", "1")
	auditPath := filepath.Join(t.TempDir(), "audit.log")
	cfg := &config.Config{
		Server:       &config.ServerConfig{APIPrefix: "/api/v1", DevMode: true},
		Database:     &config.DatabaseConfig{Driver: "memory", DSN: "memory", Schema: "px_plugin_base"},
		Runtime:      &config.RuntimeConfig{},
		RuntimeOps:   &config.RuntimeOpsDefaults{},
		Context:      &config.ContextConfig{},
		CustomerAuth: &config.CustomerAuthConfig{},
		Security:     &config.SecurityConfig{ToolGrantSecret: "test-toolgrant-secret"},
		SecurityBaseline: &config.SecurityBaselineConfig{
			BaselineVersion: "test",
			ToolGrant:       config.ToolGrantBaselineConfig{TTLHours: 24},
			ConsentDefaults: config.ConsentDefaultsConfig{AuditChannel: auditPath},
		},
		Logging:      &config.LoggingConfig{},
		GRPCUpstream: &config.GRPCUpstream{},
		GRPCServer:   &config.GRPCServer{},
		Integration:  &config.IntegrationConfig{},
		Marketplace:  &config.MarketplaceConfig{},
		Operations:   &config.OperationsConfig{},
		AdminConsole: &config.AdminConsoleConfig{},
		TaskBus:      &config.TaskBusConfig{},
	}

	deps := &app.Deps{Ctx: context.Background(), Config: cfg, DB: &gorm.DB{}}
	r := NewRouter(cfg, deps)
	engine := r.Setup()
	if engine == nil {
		t.Fatal("expected router engine")
	}

	want := map[string]struct{}{
		"POST:/v1/pricing/query":     {},
		"POST:/api/v1/pricing/query": {},
	}
	handlerByRoute := map[string]string{}
	for _, route := range engine.Routes() {
		key := route.Method + ":" + route.Path
		if _, ok := want[key]; ok {
			handlerByRoute[key] = route.Handler
		}
	}
	for key := range want {
		if _, ok := handlerByRoute[key]; !ok {
			t.Fatalf("expected route %s registered", key)
		}
	}
	if handlerByRoute["POST:/v1/pricing/query"] != handlerByRoute["POST:/api/v1/pricing/query"] {
		t.Fatalf("expected both pricing query routes to share handler, got %q vs %q",
			handlerByRoute["POST:/v1/pricing/query"], handlerByRoute["POST:/api/v1/pricing/query"])
	}
}
