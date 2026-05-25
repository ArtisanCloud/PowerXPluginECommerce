package bootstrap

import (
	"os"
	"strings"

	iamcontext "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/iam/context"
	iamcontracts "github.com/ArtisanCloud/PowerXPlugin/framework/backend/go/iam/contracts"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	iamservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/iam"
)

// IAMResolver determines whether the plugin should rely on delegated (PowerX Core)
// or local IAM. Priority: config.context.iam_mode > POWERX_PROXY.
type IAMResolver struct {
	mode   iamservice.IAMMode
	source string
}

func NewIAMResolver(cfg *config.Config) *IAMResolver {
	input := iamcontext.ResolveInput{
		ConfigMode:  resolveIAMModeInput(cfg),
		EnvMode:     strings.TrimSpace(os.Getenv("IAM_MODE")),
		Environment: strings.TrimSpace(os.Getenv("APP_ENV")),
	}
	if input.ConfigMode == "" && input.EnvMode == "" {
		input.PowerXProxy = strings.TrimSpace(os.Getenv("POWERX_PROXY"))
	}
	mode, record, err := (iamcontext.ModeResolver{}).Resolve(input)
	if err != nil {
		return &IAMResolver{mode: iamservice.IAMModeLocal, source: "framework:error"}
	}
	return &IAMResolver{mode: toServiceIAMMode(mode), source: record.Audit.Source}
}

func (r *IAMResolver) Mode() iamservice.IAMMode {
	if r == nil {
		return iamservice.IAMModeLocal
	}
	return r.mode
}

func (r *IAMResolver) Source() string {
	if r == nil {
		return "auto"
	}
	return r.source
}

func parseIAMMode(val string) (iamservice.IAMMode, bool) {
	v := strings.ToLower(strings.TrimSpace(val))
	switch v {
	case "delegated":
		return iamservice.IAMModeDelegated, true
	case "local":
		return iamservice.IAMModeLocal, true
	default:
		return iamservice.IAMMode(""), false
	}
}

func resolveIAMModeInput(cfg *config.Config) string {
	if cfg == nil || cfg.Context == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Context.IAMMode)
}

func toServiceIAMMode(mode iamcontracts.IAMMode) iamservice.IAMMode {
	switch mode {
	case iamcontracts.IAMModeDelegated:
		return iamservice.IAMModeDelegated
	default:
		return iamservice.IAMModeLocal
	}
}

func EffectiveHostMode(cfg *config.Config, iamMode string) bool {
	if strings.EqualFold(strings.TrimSpace(iamMode), string(iamservice.IAMModeDelegated)) {
		return true
	}
	if cfg != nil && cfg.Context != nil && strings.EqualFold(strings.TrimSpace(cfg.Context.IAMMode), string(iamservice.IAMModeDelegated)) {
		return true
	}
	return envTruthy("POWERX_PROXY")
}
