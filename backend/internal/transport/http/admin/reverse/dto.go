package reverse

type createWaybillRequest struct {
	OrderID     string         `json:"order_id"`
	AfterSaleID string         `json:"after_sale_id"`
	WaybillNo   string         `json:"waybill_no,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type appendTrackingRequest struct {
	Status      string         `json:"status"`
	Description string         `json:"description,omitempty"`
	OccurredAt  *string        `json:"occurred_at,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

type warehouseResultRequest struct {
	Result      string         `json:"result"`
	Disposition string         `json:"disposition"`
	OperatorID  string         `json:"operator_id,omitempty"`
	Notes       string         `json:"notes,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type createInspectionRuleRequest struct {
	Name           string         `json:"name"`
	Priority       int            `json:"priority"`
	Condition      map[string]any `json:"condition"`
	Decision       string         `json:"decision"`
	Recommendation string         `json:"recommendation"`
	Notes          string         `json:"notes,omitempty"`
	Enabled        *bool          `json:"enabled,omitempty"`
}

type evaluateInspectionRequest struct {
	Attributes map[string]any `json:"attributes"`
	OperatorID string         `json:"operator_id,omitempty"`
	Notes      string         `json:"notes,omitempty"`
}
