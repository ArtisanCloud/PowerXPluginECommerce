package models

import "gorm.io/datatypes"

// PaymentSplitRule defines revenue sharing rules.
type PaymentSplitRule struct {
	BaseModel
	Name         string         `gorm:"type:varchar(128);not null" json:"name"`
	Participants datatypes.JSON `gorm:"type:jsonb" json:"participants"`
	Ratio        datatypes.JSON `gorm:"type:jsonb" json:"ratio"`
	Status       string         `gorm:"type:varchar(32);not null;index" json:"status"`
}

func (PaymentSplitRule) TableName() string { return S(TablePaymentSplitRules) }
