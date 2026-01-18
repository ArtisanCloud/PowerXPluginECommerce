package models

// PaymentSplitResult captures revenue sharing results.
type PaymentSplitResult struct {
	BaseModel
	TransactionID uint64 `gorm:"not null;index" json:"transaction_id"`
	RuleID        uint64 `gorm:"not null;index" json:"rule_id"`
	Participant   string `gorm:"type:varchar(128);not null" json:"participant"`
	Amount        int64  `gorm:"not null;default:0" json:"amount"`
	Status        string `gorm:"type:varchar(32);not null;index" json:"status"`
}

func (PaymentSplitResult) TableName() string { return S(TablePaymentSplitResults) }
