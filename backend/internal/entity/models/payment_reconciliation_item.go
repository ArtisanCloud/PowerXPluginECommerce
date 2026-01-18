package models

import "time"

// PaymentReconciliationItem captures a single reconciliation diff record.
type PaymentReconciliationItem struct {
	BaseModel
	ReconciliationID uint64     `gorm:"not null;index" json:"reconciliation_id"`
	TransactionID    uint64     `gorm:"not null;index" json:"transaction_id"`
	DiffType         string     `gorm:"type:varchar(64);not null;index" json:"diff_type"`
	DiffAmount       int64      `gorm:"not null;default:0" json:"diff_amount"`
	Resolution       string     `gorm:"type:varchar(255)" json:"resolution,omitempty"`
	ResolvedBy       string     `gorm:"type:varchar(64)" json:"resolved_by,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
}

func (PaymentReconciliationItem) TableName() string { return S(TablePaymentReconciliationItems) }
