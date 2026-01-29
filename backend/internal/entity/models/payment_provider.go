package models

import "gorm.io/datatypes"

// PaymentProvider stores payment channel configuration.
type PaymentProvider struct {
	BaseModel
	Name            string         `gorm:"type:varchar(128);not null;index" json:"name"`
	ProviderType    string         `gorm:"column:provider_type;type:varchar(64);not null;index" json:"type"`
	AppID           string         `gorm:"column:app_id;type:varchar(128);not null;default:'';index" json:"app_id"`
	MchID           string         `gorm:"column:mch_id;type:varchar(128);not null;default:'';index" json:"mch_id"`
	SerialNo        string         `gorm:"column:serial_no;type:varchar(128);not null;default:'';index" json:"serial_no"`
	NotifyURL       string         `gorm:"column:notify_url;type:varchar(512);not null;default:''" json:"notify_url"`
	Status          string         `gorm:"type:varchar(32);not null;index" json:"status"`
	IsDefault       bool           `gorm:"type:boolean;not null;default:false;index" json:"is_default"`
	FeeRate         float64        `gorm:"type:numeric(10,4);not null;default:0" json:"fee_rate"`
	SettlementCycle string         `gorm:"type:varchar(32);not null;default:'daily'" json:"settlement_cycle"`
	Currency        string         `gorm:"type:varchar(8);not null;default:'CNY'" json:"currency"`
	Credentials     datatypes.JSON `gorm:"type:jsonb" json:"credentials,omitempty"`
	RiskPolicy      datatypes.JSON `gorm:"type:jsonb" json:"risk_policy,omitempty"`
}

func (PaymentProvider) TableName() string { return S(TablePaymentProviders) }
