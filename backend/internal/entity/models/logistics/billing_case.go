package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BillingCase tracks reconciliation dispute and resolution lifecycle.
type BillingCase struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_billing_case_no,priority:1" json:"tenant_uuid"`
	WaybillID  string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	CarrierID  string         `gorm:"column:carrier_id;type:uuid;not null;index" json:"carrier_id"`
	CaseNo     string         `gorm:"column:case_no;type:varchar(128);not null;index;uniqueIndex:uk_logistics_billing_case_no,priority:2" json:"case_no"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	DiffAmount float64        `gorm:"column:diff_amount;type:numeric(12,2);not null;default:0" json:"diff_amount"`
	Reason     string         `gorm:"column:reason;type:text" json:"reason,omitempty"`
	Resolution string         `gorm:"column:resolution;type:text" json:"resolution,omitempty"`
	Metadata   datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	ClosedAt   *time.Time     `gorm:"column:closed_at" json:"closed_at,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BillingCase) TableName() string { return BaseModel.S(BaseModel.TableLogisticsBillingCases) }
