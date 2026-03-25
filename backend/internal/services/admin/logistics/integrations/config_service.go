package integrations

import (
	"encoding/json"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	"gorm.io/datatypes"
)

// ConfigService resolves provider and service-code mapping from carrier config.
type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) ResolveProvider(carrier *LogisticsModel.Carrier) string {
	if carrier == nil {
		return "self"
	}
	cfg := ParseConfigMap(carrier.Config)
	provider := strings.TrimSpace(strings.ToLower(anyString(cfg["provider"])))
	if provider == "" {
		provider = strings.TrimSpace(strings.ToLower(carrier.Type))
	}
	if provider == "" {
		return "self"
	}
	return provider
}

func (s *ConfigService) ResolveServiceCode(carrier *LogisticsModel.Carrier, serviceCode string) string {
	serviceCode = strings.TrimSpace(serviceCode)
	if carrier == nil {
		return serviceCode
	}
	cfg := ParseConfigMap(carrier.Config)
	rawMap, ok := cfg["service_code_map"]
	if !ok {
		return serviceCode
	}
	switch mapping := rawMap.(type) {
	case map[string]any:
		if val := strings.TrimSpace(anyString(mapping[serviceCode])); val != "" {
			return val
		}
	case map[string]string:
		if val := strings.TrimSpace(mapping[serviceCode]); val != "" {
			return val
		}
	}
	return serviceCode
}

func ParseConfigMap(config datatypes.JSON) map[string]any {
	if len(config) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(config, &out); err != nil {
		return map[string]any{}
	}
	if out == nil {
		return map[string]any{}
	}
	return out
}

func anyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}
