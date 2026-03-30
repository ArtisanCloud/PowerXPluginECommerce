package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ETAPolicy defines promised pickup/delivery windows by carrier service and destination zone.
type ETAPolicy struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_eta_policy,priority:1" json:"tenant_uuid"`
	CarrierID        string         `gorm:"column:carrier_id;type:uuid;not null;index;uniqueIndex:uk_logistics_eta_policy,priority:2" json:"carrier_id"`
	ServiceCode      string         `gorm:"column:service_code;type:varchar(64);not null;uniqueIndex:uk_logistics_eta_policy,priority:3" json:"service_code"`
	DestinationZone  string         `gorm:"column:destination_zone;type:varchar(64);not null;default:'GLOBAL';uniqueIndex:uk_logistics_eta_policy,priority:4" json:"destination_zone"`
	Timezone         string         `gorm:"column:timezone;type:varchar(64);not null;default:'UTC'" json:"timezone"`
	CutoffHourLocal  int            `gorm:"column:cutoff_hour_local;not null;default:18" json:"cutoff_hour_local"`
	PickupSLAHours   int            `gorm:"column:pickup_sla_hours;not null;default:24" json:"pickup_sla_hours"`
	DeliverySLAHours int            `gorm:"column:delivery_sla_hours;not null;default:72" json:"delivery_sla_hours"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ETAPolicy) TableName() string { return BaseModel.S(BaseModel.TableLogisticsETAPolicies) }

// ETARecord stores computed promised/estimated timestamps per waybill.
type ETARecord struct {
	ID                 string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID         string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_eta_waybill,priority:1" json:"tenant_uuid"`
	WaybillID          string         `gorm:"column:waybill_id;type:uuid;not null;index;uniqueIndex:uk_logistics_eta_waybill,priority:2" json:"waybill_id"`
	WaybillNo          string         `gorm:"column:waybill_no;type:varchar(128);not null;index" json:"waybill_no"`
	CarrierID          string         `gorm:"column:carrier_id;type:uuid;not null;index" json:"carrier_id"`
	ServiceCode        string         `gorm:"column:service_code;type:varchar(64);not null" json:"service_code"`
	Timezone           string         `gorm:"column:timezone;type:varchar(64);not null;default:'UTC'" json:"timezone"`
	PickupDeadlineAt   *time.Time     `gorm:"column:pickup_deadline_at" json:"pickup_deadline_at,omitempty"`
	DeliveryDeadlineAt *time.Time     `gorm:"column:delivery_deadline_at" json:"delivery_deadline_at,omitempty"`
	PromisedAt         *time.Time     `gorm:"column:promised_at" json:"promised_at,omitempty"`
	EstimatedAt        *time.Time     `gorm:"column:estimated_at" json:"estimated_at,omitempty"`
	Source             string         `gorm:"column:source;type:varchar(32);not null;default:'calculated'" json:"source"`
	Version            int            `gorm:"column:version;not null;default:1" json:"version"`
	Metadata           datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	LastComputedAt     time.Time      `gorm:"column:last_computed_at;not null" json:"last_computed_at"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ETARecord) TableName() string { return BaseModel.S(BaseModel.TableLogisticsETARecords) }
