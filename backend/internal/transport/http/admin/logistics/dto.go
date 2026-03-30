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

type evaluateCarrierProfilesRequest struct {
	CarrierID  string `json:"carrier_id,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type confirmCarrierProfileRatingRequest struct {
	Rating     string `json:"rating"`
	OperatorID string `json:"operator_id,omitempty"`
}

type retireCarrierProfileRequest struct {
	Reason     string `json:"reason,omitempty"`
	Force      bool   `json:"force,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type restoreCarrierProfileRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
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

type upsertTrackingSyncScheduleRequest struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name"`
	CronExpr        string `json:"cron_expr"`
	CarrierID       string `json:"carrier_id,omitempty"`
	WaybillStatus   string `json:"waybill_status,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
	MaxConcurrency  int    `json:"max_concurrency,omitempty"`
	DedupeWindowSec int    `json:"dedupe_window_sec,omitempty"`
	BatchLimit      int    `json:"batch_limit,omitempty"`
	EventLimit      int    `json:"event_limit,omitempty"`
}

type toggleTrackingSyncScheduleRequest struct {
	Enabled bool `json:"enabled"`
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

type upsertExceptionOrchestrationRuleRequest struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	TriggerEvent string         `json:"trigger_event"`
	Action       string         `json:"action"`
	Priority     int            `json:"priority,omitempty"`
	Enabled      *bool          `json:"enabled,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

type executeExceptionOrchestrationRequest struct {
	RuleID    string `json:"rule_id"`
	WaybillID string `json:"waybill_id,omitempty"`
	WaybillNo string `json:"waybill_no,omitempty"`
	Trigger   string `json:"trigger,omitempty"`
}

type checkAddressValidationRequest struct {
	RequestKey string         `json:"request_key,omitempty"`
	WaybillID  string         `json:"waybill_id,omitempty"`
	WaybillNo  string         `json:"waybill_no,omitempty"`
	Address    string         `json:"address"`
	Context    map[string]any `json:"context,omitempty"`
}

type upsertRoutingOptimizerStrategyRequest struct {
	Name             string         `json:"name,omitempty"`
	TimelinessWeight float64        `json:"timeliness_weight"`
	CostWeight       float64        `json:"cost_weight"`
	QuotaWeight      float64        `json:"quota_weight"`
	RiskWeight       float64        `json:"risk_weight"`
	FallbackStrategy string         `json:"fallback_strategy,omitempty"`
	Enabled          *bool          `json:"enabled,omitempty"`
	Config           map[string]any `json:"config,omitempty"`
}

type simulateRoutingOptimizerRequest struct {
	RequestKey          string   `json:"request_key,omitempty"`
	WarehouseID         string   `json:"warehouse_id,omitempty"`
	DestinationZone     string   `json:"destination_zone,omitempty"`
	Weight              float64  `json:"weight,omitempty"`
	OrderAmount         float64  `json:"order_amount,omitempty"`
	PreferredCarrierID  string   `json:"preferred_carrier_id,omitempty"`
	AvailableCarrierIDs []string `json:"available_carrier_ids,omitempty"`
	RealtimeDegraded    bool     `json:"realtime_degraded,omitempty"`
}

type createSettlementBatchRequest struct {
	CarrierID string `json:"carrier_id,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
}

type handleSettlementDiffRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Note       string `json:"note,omitempty"`
}

type reconciliationInputRecordRequest struct {
	WaybillID     string  `json:"waybill_id,omitempty"`
	WaybillNo     string  `json:"waybill_no"`
	CarrierID     string  `json:"carrier_id,omitempty"`
	BillAmount    float64 `json:"bill_amount"`
	BankAmount    float64 `json:"bank_amount"`
	InvoiceAmount float64 `json:"invoice_amount"`
}

type createReconciliationBatchRequest struct {
	CarrierID string                             `json:"carrier_id,omitempty"`
	From      string                             `json:"from,omitempty"`
	To        string                             `json:"to,omitempty"`
	Records   []reconciliationInputRecordRequest `json:"records,omitempty"`
}

type handleReconciliationCaseRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Note       string `json:"note,omitempty"`
}

type upsertControlTowerSubscriptionRequest struct {
	ID              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	CarrierID       string         `json:"carrier_id,omitempty"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	DestinationZone string         `json:"destination_zone,omitempty"`
	MinOnTimeRate   float64        `json:"min_on_time_rate,omitempty"`
	MaxTimeoutCount int            `json:"max_timeout_count,omitempty"`
	MaxCostAmount   float64        `json:"max_cost_amount,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	Config          map[string]any `json:"config,omitempty"`
}

type upsertCapacityPlanRequest struct {
	ID               string         `json:"id,omitempty"`
	Name             string         `json:"name"`
	CarrierID        string         `json:"carrier_id"`
	WarehouseID      string         `json:"warehouse_id,omitempty"`
	DestinationZone  string         `json:"destination_zone,omitempty"`
	DailyCapacity    int            `json:"daily_capacity"`
	ReservedCapacity int            `json:"reserved_capacity,omitempty"`
	UsedCapacity     int            `json:"used_capacity,omitempty"`
	Status           string         `json:"status,omitempty"`
	Config           map[string]any `json:"config,omitempty"`
}

type upsertCrossborderDocumentRequest struct {
	ID          string         `json:"id,omitempty"`
	WaybillID   string         `json:"waybill_id,omitempty"`
	WaybillNo   string         `json:"waybill_no"`
	DocType     string         `json:"doc_type"`
	DocNo       string         `json:"doc_no"`
	CountryFrom string         `json:"country_from,omitempty"`
	CountryTo   string         `json:"country_to,omitempty"`
	Status      string         `json:"status,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type quoteCrossborderTaxRequest struct {
	RequestKey    string  `json:"request_key,omitempty"`
	WaybillID     string  `json:"waybill_id,omitempty"`
	WaybillNo     string  `json:"waybill_no,omitempty"`
	Destination   string  `json:"destination_country"`
	Currency      string  `json:"currency,omitempty"`
	DeclaredValue float64 `json:"declared_value"`
	ShippingFee   float64 `json:"shipping_fee,omitempty"`
	InsuranceFee  float64 `json:"insurance_fee,omitempty"`
}

type upsertCrossborderTrackingMapRequest struct {
	ID               string         `json:"id,omitempty"`
	Provider         string         `json:"provider"`
	ProviderStatus   string         `json:"provider_status"`
	NormalizedStatus string         `json:"normalized_status"`
	Description      string         `json:"description,omitempty"`
	Priority         int            `json:"priority,omitempty"`
	Enabled          *bool          `json:"enabled,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type normalizeCrossborderTrackingRequest struct {
	Provider       string `json:"provider"`
	ProviderStatus string `json:"provider_status"`
}

type customsRuleDefinitionRequest struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Field      string `json:"field"`
	Operator   string `json:"operator"`
	Value      any    `json:"value"`
	RiskLevel  string `json:"risk_level"`
	Suggestion string `json:"suggestion"`
	Enabled    bool   `json:"enabled"`
}

type upsertCustomsRulePackRequest struct {
	ID               string         `json:"id,omitempty"`
	Name             string         `json:"name"`
	CountryCode      string         `json:"country_code"`
	Status           string         `json:"status,omitempty"`
	Strategy         string         `json:"strategy,omitempty"`
	DefaultRiskLevel string         `json:"default_risk_level,omitempty"`
	Description      string         `json:"description,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

type publishCustomsRuleVersionRequest struct {
	VersionNo   int                            `json:"version_no,omitempty"`
	Status      string                         `json:"status,omitempty"`
	HitStrategy string                         `json:"hit_strategy,omitempty"`
	Rules       []customsRuleDefinitionRequest `json:"rules"`
	RiskConfig  map[string]any                 `json:"risk_config,omitempty"`
}

type customsPrecheckRequest struct {
	PackID        string         `json:"pack_id,omitempty"`
	CountryCode   string         `json:"country_code,omitempty"`
	WaybillNo     string         `json:"waybill_no,omitempty"`
	DeclaredValue float64        `json:"declared_value,omitempty"`
	TaxNo         string         `json:"tax_no,omitempty"`
	HSCode        string         `json:"hs_code,omitempty"`
	DocumentType  string         `json:"document_type,omitempty"`
	DocumentCount int            `json:"document_count,omitempty"`
	ManualRelease bool           `json:"manual_release,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
}

type kpiDashboardQueryRequest struct {
	WindowHours     int    `json:"window_hours,omitempty"`
	Dimension       string `json:"dimension,omitempty"`
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type upsertSLOGuardPolicyRequest struct {
	ID                string         `json:"id,omitempty"`
	Name              string         `json:"name"`
	CarrierID         string         `json:"carrier_id,omitempty"`
	WindowHours       int            `json:"window_hours,omitempty"`
	MinSuccessRate    float64        `json:"min_success_rate,omitempty"`
	MaxP95LatencyMS   int            `json:"max_p95_latency_ms,omitempty"`
	MaxFailedRequests int            `json:"max_failed_requests,omitempty"`
	Action            string         `json:"action,omitempty"`
	ThrottleRatio     int            `json:"throttle_ratio,omitempty"`
	Enabled           *bool          `json:"enabled,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

type evaluateSLOGuardRequest struct {
	WindowHours int    `json:"window_hours,omitempty"`
	CarrierID   string `json:"carrier_id,omitempty"`
}

type releaseSLOGuardRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type generateCapacityForecastRequest struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	WindowDays      int    `json:"window_days,omitempty"`
	OperatorID      string `json:"operator_id,omitempty"`
}

type applyCapacityForecastRequest struct {
	OperatorID string `json:"operator_id,omitempty"`
}

type upsertFulfillmentSandboxScenarioRequest struct {
	ID              string         `json:"id,omitempty"`
	Name            string         `json:"name"`
	CarrierID       string         `json:"carrier_id,omitempty"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	DestinationZone string         `json:"destination_zone,omitempty"`
	BaselineConfig  map[string]any `json:"baseline_config,omitempty"`
	StrategyConfig  map[string]any `json:"strategy_config,omitempty"`
	Status          string         `json:"status,omitempty"`
	Description     string         `json:"description,omitempty"`
	OperatorID      string         `json:"operator_id,omitempty"`
}

type runFulfillmentSandboxRequest struct {
	ScenarioID string `json:"scenario_id"`
	WindowDays int    `json:"window_days,omitempty"`
	Strategy   string `json:"strategy,omitempty"`
	RequestKey string `json:"request_key,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type compareFulfillmentSandboxRequest struct {
	BaselineRunID  string `json:"baseline_run_id"`
	CandidateRunID string `json:"candidate_run_id"`
}

type upsertPolicyOrchestrationFlowRequest struct {
	ID                string         `json:"id,omitempty"`
	Name              string         `json:"name"`
	Priority          int            `json:"priority,omitempty"`
	Status            string         `json:"status,omitempty"`
	FlowDefinition    map[string]any `json:"flow_definition,omitempty"`
	ConflictRelations []string       `json:"conflict_relations,omitempty"`
	GrayReleaseConfig map[string]any `json:"gray_release_config,omitempty"`
	Description       string         `json:"description,omitempty"`
	OperatorID        string         `json:"operator_id,omitempty"`
}

type previewPolicyOrchestrationConflictsRequest struct {
	FlowID            string         `json:"flow_id,omitempty"`
	Name              string         `json:"name,omitempty"`
	ConflictRelations []string       `json:"conflict_relations,omitempty"`
	FlowDefinition    map[string]any `json:"flow_definition,omitempty"`
	GrayReleaseConfig map[string]any `json:"gray_release_config,omitempty"`
}

type publishPolicyOrchestrationFlowRequest struct {
	FlowID          string         `json:"flow_id,omitempty"`
	RequestKey      string         `json:"request_key,omitempty"`
	Force           bool           `json:"force,omitempty"`
	ChangeSummary   string         `json:"change_summary,omitempty"`
	GrayReleasePlan map[string]any `json:"gray_release_plan,omitempty"`
	OperatorID      string         `json:"operator_id,omitempty"`
}

type rollbackPolicyOrchestrationFlowRequest struct {
	FlowID          string `json:"flow_id,omitempty"`
	TargetVersionID string `json:"target_version_id"`
	Reason          string `json:"reason,omitempty"`
	OperatorID      string `json:"operator_id,omitempty"`
}

type generateQualityAuditReportRequest struct {
	CarrierID        string `json:"carrier_id,omitempty"`
	WarehouseID      string `json:"warehouse_id,omitempty"`
	DestinationZone  string `json:"destination_zone,omitempty"`
	ReportPeriodFrom string `json:"report_period_from,omitempty"`
	ReportPeriodTo   string `json:"report_period_to,omitempty"`
	WindowHours      int    `json:"window_hours,omitempty"`
	OperatorID       string `json:"operator_id,omitempty"`
}

type analyzeTrackingRootCauseRequest struct {
	CarrierID       string `json:"carrier_id,omitempty"`
	WarehouseID     string `json:"warehouse_id,omitempty"`
	DestinationZone string `json:"destination_zone,omitempty"`
	WindowHours     int    `json:"window_hours,omitempty"`
	Limit           int    `json:"limit,omitempty"`
}

type handleTrackingRootCauseRequest struct {
	Action     string `json:"action"`
	Status     string `json:"status,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
	ResultNote string `json:"result_note,omitempty"`
}

type allocationRequest struct {
	RequestKey       string `json:"request_key,omitempty"`
	WaybillID        string `json:"waybill_id,omitempty"`
	OrderID          string `json:"order_id,omitempty"`
	CarrierID        string `json:"carrier_id,omitempty"`
	WarehouseID      string `json:"warehouse_id,omitempty"`
	DestinationZone  string `json:"destination_zone,omitempty"`
	Strategy         string `json:"strategy,omitempty"`
	OperatorID       string `json:"operator_id,omitempty"`
	PreferredCarrier string `json:"preferred_carrier,omitempty"`
}

type overrideAllocationRequest struct {
	RequestKey string `json:"request_key,omitempty"`
	DecisionID string `json:"decision_id,omitempty"`
	CarrierID  string `json:"carrier_id"`
	Reason     string `json:"reason,omitempty"`
	OperatorID string `json:"operator_id,omitempty"`
}

type suggestInterwarehouseAllocationRequest struct {
	RequestKey        string `json:"request_key,omitempty"`
	WaybillID         string `json:"waybill_id,omitempty"`
	OrderID           string `json:"order_id,omitempty"`
	CarrierID         string `json:"carrier_id,omitempty"`
	SourceWarehouseID string `json:"source_warehouse_id,omitempty"`
	DestinationZone   string `json:"destination_zone,omitempty"`
	RequiredQty       int    `json:"required_qty,omitempty"`
	OperatorID        string `json:"operator_id,omitempty"`
}

type confirmInterwarehouseAllocationRequest struct {
	CandidateID       string `json:"candidate_id,omitempty"`
	RequestKey        string `json:"request_key,omitempty"`
	TargetWarehouseID string `json:"target_warehouse_id,omitempty"`
	OperatorID        string `json:"operator_id,omitempty"`
}

type upsertLastmileRecoveryRuleRequest struct {
	ID           string         `json:"id,omitempty"`
	Name         string         `json:"name"`
	TriggerEvent string         `json:"trigger_event"`
	Action       string         `json:"action"`
	Priority     int            `json:"priority,omitempty"`
	MaxRetries   int            `json:"max_retries,omitempty"`
	Enabled      *bool          `json:"enabled,omitempty"`
	Config       map[string]any `json:"config,omitempty"`
}

type executeLastmileRecoveryRequest struct {
	RequestKey   string `json:"request_key,omitempty"`
	RuleID       string `json:"rule_id,omitempty"`
	WaybillID    string `json:"waybill_id,omitempty"`
	WaybillNo    string `json:"waybill_no,omitempty"`
	TriggerEvent string `json:"trigger_event,omitempty"`
}

type takeoverLastmileRecoveryRequest struct {
	Action     string `json:"action"`
	OperatorID string `json:"operator_id,omitempty"`
	Reason     string `json:"reason,omitempty"`
}
