package fulfillment

type createTaskRequest struct {
	OrderID     string         `json:"order_id"`
	WarehouseID string         `json:"warehouse_id"`
	AssignedTo  string         `json:"assigned_to,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type completeTaskRequest struct {
	OperatorID string         `json:"operator_id,omitempty"`
	Detail     map[string]any `json:"detail,omitempty"`
}

type reportExceptionRequest struct {
	TaskID    string         `json:"task_id"`
	WaybillID string         `json:"waybill_id,omitempty"`
	Type      string         `json:"type"`
	Reason    string         `json:"reason,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type createWaveRequest struct {
	Name        string         `json:"name,omitempty"`
	WarehouseID string         `json:"warehouse_id"`
	TaskIDs     []string       `json:"task_ids"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type advanceWaveRequest struct {
	Status      string   `json:"status"`
	OperatorID  string   `json:"operator_id,omitempty"`
	FailTaskIDs []string `json:"fail_task_ids,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

type reassignWaveTaskRequest struct {
	AssignedTo string `json:"assigned_to"`
	OperatorID string `json:"operator_id,omitempty"`
}

type createWaveStrategyRequest struct {
	Name            string         `json:"name"`
	WarehouseID     string         `json:"warehouse_id,omitempty"`
	CarrierCode     string         `json:"carrier_code,omitempty"`
	TimeWindow      string         `json:"time_window,omitempty"`
	PriorityBand    string         `json:"priority_band,omitempty"`
	MaxTasksPerWave int            `json:"max_tasks_per_wave,omitempty"`
	Enabled         *bool          `json:"enabled,omitempty"`
	Rules           map[string]any `json:"rules,omitempty"`
}

type previewWaveStrategyRequest struct {
	StrategyID string `json:"strategy_id"`
}

type createOutboundRequest struct {
	TaskID    string         `json:"task_id"`
	WaybillID string         `json:"waybill_id,omitempty"`
	Items     []pickLineItem `json:"items,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type pickLineItem struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type executeOutboundRequest struct {
	OperatorID string         `json:"operator_id,omitempty"`
	PackageNo  int            `json:"package_no,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type rollbackOutboundRequest struct {
	Reason string `json:"reason,omitempty"`
}
