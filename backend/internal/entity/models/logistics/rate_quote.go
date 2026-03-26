package logistics

// RateQuoteInput describes simulator inputs for template quote calculation.
type RateQuoteInput struct {
	Region      string  `json:"region"`
	Weight      float64 `json:"weight,omitempty"`
	PieceCount  int     `json:"piece_count,omitempty"`
	Volume      float64 `json:"volume,omitempty"`
	OrderAmount float64 `json:"order_amount,omitempty"`
}

// RateQuoteZoneRule is the normalized zone rule for calculation.
type RateQuoteZoneRule struct {
	Region         string  `json:"region"`
	FirstMetric    float64 `json:"first_metric"`
	FirstFee       float64 `json:"first_fee"`
	AdditionalStep float64 `json:"additional_step"`
	AdditionalFee  float64 `json:"additional_fee"`
	FreeThreshold  float64 `json:"free_threshold"`
}

// RateQuoteResult is the simulator output for UI and API.
type RateQuoteResult struct {
	TemplateID  string            `json:"template_id"`
	Template    string            `json:"template"`
	Currency    string            `json:"currency"`
	BillingType string            `json:"billing_type"`
	MatchedZone RateQuoteZoneRule `json:"matched_zone"`
	FeeAmount   float64           `json:"fee_amount"`
	Breakdown   map[string]any    `json:"breakdown,omitempty"`
}
