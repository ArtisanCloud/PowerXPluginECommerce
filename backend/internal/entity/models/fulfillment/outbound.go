package fulfillment

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Outbound struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_fulfillment_outbound_task,priority:1" json:"tenant_uuid"`
	TaskID      string         `gorm:"column:task_id;type:uuid;not null;index;uniqueIndex:uk_fulfillment_outbound_task,priority:2" json:"task_id"`
	OrderID     string         `gorm:"column:order_id;type:uuid;not null;index" json:"order_id"`
	WarehouseID string         `gorm:"column:warehouse_id;type:uuid;not null;index" json:"warehouse_id"`
	WaybillID   string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	Status      string         `gorm:"column:status;type:varchar(32);not null;default:'reserved';index" json:"status"`
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Outbound) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentOutbounds) }

type PickItem struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_fulfillment_pick_line,priority:1" json:"tenant_uuid"`
	OutboundID string         `gorm:"column:outbound_id;type:uuid;not null;index;uniqueIndex:uk_fulfillment_pick_line,priority:2" json:"outbound_id"`
	SKU        string         `gorm:"column:sku;type:varchar(128);not null;uniqueIndex:uk_fulfillment_pick_line,priority:3" json:"sku"`
	Qty        int            `gorm:"column:qty;not null;default:1" json:"qty"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PickItem) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentPickItems) }

type PackOrder struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_fulfillment_pack_outbound,priority:1" json:"tenant_uuid"`
	OutboundID string         `gorm:"column:outbound_id;type:uuid;not null;index;uniqueIndex:uk_fulfillment_pack_outbound,priority:2" json:"outbound_id"`
	PackageNo  int            `gorm:"column:package_no;not null;default:1" json:"package_no"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:'packed';index" json:"status"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PackOrder) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentPackOrders) }
