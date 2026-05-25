package bootstrap

import (
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/config"
	customerauth "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/customer/auth"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// BuildCustomerAuthenticator 根据配置构造客户鉴权器。
func BuildCustomerAuthenticator(cfg *config.Config, deps *app.Deps) (customerauth.Authenticator, customerauth.LocalAuthService, error) {
	mode := cfg.ResolveCustomerAuthMode()
	customerCfg := cfg.CustomerAuthConfigOrDefault()
	switch mode {
	case config.CustomerAuthModeDelegate:
		token := customerCfg.ServiceToken
		if strings.TrimSpace(token) == "" {
			token = strings.TrimSpace(config.GetString("POWERX_AUTH_TOKEN", ""))
		}
		authenticator, err := customerauth.NewDelegateAuthenticator(customerauth.DelegateOptions{
			Endpoint:     customerCfg.DelegateEndpoint,
			ServiceToken: token,
			CacheTTL:     customerCfg.CacheTTL,
		})
		return authenticator, nil, err
	default:
		local, err := customerauth.NewLocalService(deps.DB, customerauth.LocalConfig{
			JWTSecret: []byte(customerCfg.JWTSecret),
			Issuer:    fallback(customerCfg.JWTIssuer, "powerx-plugin-customer"),
			Audience:  fallback(customerCfg.JWTAudience, "mini-app"),
			TTL:       customerCfg.JWTExpires,
		})
		if err != nil {
			return nil, nil, err
		}
		return local, local, nil
	}
}

func fallback(value, def string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return def
	}
	return value
}
