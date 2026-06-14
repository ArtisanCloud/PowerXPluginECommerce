package promotion

import (
	"testing"
	"time"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
)

func TestValidateCampaignRules(t *testing.T) {
	now := time.Now().UTC()
	err := ValidateCampaignRules(
		promotionmodel.TypeAmountOff,
		now,
		now.Add(time.Hour),
		ConditionRule{MinOrderAmountMinor: 1000},
		ScopeRule{ScopeType: "sku", SKUIDs: []string{"sku-1"}},
		ActionRule{DiscountAmountMinor: 100},
		StackingRule{Priority: 100, Stackable: true, StackableWithCoupon: true},
	)
	if err != nil {
		t.Fatalf("expected valid rules, got %v", err)
	}
	if err := ValidateCampaignRules(promotionmodel.TypeAmountOff, now, now.Add(time.Hour), ConditionRule{}, ScopeRule{ScopeType: "sku"}, ActionRule{DiscountAmountMinor: 100}, StackingRule{}); err == nil {
		t.Fatalf("expected sku scope validation error")
	}
	if err := ValidateCampaignRules(promotionmodel.TypePercentOff, now, now.Add(time.Hour), ConditionRule{}, ScopeRule{ScopeType: "all"}, ActionRule{DiscountPercentBps: 10001}, StackingRule{}); err == nil {
		t.Fatalf("expected percent validation error")
	}
}
