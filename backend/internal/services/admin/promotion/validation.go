package promotion

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	"gorm.io/datatypes"
)

func ValidateCampaignRules(promotionType string, validFrom, validTo time.Time, condition ConditionRule, scope ScopeRule, action ActionRule, stacking StackingRule) error {
	if validFrom.After(validTo) {
		return errors.New("valid_from must be before or equal to valid_to")
	}
	if condition.MinOrderAmountMinor < 0 {
		return errors.New("min_order_amount_minor must be greater than or equal to 0")
	}
	if strings.TrimSpace(scope.ScopeType) == "" {
		scope.ScopeType = "all"
	}
	switch strings.TrimSpace(scope.ScopeType) {
	case "all":
	case "sku":
		if len(nonEmptyStrings(scope.SKUIDs)) == 0 {
			return errors.New("sku_ids are required when scope_type is sku")
		}
	default:
		return errors.New("scope_type is invalid")
	}
	switch strings.TrimSpace(promotionType) {
	case promotionmodel.TypeAmountOff:
		if action.DiscountAmountMinor <= 0 {
			return errors.New("discount_amount_minor must be greater than 0")
		}
	case promotionmodel.TypePercentOff:
		if action.DiscountPercentBps <= 0 || action.DiscountPercentBps > 10000 {
			return errors.New("discount_percent_bps must be between 1 and 10000")
		}
		if action.MaxDiscountMinor < 0 {
			return errors.New("max_discount_minor must be greater than or equal to 0")
		}
	default:
		return errors.New("promotion_type is invalid")
	}
	if stacking.Priority < 0 {
		return errors.New("priority must be greater than or equal to 0")
	}
	return nil
}

func JSONFromValue(v any) (datatypes.JSON, error) {
	if v == nil {
		return datatypes.JSON([]byte("{}")), nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
}

func ValueFromJSON[T any](raw datatypes.JSON, fallback T) T {
	if len(raw) == 0 {
		return fallback
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return fallback
	}
	return out
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			out = append(out, v)
		}
	}
	return out
}
