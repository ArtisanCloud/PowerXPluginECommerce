package models

import (
	"time"

	"gorm.io/datatypes"
)

// OrderBenefitReview records coupon/giftcard application requests for an order.
type OrderBenefitReview struct {
	BaseModel
	OrderID         string         `gorm:"type:uuid;not null;index" json:"order_id"`
	OrderNo         string         `gorm:"type:varchar(64);not null;index" json:"order_no"`
	BenefitType     string         `gorm:"type:varchar(32);not null;index" json:"benefit_type"`
	BenefitCode     string         `gorm:"type:varchar(128);not null;index" json:"benefit_code"`
	ValueType       string         `gorm:"type:varchar(32);not null" json:"value_type"`
	Value           int64          `gorm:"not null;default:0" json:"value"`
	AmountMinor     int64          `gorm:"not null;default:0" json:"amount_minor"`
	Currency        string         `gorm:"type:varchar(8);not null" json:"currency"`
	StackingAllowed bool           `gorm:"not null;default:true" json:"stacking_allowed"`
	Status          string         `gorm:"type:varchar(32);not null;index" json:"status"`
	SubmittedBy     string         `gorm:"type:varchar(128);not null" json:"submitted_by"`
	SubmittedAt     time.Time      `gorm:"autoCreateTime" json:"submitted_at"`
	ReviewedBy      string         `gorm:"type:varchar(128)" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time     `gorm:"" json:"reviewed_at,omitempty"`
	ReviewReason    string         `gorm:"type:varchar(256)" json:"review_reason,omitempty"`
	Note            string         `gorm:"type:varchar(256)" json:"note,omitempty"`
	Metadata        datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (OrderBenefitReview) TableName() string { return S(TableOrderBenefitReviews) }
