package fulfillment

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WaveStrategy defines grouping rules for auto wave planning.
type WaveStrategy struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	Name            string         `gorm:"column:name;type:varchar(128);not null" json:"name"`
	WarehouseID     string         `gorm:"column:warehouse_id;type:varchar(128);index" json:"warehouse_id,omitempty"`
	CarrierCode     string         `gorm:"column:carrier_code;type:varchar(64);index" json:"carrier_code,omitempty"`
	TimeWindow      string         `gorm:"column:time_window;type:varchar(64)" json:"time_window,omitempty"`
	PriorityBand    string         `gorm:"column:priority_band;type:varchar(64)" json:"priority_band,omitempty"`
	MaxTasksPerWave int            `gorm:"column:max_tasks_per_wave;type:int;not null;default:50" json:"max_tasks_per_wave"`
	Enabled         bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Rules           datatypes.JSON `gorm:"type:jsonb;not null" json:"rules"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WaveStrategy) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentWaveStrategies) }
