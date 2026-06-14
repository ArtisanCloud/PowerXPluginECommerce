package promotion

import (
	"time"
)

type ConditionRule struct {
	MinOrderAmountMinor int64 `json:"min_order_amount_minor,omitempty"`
}

type ScopeRule struct {
	ScopeType string   `json:"scope_type"`
	SKUIDs    []string `json:"sku_ids,omitempty"`
	Channels  []string `json:"channels,omitempty"`
}

type ActionRule struct {
	DiscountAmountMinor int64 `json:"discount_amount_minor,omitempty"`
	DiscountPercentBps  int   `json:"discount_percent_bps,omitempty"`
	MaxDiscountMinor    int64 `json:"max_discount_minor,omitempty"`
}

type StackingRule struct {
	Priority            int      `json:"priority"`
	Stackable           bool     `json:"stackable"`
	StackableWithCoupon bool     `json:"stackable_with_coupon"`
	ExclusionGroup      string   `json:"exclusion_group,omitempty"`
	ExclusionGroups     []string `json:"exclusion_groups,omitempty"`
}

type QuoteItemInput struct {
	LineID         string `json:"line_id,omitempty"`
	SKUID          string `json:"sku_id"`
	Qty            int64  `json:"qty"`
	UnitPriceMinor int64  `json:"unit_price_minor"`
}

type QuoteInput struct {
	TenantUUID  string
	UserID      string
	Channel     string
	Currency    string
	Items       []QuoteItemInput
	SubmittedAt time.Time
	OrderID     string
}

type AppliedPromotion struct {
	PromotionID         string `json:"promotion_id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	PromotionType       string `json:"promotion_type"`
	DiscountMinor       int64  `json:"discount_minor"`
	Priority            int    `json:"priority"`
	ExclusionGroup      string `json:"exclusion_group,omitempty"`
	StackableWithCoupon bool   `json:"stackable_with_coupon"`
}

type RejectedPromotion struct {
	PromotionID string `json:"promotion_id,omitempty"`
	Code        string `json:"code,omitempty"`
	Reason      string `json:"reason"`
}

type LineAllocation struct {
	LineID                    string `json:"line_id,omitempty"`
	SKUID                     string `json:"sku_id"`
	BaseAmountMinor           int64  `json:"base_amount_minor"`
	PromotionDiscountMinor    int64  `json:"promotion_discount_minor"`
	AfterPromotionAmountMinor int64  `json:"after_promotion_amount_minor"`
}

type QuoteResult struct {
	Currency                 string              `json:"currency"`
	BaseTotalMinor           int64               `json:"base_total_minor"`
	PromotionDiscountMinor   int64               `json:"promotion_discount_minor"`
	AfterPromotionTotalMinor int64               `json:"after_promotion_total_minor"`
	CouponStackingAllowed    bool                `json:"coupon_stacking_allowed"`
	AppliedPromotions        []AppliedPromotion  `json:"applied_promotions"`
	RejectedPromotions       []RejectedPromotion `json:"rejected_promotions"`
	LineAllocations          []LineAllocation    `json:"line_allocations"`
	PricedAt                 time.Time           `json:"priced_at"`
}
