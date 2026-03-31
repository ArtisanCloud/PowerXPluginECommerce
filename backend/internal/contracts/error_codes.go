package contracts

const (
	// After-sales domain error codes.
	ErrCodeAfterSalesServiceUnavailable = "AFTER_SALES_SERVICE_UNAVAILABLE"
	ErrCodeAfterSalesCaseNotFound       = "AFTER_SALES_CASE_NOT_FOUND"
	ErrCodeAfterSalesInvalidAction      = "AFTER_SALES_INVALID_ACTION"
	ErrCodeAfterSalesInvalidState       = "AFTER_SALES_INVALID_STATE_TRANSITION"
	ErrCodeAfterSalesCoreFieldsFrozen   = "AFTER_SALES_CORE_FIELDS_FROZEN"
	ErrCodeAfterSalesReasonRequired     = "AFTER_SALES_REASON_REQUIRED"
	ErrCodeAfterSalesRefundDuplicated   = "AFTER_SALES_REFUND_DUPLICATED"
	ErrCodeAfterSalesReverseNotAllowed  = "AFTER_SALES_REVERSE_NOT_ALLOWED"
	ErrCodeAfterSalesReverseNotFound    = "AFTER_SALES_REVERSE_WAYBILL_NOT_FOUND"
	ErrCodeAfterSalesReverseMismatch    = "AFTER_SALES_REVERSE_ORDER_MISMATCH"

	// Subscription reconciliation domain error codes.
	ErrCodeReconciliationBatchNotFound       = "RECONCILIATION_BATCH_NOT_FOUND"
	ErrCodeReconciliationDeltaNotFound       = "RECONCILIATION_DELTA_NOT_FOUND"
	ErrCodeReconciliationTaskAlreadyOpen     = "RECONCILIATION_TASK_ALREADY_OPEN"
	ErrCodeReconciliationUnsupportedAction   = "RECONCILIATION_UNSUPPORTED_ACTION"
	ErrCodeReconciliationInvalidBillingCycle = "RECONCILIATION_INVALID_BILLING_CYCLE"
)
