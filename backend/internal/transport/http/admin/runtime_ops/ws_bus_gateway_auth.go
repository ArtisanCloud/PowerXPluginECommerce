package runtime_ops

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/gin-gonic/gin"
)

type wsBusGatewayAuthDecision struct {
	Authorization     string
	Source            string
	TenantID          string
	ProxyEnabled      bool
	ProviderMode      string
	GatewayBaseURL    string
	GatewayAPIPrefix  string
	GatewayTimeout    time.Duration
	GatewayAuthScheme string
	GatewayToken      string
	GatewayAPIKey     string
}

func resolveWSBusGatewayAuth(c *gin.Context, deps *app.Deps) wsBusGatewayAuthDecision {
	scheme := resolveGatewayAuthScheme()
	gatewayAPIKey := strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY"))
	decision := wsBusGatewayAuthDecision{
		Source:            "none",
		ProxyEnabled:      envTruthyValue(os.Getenv("POWERX_PROXY")),
		ProviderMode:      providerModeLabel(deps),
		GatewayBaseURL:    resolveGatewayBaseURL(),
		GatewayAPIPrefix:  resolveGatewayAPIPrefix(),
		GatewayTimeout:    resolveGatewayTimeout(),
		GatewayAuthScheme: scheme,
		GatewayAPIKey:     gatewayAPIKey,
	}

	if scheme == "apikey" {
		if gatewayAPIKey != "" {
			decision.Authorization = "ApiKey " + gatewayAPIKey
			decision.Source = "env:PX_GATEWAY_API_KEY"
		}
		return decision
	}

	if deps != nil {
		if token, err := deps.PowerXAccessToken(c.Request.Context()); err == nil && strings.TrimSpace(token) != "" {
			decision.Authorization = "Bearer " + strings.TrimSpace(token)
			decision.Source = "sts:exchange"
			decision.GatewayToken = strings.TrimSpace(token)
			if deps.Config != nil && deps.Config.GRPCUpstream != nil {
				decision.TenantID = strings.TrimSpace(deps.Config.GRPCUpstream.TenantUUID)
			}
		}
	}
	return decision
}

func normalizeAuthScheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "apikey", "api_key", "api-key":
		return "apikey"
	case "bearer":
		return "bearer"
	default:
		return ""
	}
}

func resolveGatewayAuthScheme() string {
	explicit := normalizeAuthScheme(os.Getenv("PX_GATEWAY_AUTH_SCHEME"))
	if explicit != "" {
		return explicit
	}
	if strings.TrimSpace(os.Getenv("PX_GATEWAY_API_KEY")) != "" {
		return "apikey"
	}
	return "bearer"
}

func resolveGatewayAPIPrefix() string {
	raw := strings.TrimSpace(os.Getenv("PX_GATEWAY_API_PREFIX"))
	if raw == "" {
		raw = "/api/v1"
	}
	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	raw = "/" + strings.Trim(strings.TrimSpace(raw), "/")
	if raw == "/" {
		return ""
	}
	return raw
}

func resolveGatewayTimeout() time.Duration {
	const defaultTimeout = 60 * time.Second
	raw := strings.TrimSpace(os.Getenv("PX_GATEWAY_TIMEOUT"))
	if raw == "" {
		return defaultTimeout
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return defaultTimeout
}

func resolveGatewayBaseURL() string {
	base := strings.TrimSpace(os.Getenv("PX_GATEWAY_BASE_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("POWERX_CORE_ENDPOINT"))
	}
	if base == "" {
		base = strings.TrimSpace(config.GetString("POWERX_CORE_ENDPOINT", "http://localhost:8077"))
	}
	base = strings.TrimRight(base, "/")
	if base == "" {
		base = "http://localhost:8077"
	}
	// 迁移期兼容：如果历史配置把前缀写在 base_url 上，自动剥离到 api_prefix。
	apiPrefix := resolveGatewayAPIPrefix()
	if apiPrefix != "" && strings.HasSuffix(base, apiPrefix) {
		base = strings.TrimSuffix(base, apiPrefix)
		base = strings.TrimRight(base, "/")
	}
	if base == "" {
		base = "http://localhost:8077"
	}
	return base
}

func buildGatewayEndpoint(baseURL, apiPrefix, path string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	prefix := strings.TrimSpace(apiPrefix)
	route := "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	if prefix == "" {
		return base + route
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = "/" + strings.Trim(prefix, "/")
	return base + prefix + route
}

func (d wsBusGatewayAuthDecision) hasGatewayCredential() bool {
	return strings.TrimSpace(d.Authorization) != ""
}

func gatewayAuthSource(auth wsBusGatewayAuthDecision) string {
	if strings.TrimSpace(auth.Source) == "" {
		return "none"
	}
	return auth.Source
}

func envTruthyValue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func providerModeLabel(deps *app.Deps) string {
	if deps == nil {
		return "local"
	}
	mode := strings.TrimSpace(deps.ProviderMode.String())
	if mode == "" {
		return "local"
	}
	return mode
}
