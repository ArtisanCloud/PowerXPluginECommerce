package models

import "time"

// PaymentRefund represents a refund record.
type PaymentRefund struct {
	BaseModel
	TransactionID  uint64     `gorm:"not null;index" json:"transaction_id"`
	RefundNo       string     `gorm:"type:varchar(64);not null;index" json:"refund_no"`
	RefundAmount   int64      `gorm:"not null;default:0" json:"refund_amount"`
	RefundCurrency string     `gorm:"type:varchar(8);not null" json:"refund_currency"`
	Status         string     `gorm:"type:varchar(32);not null;index" json:"status"`
	Reason         string     `gorm:"type:varchar(256)" json:"reason,omitempty"`
	CompletedAt    *time.Time `gorm:"" json:"completed_at,omitempty"`
}

func (PaymentRefund) TableName() string { return S(TablePaymentRefunds) }
