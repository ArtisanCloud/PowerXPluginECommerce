package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FulfillmentFinanceRisk stores payout/chargeback linked risk profile per waybill.
type FulfillmentFinanceRisk struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_finance_risk_waybill,priority:1" json:"tenant_uuid"`
	WaybillID           string         `gorm:"column:waybill_id;type:uuid;index;uniqueIndex:uk_logistics_finance_risk_waybill,priority:2" json:"waybill_id"`
	WaybillNo           string         `gorm:"column:waybill_no;type:varchar(128);not null;index" json:"waybill_no"`
	CarrierID           string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id"`
	BillingCaseID       string         `gorm:"column:billing_case_id;type:uuid;index" json:"billing_case_id,omitempty"`
	PayoutRiskScore     float64        `gorm:"column:payout_risk_score;type:numeric(6,2);not null;default:0" json:"payout_risk_score"`
	ChargebackRiskScore float64        `gorm:"column:chargeback_risk_score;type:numeric(6,2);not null;default:0" json:"chargeback_risk_score"`
	CompositeRiskScore  float64        `gorm:"column:composite_risk_score;type:numeric(6,2);not null;default:0;index" json:"composite_risk_score"`
	RiskLevel           string         `gorm:"column:risk_level;type:varchar(16);not null;default:'low';index" json:"risk_level"`
	ThresholdValue      float64        `gorm:"column:threshold_value;type:numeric(6,2);not null;default:70" json:"threshold_value"`
	StopLossAction      string         `gorm:"column:stop_loss_action;type:varchar(32);not null;default:'observe'" json:"stop_loss_action"`
	Status              string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	Suggestion          string         `gorm:"column:suggestion;type:text" json:"suggestion,omitempty"`
	RiskFactors         datatypes.JSON `gorm:"column:risk_factors;type:jsonb" json:"risk_factors,omitempty"`
	LastAction          string         `gorm:"column:last_action;type:varchar(32)" json:"last_action,omitempty"`
	LastActionBy        string         `gorm:"column:last_action_by;type:varchar(64)" json:"last_action_by,omitempty"`
	LastActionAt        *time.Time     `gorm:"column:last_action_at;index" json:"last_action_at,omitempty"`
	ActionCount         int            `gorm:"column:action_count;not null;default:0" json:"action_count"`
	ResolvedAt          *time.Time     `gorm:"column:resolved_at;index" json:"resolved_at,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FulfillmentFinanceRisk) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsFinanceRisks)
}

// FulfillmentFinanceRiskAudit stores action history for finance risk mitigation.
type FulfillmentFinanceRiskAudit struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_finance_risk_audit_req,priority:1" json:"tenant_uuid"`
	RiskID     string         `gorm:"column:risk_id;type:uuid;not null;index;uniqueIndex:uk_logistics_finance_risk_audit_req,priority:2" json:"risk_id"`
	RequestKey string         `gorm:"column:request_key;type:varchar(128);not null;uniqueIndex:uk_logistics_finance_risk_audit_req,priority:3" json:"request_key"`
	Action     string         `gorm:"column:action;type:varchar(32);not null;index" json:"action"`
	OperatorID string         `gorm:"column:operator_id;type:varchar(64)" json:"operator_id,omitempty"`
	Note       string         `gorm:"column:note;type:text" json:"note,omitempty"`
	Payload    datatypes.JSON `gorm:"column:payload;type:jsonb" json:"payload,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FulfillmentFinanceRiskAudit) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsFinanceRiskAudits)
}
