package after_sales

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
)

// AfterSaleTimeline records immutable actions and status transitions.
type AfterSaleTimeline struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string    `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CaseID       string    `gorm:"column:case_id;type:uuid;not null;index" json:"case_id"`
	Action       string    `gorm:"column:action;type:varchar(32);not null;index" json:"action"`
	FromStatus   string    `gorm:"column:from_status;type:varchar(32)" json:"from_status,omitempty"`
	ToStatus     string    `gorm:"column:to_status;type:varchar(32);not null;index" json:"to_status"`
	OperatorType string    `gorm:"column:operator_type;type:varchar(32);not null" json:"operator_type"`
	OperatorID   string    `gorm:"column:operator_id;type:varchar(128)" json:"operator_id,omitempty"`
	Note         string    `gorm:"column:note;type:text" json:"note,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AfterSaleTimeline) TableName() string {
	return BaseModel.S(BaseModel.TableAfterSalesTimelines)
}
