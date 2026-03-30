package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RoutingRule defines warehouse-carrier routing decisions for fulfillment.
type RoutingRule struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_routing_rule_name,priority:1" json:"tenant_uuid"`
	Name            string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_routing_rule_name,priority:2" json:"name"`
	WarehouseID     string         `gorm:"column:warehouse_id;type:varchar(128);not null;default:'';index" json:"warehouse_id,omitempty"`
	DestinationZone string         `gorm:"column:destination_zone;type:varchar(64);not null;default:'GLOBAL';index" json:"destination_zone,omitempty"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;not null;index" json:"carrier_id"`
	ServiceCode     string         `gorm:"column:service_code;type:varchar(64);not null;default:'std'" json:"service_code"`
	Priority        int            `gorm:"column:priority;not null;default:100;index" json:"priority"`
	MinWeight       float64        `gorm:"column:min_weight;type:numeric(12,3);not null;default:0" json:"min_weight"`
	MaxWeight       float64        `gorm:"column:max_weight;type:numeric(12,3);not null;default:0" json:"max_weight"`
	MinOrderAmount  float64        `gorm:"column:min_order_amount;type:numeric(12,2);not null;default:0" json:"min_order_amount"`
	MaxOrderAmount  float64        `gorm:"column:max_order_amount;type:numeric(12,2);not null;default:0" json:"max_order_amount"`
	Fallback        bool           `gorm:"column:fallback;not null;default:false" json:"fallback"`
	Enabled         bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	RuleConfig      datatypes.JSON `gorm:"column:rule_config;type:jsonb" json:"rule_config,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RoutingRule) TableName() string { return BaseModel.S(BaseModel.TableLogisticsRoutingRules) }
