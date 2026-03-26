package logistics

type upsertCarrierRequest struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	Code         string         `json:"code"`
	Type         string         `json:"type,omitempty"`
	Status       string         `json:"status,omitempty"`
	ContactName  string         `json:"contact_name,omitempty"`
	ContactPhone string         `json:"contact_phone,omitempty"`
	Capabilities map[string]any `json:"capabilities,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

type upsertRateTemplateRequest struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Currency string         `json:"currency,omitempty"`
	Channels map[string]any `json:"channels,omitempty"`
	Rules    map[string]any `json:"rules,omitempty"`
}

type createWaybillRequest struct {
	OrderID              string   `json:"order_id"`
	CarrierID            string   `json:"carrier_id"`
	ServiceCode          string   `json:"service_code"`
	WaybillNo            string   `json:"waybill_no,omitempty"`
	ManualFallbackReason string   `json:"manual_fallback_reason,omitempty"`
	PackageNo            int      `json:"package_no,omitempty"`
	PackageKey           string   `json:"package_key,omitempty"`
	ShipmentItems        []string `json:"shipment_items,omitempty"`
	OrderItemCount       int      `json:"order_item_count,omitempty"`
}

type appendTrackingRequest struct {
	EventID     string         `json:"event_id,omitempty"`
	Status      string         `json:"status"`
	Source      string         `json:"source,omitempty"`
	Description string         `json:"description,omitempty"`
	OccurredAt  *string        `json:"occurred_at,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

type webhookRequest struct {
	EventID    string         `json:"event_id"`
	WaybillNo  string         `json:"waybill_no"`
	Status     string         `json:"status"`
	OccurredAt *string        `json:"occurred_at,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type updateWaybillCostRequest struct {
	ActualFeeAmount float64 `json:"actual_fee_amount"`
}
