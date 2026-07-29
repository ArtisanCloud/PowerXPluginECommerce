package bootstrap

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	iamservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/iam"
)

// ProviderResolver maps the unified provider mode to the existing auth branch.
type ProviderResolver struct {
	mode   iamservice.Mode
	source string
}

func NewProviderResolver(cfg *config.Config) (*ProviderResolver, error) {
	mode, err := cfg.ResolveProviderMode()
	if err != nil {
		return nil, err
	}
	return &ProviderResolver{mode: toServiceProviderMode(mode), source: resolveProviderModeSource(cfg)}, nil
}

func (r *ProviderResolver) Mode() iamservice.Mode {
	if r == nil {
		return iamservice.ModeLocal
	}
	return r.mode
}

func (r *ProviderResolver) Source() string {
	if r == nil {
		return ""
	}
	return r.source
}

func resolveProviderModeInput(cfg *config.Config) string {
	if cfg == nil || cfg.Context == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Context.ProviderMode)
}

func toServiceProviderMode(mode string) iamservice.Mode {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case config.ProviderModeDelegated:
		return iamservice.ModeDelegated
	default:
		return iamservice.ModeLocal
	}
}

func EffectiveHostMode() bool {
	return envTruthy("POWERX_PROXY")
}

func resolveProviderModeSource(cfg *config.Config) string {
	if cfg != nil && cfg.Context != nil && strings.TrimSpace(cfg.Context.ProviderMode) != "" {
		return "config:context.provider_mode"
	}
	return "env:POWERX_PROVIDER_MODE"
}
