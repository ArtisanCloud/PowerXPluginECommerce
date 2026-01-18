package models

import "time"

import "gorm.io/datatypes"

// PaymentTransaction represents a payment record.
type PaymentTransaction struct {
	BaseModel
	TransactionNo  string         `gorm:"type:varchar(64);not null;index" json:"transaction_no"`
	OrderID        string         `gorm:"type:uuid;index" json:"order_id,omitempty"`
	OrderNo        string         `gorm:"type:varchar(64);index" json:"order_no,omitempty"`
	ProviderID     uint64         `gorm:"not null;index" json:"provider_id"`
	PayMethod      string         `gorm:"type:varchar(32);not null" json:"pay_method"`
	AmountTotal    int64          `gorm:"not null;default:0" json:"amount_total"`
	AmountCurrency string         `gorm:"type:varchar(8);not null" json:"amount_currency"`
	FeeAmount      int64          `gorm:"not null;default:0" json:"fee_amount"`
	Status         string         `gorm:"type:varchar(32);not null;index" json:"status"`
	CompletedAt    *time.Time     `gorm:"" json:"completed_at,omitempty"`
	FailureReason  string         `gorm:"type:varchar(128)" json:"failure_reason,omitempty"`
	RiskFlag       bool           `gorm:"not null;default:false" json:"risk_flag"`
	Metadata       datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (PaymentTransaction) TableName() string { return S(TablePaymentTransactions) }
