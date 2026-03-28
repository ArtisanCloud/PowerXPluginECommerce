package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CapacityPlan defines per-carrier reservable capacity for smart allocation.
type CapacityPlan struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_capacity_plan,priority:1" json:"tenant_uuid"`
	Name             string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_capacity_plan,priority:2" json:"name"`
	CarrierID        string         `gorm:"column:carrier_id;type:uuid;not null;index" json:"carrier_id"`
	WarehouseID      string         `gorm:"column:warehouse_id;type:varchar(64);index" json:"warehouse_id,omitempty"`
	DestinationZone  string         `gorm:"column:destination_zone;type:varchar(64);index" json:"destination_zone,omitempty"`
	DailyCapacity    int            `gorm:"column:daily_capacity;not null;default:0" json:"daily_capacity"`
	ReservedCapacity int            `gorm:"column:reserved_capacity;not null;default:0" json:"reserved_capacity"`
	UsedCapacity     int            `gorm:"column:used_capacity;not null;default:0" json:"used_capacity"`
	Status           string         `gorm:"column:status;type:varchar(32);not null;default:'active';index" json:"status"`
	Config           datatypes.JSON `gorm:"column:config;type:jsonb" json:"config,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CapacityPlan) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCapacityPlans)
}
