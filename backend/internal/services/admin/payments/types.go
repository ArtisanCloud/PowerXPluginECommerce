package payments

import "time"

// ProviderDTO represents payment provider output.
type ProviderDTO struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	ProviderType    string    `json:"type"`
	Status          string    `json:"status"`
	FeeRate         float64   `json:"fee_rate"`
	SettlementCycle string    `json:"settlement_cycle"`
	Currency        string    `json:"currency"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProviderListFilter struct {
	Keyword      string
	ProviderType string
	Status       string
}

type CreateProviderRequest struct {
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Status          string                 `json:"status"`
	FeeRate         float64                `json:"fee_rate"`
	SettlementCycle string                 `json:"settlement_cycle"`
	Currency        string                 `json:"currency"`
	Credentials     map[string]interface{} `json:"credentials"`
	RiskPolicy      map[string]interface{} `json:"risk_policy"`
}

type UpdateProviderRequest struct {
	Status          *string                `json:"status"`
	FeeRate         *float64               `json:"fee_rate"`
	SettlementCycle *string                `json:"settlement_cycle"`
	Currency        *string                `json:"currency"`
	Credentials     map[string]interface{} `json:"credentials"`
	RiskPolicy      map[string]interface{} `json:"risk_policy"`
}

// TransactionDTO represents payment transaction output.
type TransactionDTO struct {
	ID             uint64     `json:"id"`
	TransactionNo  string     `json:"transaction_no"`
	OrderID        string     `json:"order_id"`
	OrderNo        string     `json:"order_no"`
	ProviderID     uint64     `json:"provider_id"`
	PayMethod      string     `json:"pay_method"`
	AmountTotal    int64      `json:"amount_total"`
	AmountCurrency string     `json:"amount_currency"`
	FeeAmount      int64      `json:"fee_amount"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	FailureReason  string     `json:"failure_reason,omitempty"`
}

type TransactionListFilter struct {
	Status     string
	ProviderID uint64
	From       *time.Time
	To         *time.Time
}

type CreateRefundRequest struct {
	AmountMinor int64  `json:"amount_minor"`
	Reason      string `json:"reason"`
}

type RefundDTO struct {
	ID           uint64     `json:"id"`
	RefundNo     string     `json:"refund_no"`
	RefundAmount int64      `json:"refund_amount"`
	Currency     string     `json:"refund_currency"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type RiskEventDTO struct {
	ID         uint64     `json:"id"`
	RiskType   string     `json:"risk_type"`
	RiskScore  int        `json:"risk_score"`
	Action     string     `json:"action"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

type SplitRuleDTO struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SplitResultDTO struct {
	ID          uint64    `json:"id"`
	RuleID      uint64    `json:"rule_id"`
	Participant string    `json:"participant"`
	Amount      int64     `json:"amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ManualPaymentCreateRequest struct {
	OrderID     string `json:"order_id"`
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
	PayMethod   string `json:"pay_method"`
	ProviderID  uint64 `json:"provider_id"`
	ProofNo     string `json:"proof_no"`
	Note        string `json:"note"`
}

type ManualPaymentReviewRequest struct {
	Reason string `json:"reason"`
}

type ManualPaymentReviewDTO struct {
	ID            uint64     `json:"id"`
	OrderID       string     `json:"order_id"`
	OrderNo       string     `json:"order_no"`
	TransactionID *uint64    `json:"transaction_id,omitempty"`
	ProviderID    uint64     `json:"provider_id"`
	PayMethod     string     `json:"pay_method"`
	AmountMinor   int64      `json:"amount_minor"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	SubmittedBy   string     `json:"submitted_by"`
	SubmittedAt   time.Time  `json:"submitted_at"`
	ReviewedBy    string     `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewReason  string     `json:"review_reason,omitempty"`
	ProofNo       string     `json:"proof_no,omitempty"`
	Note          string     `json:"note,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ManualPaymentReviewLogDTO struct {
	ID           uint64     `json:"id"`
	ReviewID     uint64     `json:"review_id"`
	OrderID      string     `json:"order_id"`
	OrderNo      string     `json:"order_no"`
	PayMethod    string     `json:"pay_method"`
	AmountMinor  int64      `json:"amount_minor"`
	Currency     string     `json:"currency"`
	Status       string     `json:"status"`
	Action       string     `json:"action"`
	SubmittedBy  string     `json:"submitted_by"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	ReviewedBy   string     `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	ReviewReason string     `json:"review_reason,omitempty"`
	ProofNo      string     `json:"proof_no,omitempty"`
	Note         string     `json:"note,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ReconciliationDTO struct {
	ID              uint64     `json:"id"`
	PeriodType      string     `json:"period_type"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	DiffCount       int        `json:"diff_count"`
	DiffTotalAmount int64      `json:"diff_total_amount"`
	Status          string     `json:"status"`
	ProcessedBy     string     `json:"processed_by,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ReconciliationItemDTO struct {
	ID               uint64     `json:"id"`
	ReconciliationID uint64     `json:"reconciliation_id"`
	TransactionID    uint64     `json:"transaction_id"`
	DiffType         string     `json:"diff_type"`
	DiffAmount       int64      `json:"diff_amount"`
	Resolution       string     `json:"resolution,omitempty"`
	ResolvedBy       string     `json:"resolved_by,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type CreateReconciliationItemInput struct {
	TransactionID uint64 `json:"transaction_id"`
	DiffType      string `json:"diff_type"`
	DiffAmount    int64  `json:"diff_amount"`
}

type CreateReconciliationRequest struct {
	PeriodType  string                         `json:"period_type"`
	PeriodStart time.Time                      `json:"period_start"`
	PeriodEnd   time.Time                      `json:"period_end"`
	Items       []CreateReconciliationItemInput `json:"items"`
}

type ResolveReconciliationItemRequest struct {
	Resolution string `json:"resolution"`
}
