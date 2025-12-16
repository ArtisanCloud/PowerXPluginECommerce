package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelAlert captures credential/sync/KPI alerts.
type ChannelAlert struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"type:uuid;not null;index:idx_channel_alert_tenant,priority:1" json:"tenant_uuid"`
	ChannelID    string         `gorm:"type:uuid;not null;index:idx_channel_alert_tenant,priority:2" json:"channel_id"`
	Type         string         `gorm:"type:varchar(64);not null;index;comment:alert 类型" json:"type"`
	Severity     string         `gorm:"type:varchar(16);not null;default:'info'" json:"severity"`
	Title        string         `gorm:"type:varchar(255);not null" json:"title"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	Status       string         `gorm:"type:varchar(32);not null;default:'open';index" json:"status"`
	AssigneeUUID string         `gorm:"column:assignee_uuid;type:uuid" json:"assignee_uuid,omitempty"`
	TaskID       string         `gorm:"type:varchar(128);comment:关联任务" json:"task_id,omitempty"`
	TriggeredAt  time.Time      `gorm:"not null" json:"triggered_at"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelAlert) TableName() string {
	return models.S(models.TableChannelAlerts)
}
