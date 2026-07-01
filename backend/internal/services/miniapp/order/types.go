package order

import "time"

type CreateOrderItemInput struct {
	SKUID string `json:"skuId"`
	Qty   int64  `json:"qty"`
}

type CreateOrderRequest struct {
	Channel           string                 `json:"channel"`
	ShippingAddressID string                 `json:"shippingAddressId,omitempty"`
	ShippingAddress   *ShippingAddress       `json:"shippingAddress,omitempty"`
	Items             []CreateOrderItemInput `json:"items"`
	Locale            string                 `json:"locale,omitempty"`
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
	Status                  string               `json:"status"`
	Amounts                 MoneyDTO             `json:"amounts"`
	Promotion               *PromotionSummaryDTO `json:"promotion,omitempty"`
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
