package fulfillment

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Wave represents a grouped execution batch for fulfillment tasks.
type Wave struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	Name        string         `gorm:"column:name;type:varchar(128);not null" json:"name"`
	WarehouseID string         `gorm:"column:warehouse_id;type:uuid;not null;index" json:"warehouse_id"`
	Status      string         `gorm:"type:varchar(32);not null;default:'pending';index" json:"status"`
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Wave) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentWaves) }

// WaveTaskLink represents task membership and execution result inside a wave.
type WaveTaskLink struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_fulfillment_wave_task,priority:1" json:"tenant_uuid"`
	WaveID     string         `gorm:"column:wave_id;type:uuid;not null;index;uniqueIndex:uk_fulfillment_wave_task,priority:2" json:"wave_id"`
	TaskID     string         `gorm:"column:task_id;type:uuid;not null;index;uniqueIndex:uk_fulfillment_wave_task,priority:3" json:"task_id"`
	Result     string         `gorm:"type:varchar(32);not null;default:'pending'" json:"result"`
	Message    string         `gorm:"type:text" json:"message,omitempty"`
	Metadata   datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WaveTaskLink) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentWaveTaskLinks) }
