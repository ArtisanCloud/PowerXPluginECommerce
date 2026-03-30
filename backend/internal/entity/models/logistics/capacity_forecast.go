package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CapacityForecast stores forecast results and quota recommendations for capacity governance.
type CapacityForecast struct {
	ID                   string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID           string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_capacity_forecast_scope,priority:1" json:"tenant_uuid"`
	PlanID               string         `gorm:"column:plan_id;type:uuid;index" json:"plan_id,omitempty"`
	CarrierID            string         `gorm:"column:carrier_id;type:uuid;index;index:idx_logistics_capacity_forecast_scope,priority:2" json:"carrier_id,omitempty"`
	WarehouseID          string         `gorm:"column:warehouse_id;type:varchar(64);index;index:idx_logistics_capacity_forecast_scope,priority:3" json:"warehouse_id,omitempty"`
	DestinationZone      string         `gorm:"column:destination_zone;type:varchar(64);index;index:idx_logistics_capacity_forecast_scope,priority:4" json:"destination_zone,omitempty"`
	WindowDays           int            `gorm:"column:window_days;not null;default:7;index" json:"window_days"`
	CurrentDailyCapacity int            `gorm:"column:current_daily_capacity;not null;default:0" json:"current_daily_capacity"`
	PredictedDailyVolume int            `gorm:"column:predicted_daily_volume;not null;default:0" json:"predicted_daily_volume"`
	TargetCapacity       int            `gorm:"column:target_capacity;not null;default:0" json:"target_capacity"`
	RecommendedQuota     int            `gorm:"column:recommended_quota;not null;default:0" json:"recommended_quota"`
	Confidence           float64        `gorm:"column:confidence;type:numeric(7,2);not null;default:0" json:"confidence"`
	RiskLevel            string         `gorm:"column:risk_level;type:varchar(32);not null;default:'low';index" json:"risk_level"`
	Strategy             string         `gorm:"column:strategy;type:varchar(32);not null;default:'keep_quota';index" json:"strategy"`
	Status               string         `gorm:"column:status;type:varchar(32);not null;default:'suggested';index" json:"status"`
	Metrics              datatypes.JSON `gorm:"column:metrics;type:jsonb" json:"metrics,omitempty"`
	AppliedAt            *time.Time     `gorm:"column:applied_at" json:"applied_at,omitempty"`
	CreatedBy            string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	UpdatedBy            string         `gorm:"column:updated_by;type:varchar(64)" json:"updated_by,omitempty"`
	CreatedAt            time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt            time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CapacityForecast) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCapacityForecasts)
}
