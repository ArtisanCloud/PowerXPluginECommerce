package after_sales

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ReturnLogisticsLink links after-sale case and reverse waybill.
type ReturnLogisticsLink struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_after_sales_reverse_link_case,priority:1" json:"tenant_uuid"`
	CaseID           string         `gorm:"column:case_id;type:uuid;not null;index;index:idx_after_sales_reverse_link_case,priority:2" json:"case_id"`
	ReverseWaybillID string         `gorm:"column:reverse_waybill_id;type:uuid;index" json:"reverse_waybill_id,omitempty"`
	ReverseWaybillNo string         `gorm:"column:reverse_waybill_no;type:varchar(128);index" json:"reverse_waybill_no,omitempty"`
	CarrierCode      string         `gorm:"column:carrier_code;type:varchar(64)" json:"carrier_code,omitempty"`
	ReceiveStatus    string         `gorm:"column:receive_status;type:varchar(32);not null;default:'pending';index" json:"receive_status"`
	ReceivedAt       *time.Time     `gorm:"column:received_at;index" json:"received_at,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReturnLogisticsLink) TableName() string {
	return BaseModel.S(BaseModel.TableAfterSalesReverseLogisticsLinks)
}
