package promotion

import "testing"

func TestSelectApplicablePromotions(t *testing.T) {
	applied, rejected := selectApplicablePromotions([]quoteCandidate{
		{Applied: AppliedPromotion{PromotionID: "p1", Code: "A", DiscountMinor: 100}, Stacking: StackingRule{Priority: 20, Stackable: true, StackableWithCoupon: true, ExclusionGroup: "g"}},
		{Applied: AppliedPromotion{PromotionID: "p2", Code: "B", DiscountMinor: 200}, Stacking: StackingRule{Priority: 10, Stackable: false, StackableWithCoupon: true, ExclusionGroup: "g"}},
		{Applied: AppliedPromotion{PromotionID: "p3", Code: "C", DiscountMinor: 50}, Stacking: StackingRule{Priority: 30, Stackable: true, StackableWithCoupon: true}},
	})
	if len(applied) != 1 || applied[0].PromotionID != "p2" {
		t.Fatalf("expected p2 only applied, got applied=%+v rejected=%+v", applied, rejected)
	}
	if len(rejected) != 2 {
		t.Fatalf("expected 2 rejected promotions, got %+v", rejected)
	}
}

func TestSelectApplicablePromotionsWithMultipleExclusionGroups(t *testing.T) {
	applied, rejected := selectApplicablePromotions([]quoteCandidate{
		{Applied: AppliedPromotion{PromotionID: "p1", Code: "A", DiscountMinor: 100}, Stacking: StackingRule{Priority: 10, Stackable: true, StackableWithCoupon: true, ExclusionGroups: []string{"order_discount", "channel_campaign"}}},
		{Applied: AppliedPromotion{PromotionID: "p2", Code: "B", DiscountMinor: 180}, Stacking: StackingRule{Priority: 20, Stackable: true, StackableWithCoupon: true, ExclusionGroup: "channel_campaign"}},
		{Applied: AppliedPromotion{PromotionID: "p3", Code: "C", DiscountMinor: 80}, Stacking: StackingRule{Priority: 30, Stackable: true, StackableWithCoupon: true, ExclusionGroup: "amount_off_campaign"}},
	})
	if len(applied) != 2 {
		t.Fatalf("expected 2 applied promotions, got applied=%+v rejected=%+v", applied, rejected)
	}
	if applied[0].PromotionID != "p2" || applied[1].PromotionID != "p3" {
		t.Fatalf("expected p2 and p3 applied by priority, got %+v", applied)
	}
	if len(rejected) != 1 || rejected[0].PromotionID != "p1" {
		t.Fatalf("expected p1 rejected by exclusive conflict, got %+v", rejected)
	}
}
