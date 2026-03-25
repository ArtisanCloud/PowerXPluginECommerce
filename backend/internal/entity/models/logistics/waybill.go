package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Waybill represents a forward logistics document linked to an order.
type Waybill struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_waybill_no,priority:1" json:"tenant_uuid"`
	OrderID     string         `gorm:"column:order_id;type:uuid;not null;index" json:"order_id"`
	CarrierID   string         `gorm:"column:carrier_id;type:uuid;not null;index" json:"carrier_id"`
	ServiceCode string         `gorm:"column:service_code;type:varchar(64);not null" json:"service_code"`
	WaybillNo   string         `gorm:"column:waybill_no;type:varchar(128);not null;uniqueIndex:uk_logistics_waybill_no,priority:2" json:"waybill_no"`
	Status      string         `gorm:"type:varchar(32);not null;default:'created';index" json:"status"`
	FeeAmount   float64        `gorm:"column:fee_amount;type:numeric(12,2);not null;default:0" json:"fee_amount"`
	LabelURL    string         `gorm:"column:label_url;type:text" json:"label_url,omitempty"`
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Waybill) TableName() string { return BaseModel.S(BaseModel.TableLogisticsWaybills) }

// TrackingEvent keeps immutable provider/manual tracking events.
type TrackingEvent struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_tracking_event,priority:1" json:"tenant_uuid"`
	WaybillID   string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	WaybillNo   string         `gorm:"column:waybill_no;type:varchar(128);not null;uniqueIndex:uk_logistics_tracking_event,priority:2" json:"waybill_no"`
	EventID     string         `gorm:"column:event_id;type:varchar(128);not null;uniqueIndex:uk_logistics_tracking_event,priority:3" json:"event_id"`
	Status      string         `gorm:"type:varchar(64);not null;index" json:"status"`
	Source      string         `gorm:"type:varchar(32);not null;default:'provider'" json:"source"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	OccurredAt  *time.Time     `gorm:"column:occurred_at" json:"occurred_at,omitempty"`
	Payload     datatypes.JSON `gorm:"type:jsonb" json:"payload,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TrackingEvent) TableName() string { return BaseModel.S(BaseModel.TableLogisticsTrackingEvents) }
