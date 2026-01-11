package pricing

import "time"

type PricingQueryRequest struct {
	SKUID           string     `json:"sku_id" binding:"required"`
	Currency        string     `json:"currency" binding:"required"`
	ChannelID       *string    `json:"channel_id,omitempty"`
	CustomerGroupID *string    `json:"customer_group_id,omitempty"`
	SupplierID      *string    `json:"supplier_id,omitempty"`
	AsOf            *time.Time `json:"as_of,omitempty"`
}

type Money struct {
	AmountMinor int64  `json:"amount_minor"`
	Amount      string `json:"amount"`
}

type PricingMatched struct {
	PricebookID   string `json:"pricebook_id"`
	PricebookCode string `json:"pricebook_code"`
	VersionID     string `json:"version_id"`
	Version       int    `json:"version"`
	SKUID         string `json:"sku_id"`
}

type PricingTrace struct {
	Fallback      string  `json:"fallback"`
	NoPriceReason *string `json:"no_price_reason,omitempty"`
}

type PricingQueryResponse struct {
	Currency    string          `json:"currency"`
	Priced      bool            `json:"priced"`
	SourceField *string         `json:"source_field,omitempty"`
	Price       *Money          `json:"price,omitempty"`
	Matched     *PricingMatched `json:"matched,omitempty"`
	Fields      map[string]any  `json:"fields,omitempty"`
	Trace       PricingTrace    `json:"trace"`
}
