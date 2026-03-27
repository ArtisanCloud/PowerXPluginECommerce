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

type etaQueryRequest struct {
	WaybillIDs      []string `json:"waybill_ids,omitempty"`
	DestinationZone string   `json:"destination_zone,omitempty"`
	Timezone        string   `json:"timezone,omitempty"`
	ForceRecompute  bool     `json:"force_recompute,omitempty"`
}

type upsertRoutingRuleRequest struct {
	ID              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	DestinationZone string         `json:"destination_zone,omitempty"`
	CarrierID       string         `json:"carrier_id"`
	ServiceCode     string         `json:"service_code,omitempty"`
	Priority        int            `json:"priority,omitempty"`
	MinWeight       float64        `json:"min_weight,omitempty"`
	MaxWeight       float64        `json:"max_weight,omitempty"`
	MinOrderAmount  float64        `json:"min_order_amount,omitempty"`
	MaxOrderAmount  float64        `json:"max_order_amount,omitempty"`
	Fallback        *bool          `json:"fallback,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	RuleConfig      map[string]any `json:"rule_config,omitempty"`
}

type previewRoutingRequest struct {
	WarehouseID        string  `json:"warehouse_id,omitempty"`
	DestinationZone    string  `json:"destination_zone,omitempty"`
	Weight             float64 `json:"weight,omitempty"`
	OrderAmount        float64 `json:"order_amount,omitempty"`
	PreferredCarrierID string  `json:"preferred_carrier_id,omitempty"`
	ServiceCode        string  `json:"service_code,omitempty"`
}

type initiateRedeliveryRequest struct {
	WaybillID  string         `json:"waybill_id"`
	RequestKey string         `json:"request_key,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	OperatorID string         `json:"operator_id,omitempty"`
	Address    map[string]any `json:"address,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type updateRedeliveryAddressRequest struct {
	Address    map[string]any `json:"address"`
	OperatorID string         `json:"operator_id,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

type redispatchRedeliveryRequest struct {
	RequestKey string         `json:"request_key,omitempty"`
	OperatorID string         `json:"operator_id,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type closeRedeliveryRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type upsertRiskRuleRequest struct {
	ID          string         `json:"id,omitempty"`
	Name        string         `json:"name"`
	MatchField  string         `json:"match_field,omitempty"`
	MatchMode   string         `json:"match_mode,omitempty"`
	Pattern     string         `json:"pattern"`
	Decision    string         `json:"decision,omitempty"`
	RiskLevel   string         `json:"risk_level,omitempty"`
	Priority    int            `json:"priority,omitempty"`
	Enabled     *bool          `json:"enabled,omitempty"`
	Description string         `json:"description,omitempty"`
	RuleConfig  map[string]any `json:"rule_config,omitempty"`
}

type upsertBlacklistEntryRequest struct {
	ID             string         `json:"id,omitempty"`
	EntryType      string         `json:"entry_type,omitempty"`
	RecipientName  string         `json:"recipient_name,omitempty"`
	RecipientPhone string         `json:"recipient_phone,omitempty"`
	AddressLine    string         `json:"address_line,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Status         string         `json:"status,omitempty"`
	ExpiresAt      *string        `json:"expires_at,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type evaluateRiskRequest struct {
	WaybillID       string         `json:"waybill_id,omitempty"`
	WaybillNo       string         `json:"waybill_no,omitempty"`
	RecipientName   string         `json:"recipient_name,omitempty"`
	RecipientPhone  string         `json:"recipient_phone,omitempty"`
	DestinationLine string         `json:"destination_line,omitempty"`
	OperatorID      string         `json:"operator_id,omitempty"`
	Context         map[string]any `json:"context,omitempty"`
}

type releaseRiskHitRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type batchPrintLabelsRequest struct {
	WaybillIDs     []string `json:"waybill_ids"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
	ReprintReason  string   `json:"reprint_reason,omitempty"`
	MaxAttempts    int      `json:"max_attempts,omitempty"`
}

type retryLabelPrintRequest struct {
	TaskIDs []string `json:"task_ids,omitempty"`
}

type createTrackingSyncJobRequest struct {
	CarrierID     string `json:"carrier_id,omitempty"`
	WaybillStatus string `json:"waybill_status,omitempty"`
	BatchLimit    int    `json:"batch_limit,omitempty"`
	EventLimit    int    `json:"event_limit,omitempty"`
}

type cancelTrackingSyncJobRequest struct {
	Reason string `json:"reason,omitempty"`
}

type rateQuoteRequest struct {
	Region      string  `json:"region"`
	Weight      float64 `json:"weight,omitempty"`
	PieceCount  int     `json:"piece_count,omitempty"`
	Volume      float64 `json:"volume,omitempty"`
	OrderAmount float64 `json:"order_amount,omitempty"`
}

type createBillingCaseRequest struct {
	WaybillID string         `json:"waybill_id"`
	Reason    string         `json:"reason,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type transitionBillingCaseRequest struct {
	Action     string         `json:"action"`
	OperatorID string         `json:"operator_id,omitempty"`
	Note       string         `json:"note,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type upsertNotificationTemplateRequest struct {
	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name"`
	Event    string         `json:"event"`
	Channel  string         `json:"channel,omitempty"`
	Title    string         `json:"title,omitempty"`
	Body     string         `json:"body"`
	Enabled  *bool          `json:"enabled,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type sendNotificationRequest struct {
	WaybillID      string         `json:"waybill_id"`
	Event          string         `json:"event"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
}

type retryNotificationRequest struct {
	RecordID string `json:"record_id,omitempty"`
}
