package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TrackingSyncSchedule defines tenant-owned recurring tracking sync policy.
type TrackingSyncSchedule struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_sync_schedule_name,priority:1" json:"tenant_uuid"`
	Name            string         `gorm:"column:name;type:varchar(120);not null;uniqueIndex:uk_logistics_sync_schedule_name,priority:2" json:"name"`
	CronExpr        string         `gorm:"column:cron_expr;type:varchar(128);not null" json:"cron_expr"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	WaybillStatus   string         `gorm:"column:waybill_status;type:varchar(32);index" json:"waybill_status,omitempty"`
	Enabled         bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	MaxConcurrency  int            `gorm:"column:max_concurrency;not null;default:1" json:"max_concurrency"`
	DedupeWindowSec int            `gorm:"column:dedupe_window_sec;not null;default:300" json:"dedupe_window_sec"`
	BatchLimit      int            `gorm:"column:batch_limit;not null;default:20" json:"batch_limit"`
	EventLimit      int            `gorm:"column:event_limit;not null;default:20" json:"event_limit"`
	LastTriggeredAt *time.Time     `gorm:"column:last_triggered_at" json:"last_triggered_at,omitempty"`
	NextTriggerAt   *time.Time     `gorm:"column:next_trigger_at;index" json:"next_trigger_at,omitempty"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TrackingSyncSchedule) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsTrackingSyncSchedules)
}
