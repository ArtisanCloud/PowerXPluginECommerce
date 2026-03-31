package after_sales

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
)

// AfterSaleDecision stores each operator review decision.
type AfterSaleDecision struct {
	ID               string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string    `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CaseID           string    `gorm:"column:case_id;type:uuid;not null;index" json:"case_id"`
	Decision         string    `gorm:"column:decision;type:varchar(32);not null;index" json:"decision"`
	RejectReasonCode string    `gorm:"column:reject_reason_code;type:varchar(64)" json:"reject_reason_code,omitempty"`
	DecisionNote     string    `gorm:"column:decision_note;type:text" json:"decision_note,omitempty"`
	DecidedBy        string    `gorm:"column:decided_by;type:varchar(128);not null" json:"decided_by"`
	DecidedAt        time.Time `gorm:"column:decided_at;not null;index" json:"decided_at"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AfterSaleDecision) TableName() string {
	return BaseModel.S(BaseModel.TableAfterSalesDecisions)
}
