package promotion

type quoteItemRequest struct {
	LineID         string `json:"line_id,omitempty"`
	SKUID          string `json:"sku_id"`
	Qty            int64  `json:"qty"`
	UnitPriceMinor int64  `json:"unit_price_minor"`
}

type quoteRequest struct {
	UserID      string             `json:"user_id,omitempty"`
	Channel     string             `json:"channel"`
	Currency    string             `json:"currency"`
	SubmittedAt string             `json:"submitted_at,omitempty"`
	Items       []quoteItemRequest `json:"items"`
}
