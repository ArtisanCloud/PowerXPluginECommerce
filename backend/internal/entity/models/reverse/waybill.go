package reverse

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Waybill models after-sale reverse shipment flow.
type Waybill struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_reverse_waybill_no,priority:1" json:"tenant_uuid"`
	OrderID          string         `gorm:"column:order_id;type:uuid;not null;index" json:"order_id"`
	AfterSaleID      string         `gorm:"column:after_sale_id;type:uuid;not null;index" json:"after_sale_id"`
	WaybillNo        string         `gorm:"column:waybill_no;type:varchar(128);not null;uniqueIndex:uk_reverse_waybill_no,priority:2" json:"waybill_no"`
	Status           string         `gorm:"type:varchar(32);not null;default:'created';index" json:"status"`
	InspectionResult string         `gorm:"column:inspection_result;type:varchar(32)" json:"inspection_result,omitempty"`
	Disposition      string         `gorm:"type:varchar(32)" json:"disposition,omitempty"`
	Metadata         datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Waybill) TableName() string { return BaseModel.S(BaseModel.TableReverseWaybills) }

// TrackingEvent records reverse-shipment trajectory updates.
type TrackingEvent struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	WaybillID   string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	Status      string         `gorm:"type:varchar(64);not null;index" json:"status"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	OccurredAt  *time.Time     `gorm:"column:occurred_at" json:"occurred_at,omitempty"`
	Payload     datatypes.JSON `gorm:"type:jsonb" json:"payload,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (TrackingEvent) TableName() string {
	return BaseModel.S(BaseModel.TableReverseWaybillTrackingEvents)
}

// WarehouseResult captures inbound inspection outcomes.
type WarehouseResult struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	WaybillID  string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	Result     string         `gorm:"type:varchar(64);not null" json:"result"`
	OperatorID string         `gorm:"column:operator_id;type:varchar(128)" json:"operator_id,omitempty"`
	Notes      string         `gorm:"type:text" json:"notes,omitempty"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WarehouseResult) TableName() string {
	return BaseModel.S(BaseModel.TableReverseWaybillWarehouseResults)
}
