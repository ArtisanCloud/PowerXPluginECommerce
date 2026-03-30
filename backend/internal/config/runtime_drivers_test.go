package config

import "testing"

func TestResolveRuntimeDrivers_DefaultLocalWithoutProxy(t *testing.T) {
	t.Setenv("POWERX_PROXY", "0")
	cfg := &Config{Runtime: &RuntimeConfig{Drivers: &RuntimeDriversConfig{}}}

	if got := cfg.ResolveWebSocketDriver(); got != RuntimeDriverLocal {
		t.Fatalf("unexpected websocket driver: %s", got)
	}
	if got := cfg.ResolveEventTopicDriver(); got != RuntimeDriverLocal {
		t.Fatalf("unexpected event topic driver: %s", got)
	}
	if got := cfg.ResolveTaskDriver(); got != RuntimeDriverLocal {
		t.Fatalf("unexpected task driver: %s", got)
	}
	if got := cfg.ResolveCacheDriver(); got != RuntimeDriverLocal {
		t.Fatalf("unexpected cache driver: %s", got)
	}
}

func TestResolveRuntimeDrivers_DefaultFrameworkWithProxy(t *testing.T) {
	t.Setenv("POWERX_PROXY", "1")
	cfg := &Config{Runtime: &RuntimeConfig{Drivers: &RuntimeDriversConfig{}}}

	if got := cfg.ResolveWebSocketDriver(); got != RuntimeDriverFramework {
		t.Fatalf("unexpected websocket driver: %s", got)
	}
	if got := cfg.ResolveEventTopicDriver(); got != RuntimeDriverFramework {
		t.Fatalf("unexpected event topic driver: %s", got)
	}
	if got := cfg.ResolveTaskDriver(); got != RuntimeDriverFramework {
		t.Fatalf("unexpected task driver: %s", got)
	}
	if got := cfg.ResolveCacheDriver(); got != RuntimeDriverFramework {
		t.Fatalf("unexpected cache driver: %s", got)
	}
}

func TestResolveRuntimeDrivers_EnvOverride(t *testing.T) {
	t.Setenv("POWERX_PROXY", "0")
	t.Setenv("POWERX_RUNTIME_DRIVER_TASK", "framework")
	cfg := &Config{Runtime: &RuntimeConfig{Drivers: &RuntimeDriversConfig{Task: "local"}}}

	if got := cfg.ResolveTaskDriver(); got != RuntimeDriverFramework {
		t.Fatalf("unexpected task driver: %s", got)
	}
}
