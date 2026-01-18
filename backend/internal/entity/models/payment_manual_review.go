package models

import (
	"time"

	"gorm.io/datatypes"
)

// PaymentManualReview records manual payment submission and review status.
type PaymentManualReview struct {
	BaseModel
	OrderID       string         `gorm:"type:uuid;not null;index" json:"order_id"`
	OrderNo       string         `gorm:"type:varchar(64);not null;index" json:"order_no"`
	TransactionID *uint64        `gorm:"index" json:"transaction_id,omitempty"`
	ProviderID    uint64         `gorm:"not null;default:0" json:"provider_id"`
	PayMethod     string         `gorm:"type:varchar(32);not null" json:"pay_method"`
	AmountMinor   int64          `gorm:"not null;default:0" json:"amount_minor"`
	Currency      string         `gorm:"type:varchar(8);not null" json:"currency"`
	Status        string         `gorm:"type:varchar(32);not null;index" json:"status"`
	SubmittedBy   string         `gorm:"type:varchar(128);not null" json:"submitted_by"`
	SubmittedAt   time.Time      `gorm:"autoCreateTime" json:"submitted_at"`
	ReviewedBy    string         `gorm:"type:varchar(128)" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time     `gorm:"" json:"reviewed_at,omitempty"`
	ReviewReason  string         `gorm:"type:varchar(256)" json:"review_reason,omitempty"`
	ProofNo       string         `gorm:"type:varchar(128)" json:"proof_no,omitempty"`
	Note          string         `gorm:"type:varchar(256)" json:"note,omitempty"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (PaymentManualReview) TableName() string { return S(TablePaymentManualReviews) }
