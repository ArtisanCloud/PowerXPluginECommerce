package bootstrap

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
)

// RuntimeModeDecision captures startup runtime semantics for IAM / host routing.
type RuntimeModeDecision struct {
	IAMInput            string
	IAMMode             string
	IAMSource           string
	PowerXProxy         bool
	CapabilityRoute     string
	WSRoute             string
	OutboundTokenSource string
	GatewayReady        bool
	TokenTenantID       string
}

func ResolveRuntimeModeDecision(cfg *config.Config, iamMode string, iamSource string) RuntimeModeDecision {
	token, tokenSource := ResolveToolToken()
	tenantID, _ := ParseTenantIDFromJWT(token)
	proxy := envTruthy("POWERX_PROXY")
	capRoute := "local"
	wsRoute := "local"
	if proxy {
		capRoute = "host"
		wsRoute = "host"
	}
	return RuntimeModeDecision{
		IAMInput:            resolveIAMInput(cfg),
		IAMMode:             strings.ToLower(strings.TrimSpace(iamMode)),
		IAMSource:           strings.TrimSpace(iamSource),
		PowerXProxy:         proxy,
		CapabilityRoute:     capRoute,
		WSRoute:             wsRoute,
		OutboundTokenSource: tokenSource,
		GatewayReady:        strings.TrimSpace(token) != "",
		TokenTenantID:       tenantID,
	}
}

func resolveIAMInput(cfg *config.Config) string {
	if cfg == nil || cfg.Context == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Context.IAMMode)
}

// ResolveToolToken returns outbound tool token by new precedence.
// Priority: PX_TOOL_TOKEN > PX_PLUGIN_TOOL_TOKEN > POWERX_AUTH_TOKEN.
func ResolveToolToken() (token string, source string) {
	if v := strings.TrimSpace(os.Getenv("PX_TOOL_TOKEN")); v != "" {
		return v, "env:PX_TOOL_TOKEN"
	}
	if v := strings.TrimSpace(os.Getenv("PX_PLUGIN_TOOL_TOKEN")); v != "" {
		return v, "env:PX_PLUGIN_TOOL_TOKEN"
	}
	if v := strings.TrimSpace(os.Getenv("POWERX_AUTH_TOKEN")); v != "" {
		return v, "env:POWERX_AUTH_TOKEN"
	}
	return "", ""
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
