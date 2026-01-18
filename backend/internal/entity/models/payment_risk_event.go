package models

import "time"

// PaymentRiskEvent captures risk events for a transaction.
type PaymentRiskEvent struct {
	BaseModel
	TransactionID  uint64     `gorm:"not null;index" json:"transaction_id"`
	RiskType       string     `gorm:"type:varchar(64);not null" json:"risk_type"`
	RiskScore      int        `gorm:"not null;default:0" json:"risk_score"`
	Action         string     `gorm:"type:varchar(32);not null" json:"action"`
	ResolvedAt     *time.Time `gorm:"" json:"resolved_at,omitempty"`
	ResolutionNote string     `gorm:"type:varchar(256)" json:"resolution_note,omitempty"`
}

func (PaymentRiskEvent) TableName() string { return S(TablePaymentRiskEvents) }
