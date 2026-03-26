package logistics

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"

	LogisticsModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/logistics"
	LogisticsRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/logistics"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type RateQuoteService struct {
	tplRepo *LogisticsRepo.RateTemplateRepository
}

func NewRateQuoteService(deps *app.Deps) *RateQuoteService {
	if deps == nil || deps.DB == nil {
		return &RateQuoteService{}
	}
	return &RateQuoteService{
		tplRepo: LogisticsRepo.NewRateTemplateRepository(deps.DB),
	}
}

type QuoteRateRequest struct {
	TemplateID  string  `json:"template_id"`
	Region      string  `json:"region"`
	Weight      float64 `json:"weight,omitempty"`
	PieceCount  int     `json:"piece_count,omitempty"`
	Volume      float64 `json:"volume,omitempty"`
	OrderAmount float64 `json:"order_amount,omitempty"`
}

func (s *RateQuoteService) Quote(ctx context.Context, tenantUUID string, req QuoteRateRequest) (*LogisticsModel.RateQuoteResult, error) {
	if s == nil || s.tplRepo == nil {
		return nil, errors.New("rate quote service unavailable")
	}
	ctx = withTenantContext(ctx, tenantUUID)
	if strings.TrimSpace(req.TemplateID) == "" {
		return nil, errors.New("template_id required")
	}
	tpl, err := s.tplRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, err
	}
	rules := decodeJSONMap(tpl.Rules)
	billingType := normalizeBillingType(stringValue(rules, "billing", "billing_type"))
	zone := s.matchZoneRule(rules, strings.TrimSpace(req.Region), billingType)
	metricValue := metricByType(billingType, req.Weight, req.PieceCount, req.Volume)
	fee := calculateQuoteFee(zone, metricValue, req.OrderAmount)
	return &LogisticsModel.RateQuoteResult{
		TemplateID:  tpl.ID,
		Template:    tpl.Name,
		Currency:    defaultString(tpl.Currency, "CNY"),
		BillingType: billingType,
		MatchedZone: zone,
		FeeAmount:   fee,
		Breakdown: map[string]any{
			"input_metric": metricValue,
			"region":       strings.TrimSpace(req.Region),
		},
	}, nil
}

func (s *RateQuoteService) matchZoneRule(rules map[string]any, region, billingType string) LogisticsModel.RateQuoteZoneRule {
	region = strings.TrimSpace(region)
	zones := extractZoneRules(rules, billingType)
	if len(zones) == 0 {
		return LogisticsModel.RateQuoteZoneRule{Region: defaultString(region, "default")}
	}
	fallback := zones[0]
	for _, zone := range zones {
		if strings.EqualFold(strings.TrimSpace(zone.Region), region) {
			return zone
		}
		if strings.TrimSpace(zone.Region) == "*" || strings.EqualFold(strings.TrimSpace(zone.Region), "default") {
			fallback = zone
		}
	}
	return fallback
}

func extractZoneRules(rules map[string]any, billingType string) []LogisticsModel.RateQuoteZoneRule {
	rawZones, ok := rules["zones"]
	if !ok {
		rawZones, ok = rules["zone_list"]
	}
	zones := make([]LogisticsModel.RateQuoteZoneRule, 0)
	if ok {
		for _, raw := range castArray(rawZones) {
			zone := parseZoneRule(raw, billingType)
			if zone.Region != "" {
				zones = append(zones, zone)
			}
		}
	}
	if len(zones) > 0 {
		return zones
	}
	defaultZone := parseZoneRule(rules["defaultZone"], billingType)
	if defaultZone.Region == "" {
		defaultZone = parseZoneRule(rules["default_zone"], billingType)
	}
	if defaultZone.Region == "" {
		defaultZone.Region = "default"
	}
	return []LogisticsModel.RateQuoteZoneRule{defaultZone}
}

func parseZoneRule(raw any, billingType string) LogisticsModel.RateQuoteZoneRule {
	m := castMap(raw)
	return LogisticsModel.RateQuoteZoneRule{
		Region:         defaultString(stringValue(m, "region", "name"), ""),
		FirstMetric:    floatValue(m, "first_weight", "firstWeight", "first_piece", "firstPiece", "first_volume", "firstVolume"),
		FirstFee:       floatValue(m, "first_fee", "firstFee"),
		AdditionalStep: floatValue(m, "additional_weight", "additionalWeight", "additional_piece", "additionalPiece", "additional_volume", "additionalVolume", "add_weight", "addWeight"),
		AdditionalFee:  floatValue(m, "additional_fee", "additionalFee", "add_fee", "addFee"),
		FreeThreshold:  floatValue(m, "free_threshold", "freeThreshold"),
	}
}

func calculateQuoteFee(zone LogisticsModel.RateQuoteZoneRule, metricValue, orderAmount float64) float64 {
	if zone.FreeThreshold > 0 && orderAmount >= zone.FreeThreshold {
		return 0
	}
	if zone.FirstFee <= 0 {
		return 0
	}
	if zone.FirstMetric <= 0 || metricValue <= zone.FirstMetric || zone.AdditionalStep <= 0 || zone.AdditionalFee <= 0 {
		return round2(zone.FirstFee)
	}
	extraUnits := math.Ceil((metricValue - zone.FirstMetric) / zone.AdditionalStep)
	return round2(zone.FirstFee + extraUnits*zone.AdditionalFee)
}

func metricByType(billingType string, weight float64, pieceCount int, volume float64) float64 {
	switch normalizeBillingType(billingType) {
	case "piece":
		if pieceCount <= 0 {
			return 0
		}
		return float64(pieceCount)
	case "volume":
		if volume < 0 {
			return 0
		}
		return volume
	default:
		if weight < 0 {
			return 0
		}
		return weight
	}
}

func normalizeBillingType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "piece":
		return "piece"
	case "volume":
		return "volume"
	default:
		return "weight"
	}
}

func decodeJSONMap(raw []byte) map[string]any {
	var out map[string]any
	if len(raw) == 0 {
		return map[string]any{}
	}
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		out = map[string]any{}
	}
	return out
}

func castMap(raw any) map[string]any {
	if raw == nil {
		return map[string]any{}
	}
	if m, ok := raw.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func castArray(raw any) []any {
	if raw == nil {
		return []any{}
	}
	if arr, ok := raw.([]any); ok {
		return arr
	}
	return []any{}
}

func stringValue(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			if s, ok := value.(string); ok {
				if strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
	}
	return ""
}

func floatValue(m map[string]any, keys ...string) float64 {
	for _, key := range keys {
		value, ok := m[key]
		if !ok || value == nil {
			continue
		}
		switch v := value.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		}
	}
	return 0
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
