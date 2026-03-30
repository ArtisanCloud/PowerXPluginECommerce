package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// KPISnapshot stores aggregated fulfillment KPIs by dimension and window.
type KPISnapshot struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_kpi_scope,priority:1" json:"tenant_uuid"`
	WindowHours         int            `gorm:"column:window_hours;not null;default:24;index:idx_logistics_kpi_scope,priority:2" json:"window_hours"`
	DimensionType       string         `gorm:"column:dimension_type;type:varchar(32);not null;index:idx_logistics_kpi_scope,priority:3" json:"dimension_type"`
	DimensionKey        string         `gorm:"column:dimension_key;type:varchar(128);not null;index:idx_logistics_kpi_scope,priority:4" json:"dimension_key"`
	TotalWaybills       int            `gorm:"column:total_waybills;not null;default:0" json:"total_waybills"`
	DeliveredCount      int            `gorm:"column:delivered_count;not null;default:0" json:"delivered_count"`
	ExceptionCount      int            `gorm:"column:exception_count;not null;default:0" json:"exception_count"`
	TimeoutCount        int            `gorm:"column:timeout_count;not null;default:0" json:"timeout_count"`
	OnTimeRate          float64        `gorm:"column:on_time_rate;type:numeric(7,2);not null;default:0" json:"on_time_rate"`
	DeliverySuccessRate float64        `gorm:"column:delivery_success_rate;type:numeric(7,2);not null;default:0" json:"delivery_success_rate"`
	AvgTransitHours     float64        `gorm:"column:avg_transit_hours;type:numeric(10,2);not null;default:0" json:"avg_transit_hours"`
	AvgCost             float64        `gorm:"column:avg_cost;type:numeric(18,6);not null;default:0" json:"avg_cost"`
	TotalCost           float64        `gorm:"column:total_cost;type:numeric(18,6);not null;default:0" json:"total_cost"`
	Metadata            datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KPISnapshot) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsKPISnapshots)
}
