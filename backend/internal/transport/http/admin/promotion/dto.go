package promotion

import (
	"time"

	promotionsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
)

type campaignCreateRequest struct {
	Code          string                     `json:"code"`
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	PromotionType string                     `json:"promotion_type"`
	ConditionRule promotionsvc.ConditionRule `json:"condition_rule"`
	ScopeRule     promotionsvc.ScopeRule     `json:"scope_rule"`
	ActionRule    promotionsvc.ActionRule    `json:"action_rule"`
	StackingRule  promotionsvc.StackingRule  `json:"stacking_rule"`
	ValidFrom     time.Time                  `json:"valid_from"`
	ValidTo       time.Time                  `json:"valid_to"`
	SaveAction    string                     `json:"save_action"`
}

type campaignUpdateRequest struct {
	Name          *string                     `json:"name"`
	Description   *string                     `json:"description"`
	ConditionRule *promotionsvc.ConditionRule `json:"condition_rule"`
	ScopeRule     *promotionsvc.ScopeRule     `json:"scope_rule"`
	ActionRule    *promotionsvc.ActionRule    `json:"action_rule"`
	StackingRule  *promotionsvc.StackingRule  `json:"stacking_rule"`
	ValidFrom     *time.Time                  `json:"valid_from"`
	ValidTo       *time.Time                  `json:"valid_to"`
}
