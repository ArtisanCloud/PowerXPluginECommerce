package coupon

type quoteItemRequest struct {
	LineID         string `json:"line_id,omitempty"`
	SKUID          string `json:"sku_id"`
	Qty            int64  `json:"qty"`
	UnitPriceMinor int64  `json:"unit_price_minor"`
}

type quoteRequest struct {
	UserID    string             `json:"user_id"`
	Channel   string             `json:"channel"`
	CouponIDs []string           `json:"coupon_ids,omitempty"`
	Currency  string             `json:"currency,omitempty"`
	Items     []quoteItemRequest `json:"items"`
}
