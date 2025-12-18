package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ChannelTaskLink stores associations between a channel and remediation tasks.
type ChannelTaskLink struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index:idx_channel_task_link_tenant,priority:1" json:"tenant_uuid"`
	ChannelID  string         `gorm:"type:uuid;not null;index:idx_channel_task_link_tenant,priority:2" json:"channel_id"`
	TaskID     string         `gorm:"type:varchar(128);not null;index;comment:任务中心ID" json:"task_id"`
	TaskSource string         `gorm:"type:varchar(32);not null;comment:任务来源" json:"task_source"`
	Status     string         `gorm:"type:varchar(32);not null;default:'open'" json:"status"`
	Note       string         `gorm:"type:text" json:"note,omitempty"`
	LinkedBy   string         `gorm:"type:varchar(64);not null" json:"linked_by"`
	LinkedAt   time.Time      `gorm:"autoCreateTime" json:"linked_at"`
	ResolvedAt *time.Time     `gorm:"comment:解除时间" json:"resolved_at,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelTaskLink) TableName() string {
	return models.S(models.TableChannelTaskLinks)
}

// ChannelNote persists tenant-scoped operator notes for a channel.
type ChannelNote struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index:idx_channel_notes_tenant,priority:1" json:"tenant_uuid"`
	ChannelID  string         `gorm:"type:uuid;not null;index:idx_channel_notes_tenant,priority:2" json:"channel_id"`
	AuthorUUID string         `gorm:"type:varchar(64);not null" json:"author_uuid"`
	Visibility string         `gorm:"type:varchar(32);not null;default:'team'" json:"visibility"`
	Body       string         `gorm:"type:text;not null" json:"body"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChannelNote) TableName() string {
	return models.S(models.TableChannelNotes)
}
