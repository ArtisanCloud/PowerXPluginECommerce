package bootstrap

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
)

// RuntimeModeDecision captures startup runtime semantics for provider / host routing.
type RuntimeModeDecision struct {
	ProviderInput       string
	ProviderMode        string
	ProviderSource      string
	PowerXProxy         bool
	EffectiveProxy      bool
	CapabilityRoute     string
	WSRoute             string
	OutboundTokenSource string
	GatewayReady        bool
	TokenTenantID       string
}

func ResolveRuntimeModeDecision(cfg *config.Config, providerMode string, providerSource string) RuntimeModeDecision {
	token, tokenSource := ResolvePowerXAccessTokenBootstrap(cfg, EffectiveHostMode())
	tenantID := resolveSTSTenantID(cfg)
	if tenantID == "" {
		tenantID, _ = ParseTenantIDFromJWT(token)
	}
	proxy := envTruthy("POWERX_PROXY")
	effectiveProxy := EffectiveHostMode()
	capRoute := "local"
	wsRoute := "local"
	if effectiveProxy {
		capRoute = "host"
		wsRoute = "host"
	}
	return RuntimeModeDecision{
		ProviderInput:       resolveProviderInput(cfg),
		ProviderMode:        strings.ToLower(strings.TrimSpace(providerMode)),
		ProviderSource:      strings.TrimSpace(providerSource),
		PowerXProxy:         proxy,
		EffectiveProxy:      effectiveProxy,
		CapabilityRoute:     capRoute,
		WSRoute:             wsRoute,
		OutboundTokenSource: tokenSource,
		GatewayReady:        PowerXSTSConfigured(cfg) || strings.TrimSpace(token) != "",
		TokenTenantID:       tenantID,
	}
}

func resolveProviderInput(cfg *config.Config) string {
	if cfg == nil || cfg.Context == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Context.ProviderMode)
}

func ResolvePowerXAccessTokenBootstrap(cfg *config.Config, hostMode bool) (token string, source string) {
	if hostMode {
		if PowerXSTSConfigured(cfg) {
			return "", "sts:configured"
		}
		return "", "sts:missing"
	}
	return ResolveDebugToken()
}

// PowerXSTSConfigured reports whether the plugin can exchange a short-lived
// powerx:api access token for outbound PowerX calls.
func PowerXSTSConfigured(cfg *config.Config) bool {
	if cfg == nil || cfg.GRPCUpstream == nil {
		return false
	}
	return strings.TrimSpace(cfg.GRPCUpstream.STSClientID) != "" &&
		strings.TrimSpace(cfg.GRPCUpstream.STSClientSecret) != ""
}

// ResolveDebugToken returns a local-only debug token. Host mode must use STS.
func ResolveDebugToken() (token string, source string) {
	if v := strings.TrimSpace(os.Getenv("POWERX_GRPC_UPSTREAM_TOKEN")); v != "" {
		return v, "env:POWERX_GRPC_UPSTREAM_TOKEN"
	}
	if v := strings.TrimSpace(os.Getenv("POWERX_AUTH_TOKEN")); v != "" {
		return v, "env:POWERX_AUTH_TOKEN"
	}
	return "", ""
}

func resolveSTSTenantID(cfg *config.Config) string {
	if cfg == nil || cfg.GRPCUpstream == nil {
		return ""
	}
	return strings.TrimSpace(cfg.GRPCUpstream.TenantUUID)
}

// ParseTenantIDFromJWT extracts tid claim from JWT payload without signature validation.
func ParseTenantIDFromJWT(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return "", false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}
	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", false
	}
	tid := strings.TrimSpace(toString(claims["tid"]))
	if tid == "" {
		return "", false
	}
	return tid, true
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimSuffix(formatFloat(x), "0"), "."))
	default:
		return ""
	}
}

func formatFloat(v float64) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func envTruthy(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
