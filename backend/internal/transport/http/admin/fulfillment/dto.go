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
