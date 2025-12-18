package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelSyncHistory captures manual/automated sync runs per channel.
type ChannelSyncHistory struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"type:uuid;not null;index:idx_channel_sync_history_tenant,priority:1" json:"tenant_uuid"`
	ChannelID   string         `gorm:"type:uuid;not null;index:idx_channel_sync_history_tenant,priority:2" json:"channel_id"`
	TriggerType string         `gorm:"type:varchar(32);not null;comment:manual/scheduled/retry" json:"trigger_type"`
	TriggeredBy string         `gorm:"type:varchar(64);comment:触发者" json:"triggered_by"`
	DurationMs  int64          `gorm:"type:bigint;comment:耗时毫秒" json:"duration_ms"`
	Result      string         `gorm:"type:varchar(32);not null;comment:success/failed/partial" json:"result"`
	Payload     datatypes.JSON `gorm:"type:jsonb;comment:任务摘要" json:"payload"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelSyncHistory) TableName() string {
	return models.S(models.TableChannelSyncHistory)
}
