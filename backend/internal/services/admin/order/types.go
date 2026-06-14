package order

import "time"

type CreateOrderItemInput struct {
	SKUID string `json:"skuId"`
	Qty   int64  `json:"qty"`
}

type CreateOrderRequest struct {
	CustomerID        string                 `json:"customerId"`
	Channel           string                 `json:"channel"`
	CouponIDs         []string               `json:"couponIds,omitempty"`
	ShippingAddressID string                 `json:"shippingAddressId,omitempty"`
	ShippingAddress   *ShippingAddress       `json:"shippingAddress,omitempty"`
	Items             []CreateOrderItemInput `json:"items"`
	Note              string                 `json:"note,omitempty"`
}

type ShippingAddress struct {
	Label          string         `json:"label,omitempty"`
	RecipientName  string         `json:"recipientName"`
	RecipientPhone string         `json:"recipientPhone"`
	CountryCode    string         `json:"countryCode,omitempty"`
	Province       string         `json:"province,omitempty"`
	City           string         `json:"city,omitempty"`
	District       string         `json:"district,omitempty"`
	Address1       string         `json:"address1"`
	Address2       string         `json:"address2,omitempty"`
	PostalCode     string         `json:"postalCode,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type MoneyDTO struct {
	Currency string `json:"currency"`
	Subtotal int64  `json:"subtotal"`
	Total    int64  `json:"total"`
}

type OrderSummaryDTO struct {
	OrderID                 string               `json:"orderId"`
	OrderNo                 string               `json:"orderNo"`
	CustomerID              string               `json:"customerId"`
	Channel                 string               `json:"channel"`
	CreatedByType           string               `json:"createdByType"`
	Status                  string               `json:"status"`
	Amounts                 MoneyDTO             `json:"amounts"`
	Promotion               *PromotionSummaryDTO `json:"promotion,omitempty"`
	Coupon                  *CouponSummaryDTO    `json:"coupon,omitempty"`
	ShippingAddressSnapshot *ShippingAddress     `json:"shippingAddressSnapshot,omitempty"`
	CreatedAt               time.Time            `json:"createdAt"`
}

type OrderListResponse struct {
	Items    []OrderSummaryDTO `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"pageSize"`
	Total    int64             `json:"total"`
}

type OrderItemDTO struct {
	SKUID       string `json:"skuId"`
	Qty         int64  `json:"qty"`
	UnitPrice   int64  `json:"unitPrice"`
	LineAmount  int64  `json:"lineAmount"`
	PriceSource string `json:"priceSource,omitempty"`
}

type OrderEventDTO struct {
	EventType    string    `json:"eventType"`
	OperatorType string    `json:"operatorType"`
	Operator     string    `json:"operator,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type OrderDetailDTO struct {
	Summary OrderSummaryDTO `json:"summary"`
	Items   []OrderItemDTO  `json:"items"`
	Events  []OrderEventDTO `json:"events"`
}

type CouponLineAllocationDTO struct {
	LineID        string `json:"line_id"`
	BaseMinor     int64  `json:"base_minor"`
	DiscountMinor int64  `json:"discount_minor"`
}

type PromotionLineAllocationDTO struct {
	LineID                    string `json:"line_id,omitempty"`
	SKUID                     string `json:"sku_id"`
	BaseAmountMinor           int64  `json:"base_amount_minor"`
	PromotionDiscountMinor    int64  `json:"promotion_discount_minor"`
	AfterPromotionAmountMinor int64  `json:"after_promotion_amount_minor"`
}

type PromotionAppliedDTO struct {
	PromotionID         string `json:"promotion_id"`
	Code                string `json:"code"`
	Name                string `json:"name"`
	PromotionType       string `json:"promotion_type"`
	DiscountMinor       int64  `json:"discount_minor"`
	Priority            int    `json:"priority"`
	ExclusionGroup      string `json:"exclusion_group,omitempty"`
	StackableWithCoupon bool   `json:"stackable_with_coupon"`
}

type PromotionRejectedDTO struct {
	PromotionID string `json:"promotion_id,omitempty"`
	Code        string `json:"code,omitempty"`
	Reason      string `json:"reason"`
}

type PromotionSummaryDTO struct {
	Currency                 string                       `json:"currency"`
	BaseTotalMinor           int64                        `json:"base_total_minor"`
	PromotionDiscountMinor   int64                        `json:"promotion_discount_minor"`
	AfterPromotionTotalMinor int64                        `json:"after_promotion_total_minor"`
	CouponStackingAllowed    bool                         `json:"coupon_stacking_allowed"`
	LineAllocations          []PromotionLineAllocationDTO `json:"line_allocations"`
	AppliedPromotions        []PromotionAppliedDTO        `json:"applied_promotions"`
	RejectedPromotions       []PromotionRejectedDTO       `json:"rejected_promotions"`
	PricedAt                 time.Time                    `json:"priced_at"`
}

type CouponAppliedDTO struct {
	AssetID       string `json:"asset_id"`
	TemplateID    string `json:"template_id"`
	CouponCode    string `json:"coupon_code"`
	DiscountMinor int64  `json:"discount_minor"`
	Level         string `json:"level"`
}

type CouponRejectedDTO struct {
	AssetID string `json:"asset_id"`
	Reason  string `json:"reason"`
}

type CouponSummaryDTO struct {
	Currency           string                    `json:"currency"`
	BaseTotalMinor     int64                     `json:"base_total_minor"`
	DiscountTotalMinor int64                     `json:"discount_total_minor"`
	PayableTotalMinor  int64                     `json:"payable_total_minor"`
	LineAllocations    []CouponLineAllocationDTO `json:"line_allocations"`
	AppliedCoupons     []CouponAppliedDTO        `json:"applied_coupons"`
	RejectedCoupons    []CouponRejectedDTO       `json:"rejected_coupons"`
	PricedAt           time.Time                 `json:"priced_at"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason,omitempty"`
}

type UpdateOrderShippingAddressRequest struct {
	ShippingAddress *ShippingAddress `json:"shippingAddress"`
}

type BenefitReviewCreateRequest struct {
	BenefitType     string  `json:"benefitType"`
	BenefitCode     string  `json:"benefitCode"`
	ValueType       string  `json:"valueType"`
	Value           float64 `json:"value"`
	Currency        string  `json:"currency,omitempty"`
	StackingAllowed bool    `json:"stackingAllowed"`
	Note            string  `json:"note,omitempty"`
}

type BenefitReviewDecisionRequest struct {
	ReviewIDs []uint64 `json:"reviewIds"`
	Reason    string   `json:"reason,omitempty"`
}

type OrderBenefitReviewDTO struct {
	ID              uint64     `json:"id"`
	OrderID         string     `json:"order_id"`
	OrderNo         string     `json:"order_no"`
	BenefitType     string     `json:"benefit_type"`
	BenefitCode     string     `json:"benefit_code"`
	ValueType       string     `json:"value_type"`
	Value           int64      `json:"value"`
	AmountMinor     int64      `json:"amount_minor"`
	Currency        string     `json:"currency"`
	StackingAllowed bool       `json:"stacking_allowed"`
	Status          string     `json:"status"`
	SubmittedBy     string     `json:"submitted_by"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ReviewedBy      string     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewReason    string     `json:"review_reason,omitempty"`
	Note            string     `json:"note,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
