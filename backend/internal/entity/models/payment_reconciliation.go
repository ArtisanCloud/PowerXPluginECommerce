package models

import "time"

// PaymentReconciliation represents a reconciliation batch.
type PaymentReconciliation struct {
	BaseModel
	PeriodType      string     `gorm:"type:varchar(16);not null;index" json:"period_type"`
	PeriodStart     time.Time  `gorm:"not null;index" json:"period_start"`
	PeriodEnd       time.Time  `gorm:"not null;index" json:"period_end"`
	DiffCount       int        `gorm:"not null;default:0" json:"diff_count"`
	DiffTotalAmount int64      `gorm:"not null;default:0" json:"diff_total_amount"`
	Status          string     `gorm:"type:varchar(32);not null;index" json:"status"`
	ProcessedBy     string     `gorm:"type:varchar(64)" json:"processed_by,omitempty"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
}

func (PaymentReconciliation) TableName() string { return S(TablePaymentReconciliations) }
