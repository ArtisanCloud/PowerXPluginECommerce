package models

import (
	"time"

	"gorm.io/datatypes"
)

// PaymentManualReviewLog records manual payment review actions.
type PaymentManualReviewLog struct {
	BaseModel
	ReviewID     uint64         `gorm:"not null;index" json:"review_id"`
	OrderID      string         `gorm:"type:uuid;not null;index" json:"order_id"`
	OrderNo      string         `gorm:"type:varchar(64);not null;index" json:"order_no"`
	PayMethod    string         `gorm:"type:varchar(32);not null" json:"pay_method"`
	AmountMinor  int64          `gorm:"not null;default:0" json:"amount_minor"`
	Currency     string         `gorm:"type:varchar(8);not null" json:"currency"`
	Status       string         `gorm:"type:varchar(32);not null;index" json:"status"`
	Action       string         `gorm:"type:varchar(32);not null;index" json:"action"`
	SubmittedBy  string         `gorm:"type:varchar(128);not null" json:"submitted_by"`
	SubmittedAt  time.Time      `gorm:"not null" json:"submitted_at"`
	ReviewedBy   string         `gorm:"type:varchar(128)" json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time     `gorm:"" json:"reviewed_at,omitempty"`
	ReviewReason string         `gorm:"type:varchar(256)" json:"review_reason,omitempty"`
	ProofNo      string         `gorm:"type:varchar(128)" json:"proof_no,omitempty"`
	Note         string         `gorm:"type:varchar(256)" json:"note,omitempty"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (PaymentManualReviewLog) TableName() string { return S(TablePaymentManualReviewLogs) }
