package router

import (
	"testing"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
)

func TestBuildJWTInProxyMode(t *testing.T) {
	cfg := &config.Config{
		Server: &config.ServerConfig{},
	}
	r := &Router{cfg: cfg}

	t.Setenv("POWERX_PROXY", "1")
	t.Setenv("POWERX_SECURITY_JWT_ISSUER", "powerx-auth")
	t.Setenv("POWERX_SECURITY_JWT_AUDIENCE", "plugin:com.powerx.plugins.base")
	t.Setenv("POWERX_SECURITY_JWT_SECRET", "secret")
	t.Setenv("POWERX_SECURITY_CTX_HMAC_SECRET", "ctx-secret")

	jwtCfg := r.buildJWT()

	if jwtCfg.Optional {
		t.Fatal("expected strict JWT validation when running in PowerX proxy")
	}
	if jwtCfg.AllowSignedContext {
		t.Fatal("expected signed context disabled in proxy mode")
	}
	if jwtCfg.Issuer != "powerx-auth" {
		t.Fatalf("unexpected issuer, got %s", jwtCfg.Issuer)
	}
}

func TestBuildJWTInDelegatedModeWithoutProxyEnv(t *testing.T) {
	cfg := &config.Config{
		Context: &config.ContextConfig{IAMMode: "delegated"},
		Server:  &config.ServerConfig{},
	}
	r := &Router{cfg: cfg}

	t.Setenv("POWERX_PROXY", "0")
	t.Setenv("POWERX_SECURITY_JWT_ISSUER", "powerx-auth")
	t.Setenv("POWERX_SECURITY_JWT_AUDIENCE", "plugin:com.powerx.plugins.base")
	t.Setenv("POWERX_SECURITY_JWT_SECRET", "secret")
	t.Setenv("POWERX_SECURITY_CTX_HMAC_SECRET", "ctx-secret")

	jwtCfg := r.buildJWT()

	if jwtCfg.Optional {
		t.Fatal("expected strict JWT validation when IAM mode is delegated")
	}
	if jwtCfg.AllowSignedContext {
		t.Fatal("expected signed context disabled in delegated mode")
	}
}

func TestBuildJWTInDevModeStrictByDefault(t *testing.T) {
	cfg := &config.Config{
		Server: &config.ServerConfig{DevMode: true},
	}
	r := &Router{cfg: cfg}

	t.Setenv("POWERX_PROXY", "0")
	t.Setenv("POWERX_SECURITY_JWT_ISSUER", "")
	t.Setenv("POWERX_SECURITY_JWT_AUDIENCE", "")
	t.Setenv("POWERX_SECURITY_JWT_SECRET", "")

	jwtCfg := r.buildJWT()
	if jwtCfg.Optional {
		t.Fatal("expected strict JWT when running locally in dev mode unless POWERX_AUTH_OPTIONAL enabled")
	}
	if jwtCfg.AllowSignedContext {
		t.Fatal("expected signed context disabled for local dev by default")
	}
}

func TestBuildJWTHonorsOptionalEnvVar(t *testing.T) {
	cfg := &config.Config{
		Server: &config.ServerConfig{DevMode: true},
	}
	r := &Router{cfg: cfg}

	t.Setenv("POWERX_PROXY", "0")
	t.Setenv("POWERX_SECURITY_JWT_ISSUER", "")
	t.Setenv("POWERX_SECURITY_JWT_AUDIENCE", "")
	t.Setenv("POWERX_SECURITY_JWT_SECRET", "")
	t.Setenv("POWERX_AUTH_OPTIONAL", "1")

	jwtCfg := r.buildJWT()
	if !jwtCfg.Optional {
		t.Fatal("expected optional JWT when POWERX_AUTH_OPTIONAL=1")
	}
	if jwtCfg.AllowSignedContext {
		t.Fatal("expected signed context disabled for local dev by default")
	}
}
