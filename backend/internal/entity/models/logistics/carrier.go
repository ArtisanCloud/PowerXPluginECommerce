package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Carrier describes a tenant scoped logistics provider.
type Carrier struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_carrier_name,priority:1" json:"tenant_uuid"`
	Name         string         `gorm:"type:varchar(128);not null;uniqueIndex:uk_logistics_carrier_name,priority:2" json:"name"`
	Code         string         `gorm:"type:varchar(64);not null;index" json:"code"`
	Type         string         `gorm:"type:varchar(32);not null;default:'self'" json:"type"`
	Status       string         `gorm:"type:varchar(32);not null;default:'active';index" json:"status"`
	ContactName  string         `gorm:"type:varchar(120)" json:"contact_name,omitempty"`
	ContactPhone string         `gorm:"type:varchar(64)" json:"contact_phone,omitempty"`
	Capabilities datatypes.JSON `gorm:"type:jsonb" json:"capabilities,omitempty"`
	Config       datatypes.JSON `gorm:"type:jsonb" json:"config,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Carrier) TableName() string { return BaseModel.S(BaseModel.TableLogisticsCarriers) }

// CarrierService captures service-level capabilities for one carrier.
type CarrierService struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_carrier_service,priority:1" json:"tenant_uuid"`
	CarrierID   string         `gorm:"column:carrier_id;type:uuid;not null;index;uniqueIndex:uk_logistics_carrier_service,priority:2" json:"carrier_id"`
	Code        string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_logistics_carrier_service,priority:3" json:"code"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	ServiceType string         `gorm:"column:service_type;type:varchar(32);not null;default:'express'" json:"service_type"`
	Status      string         `gorm:"type:varchar(32);not null;default:'active';index" json:"status"`
	Config      datatypes.JSON `gorm:"type:jsonb" json:"config,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CarrierService) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCarrierServices)
}
