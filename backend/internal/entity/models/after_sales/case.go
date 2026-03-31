package after_sales

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// AfterSaleCase stores the primary after-sale application.
type AfterSaleCase struct {
	ID                   string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID           string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_after_sales_case_no,priority:1;index:idx_after_sales_case_item_status,priority:1" json:"tenant_uuid"`
	CaseNo               string         `gorm:"column:case_no;type:varchar(64);not null;uniqueIndex:uk_after_sales_case_no,priority:2" json:"case_no"`
	OrderID              string         `gorm:"column:order_id;type:uuid;not null;index" json:"order_id"`
	OrderItemID          string         `gorm:"column:order_item_id;type:uuid;not null;index;index:idx_after_sales_case_item_status,priority:2" json:"order_item_id"`
	CustomerID           string         `gorm:"column:customer_id;type:varchar(128);not null;index" json:"customer_id"`
	CaseType             string         `gorm:"column:case_type;type:varchar(32);not null;index" json:"case_type"`
	Status               string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index;index:idx_after_sales_case_item_status,priority:3" json:"status"`
	ReasonCode           string         `gorm:"column:reason_code;type:varchar(64)" json:"reason_code,omitempty"`
	ReasonDetail         string         `gorm:"column:reason_detail;type:text" json:"reason_detail,omitempty"`
	RequestedQty         int            `gorm:"column:requested_qty;not null;default:1" json:"requested_qty"`
	RequestedAmountMinor int64          `gorm:"column:requested_amount_minor;not null;default:0" json:"requested_amount_minor"`
	Currency             string         `gorm:"column:currency;type:varchar(16);not null;default:'CNY'" json:"currency"`
	SourceChannel        string         `gorm:"column:source_channel;type:varchar(32);not null;default:'miniapp'" json:"source_channel"`
	ClosedAt             *time.Time     `gorm:"column:closed_at;index" json:"closed_at,omitempty"`
	CreatedAt            time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AfterSaleCase) TableName() string {
	return BaseModel.S(BaseModel.TableAfterSalesCases)
}
