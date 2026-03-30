package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ControlTowerSnapshot stores materialized metrics for logistics control tower views.
type ControlTowerSnapshot struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_control_tower_scope,priority:1" json:"tenant_uuid"`
	WindowHours     int            `gorm:"column:window_hours;not null;default:24;index:idx_logistics_control_tower_scope,priority:2" json:"window_hours"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index:idx_logistics_control_tower_scope,priority:3" json:"carrier_id,omitempty"`
	WarehouseID     string         `gorm:"column:warehouse_id;type:varchar(64);index:idx_logistics_control_tower_scope,priority:4" json:"warehouse_id,omitempty"`
	DestinationZone string         `gorm:"column:destination_zone;type:varchar(64);index:idx_logistics_control_tower_scope,priority:5" json:"destination_zone,omitempty"`
	InTransitCount  int            `gorm:"column:in_transit_count;not null;default:0" json:"in_transit_count"`
	ExceptionCount  int            `gorm:"column:exception_count;not null;default:0" json:"exception_count"`
	TimeoutCount    int            `gorm:"column:timeout_count;not null;default:0" json:"timeout_count"`
	DeliveredCount  int            `gorm:"column:delivered_count;not null;default:0" json:"delivered_count"`
	OnTimeRate      float64        `gorm:"column:on_time_rate;type:numeric(7,2);not null;default:0" json:"on_time_rate"`
	TotalCost       float64        `gorm:"column:total_cost;type:numeric(18,6);not null;default:0" json:"total_cost"`
	AlertCount      int            `gorm:"column:alert_count;not null;default:0" json:"alert_count"`
	Summary         datatypes.JSON `gorm:"column:summary;type:jsonb" json:"summary,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ControlTowerSnapshot) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsControlTowerSnapshots)
}

// ControlTowerAlertSubscription stores per-tenant alert watch configuration.
type ControlTowerAlertSubscription struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_control_tower_subscription,priority:1" json:"tenant_uuid"`
	Name            string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_control_tower_subscription,priority:2" json:"name"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	WarehouseID     string         `gorm:"column:warehouse_id;type:varchar(64);index" json:"warehouse_id,omitempty"`
	DestinationZone string         `gorm:"column:destination_zone;type:varchar(64);index" json:"destination_zone,omitempty"`
	MinOnTimeRate   float64        `gorm:"column:min_on_time_rate;type:numeric(7,2);not null;default:95" json:"min_on_time_rate"`
	MaxTimeoutCount int            `gorm:"column:max_timeout_count;not null;default:5" json:"max_timeout_count"`
	MaxCostAmount   float64        `gorm:"column:max_cost_amount;type:numeric(18,6);not null;default:0" json:"max_cost_amount"`
	Enabled         bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Config          datatypes.JSON `gorm:"column:config;type:jsonb" json:"config,omitempty"`
	LastNotifiedAt  *time.Time     `gorm:"column:last_notified_at" json:"last_notified_at,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ControlTowerAlertSubscription) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsControlTowerAlertSubscriptions)
}
