package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelMaster captures tenant-scoped channel/store master data.
type ChannelMaster struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:渠道ID" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;uniqueIndex:idx_channel_master_tenant_store,priority:1;comment:租户" json:"tenant_uuid"`
	Platform     string         `gorm:"type:varchar(64);not null;index;comment:平台" json:"platform"`
	ChannelType  string         `gorm:"type:varchar(32);not null;index;comment:渠道类型" json:"channel_type"`
	StoreID      string         `gorm:"column:store_id;type:varchar(128);not null;uniqueIndex:idx_channel_master_tenant_store,priority:2;comment:店铺/渠道ID" json:"store_id"`
	Name         string         `gorm:"type:varchar(255);not null;comment:渠道名称" json:"name"`
	Domain       string         `gorm:"type:varchar(255);comment:域名" json:"domain,omitempty"`
	Region       string         `gorm:"type:varchar(64);not null;index;comment:区域" json:"region"`
	Status       string         `gorm:"type:varchar(32);not null;index;comment:状态" json:"status"`
	Tags         pq.StringArray `gorm:"type:text[];comment:标签" json:"tags"`
	OwnerUUID    string         `gorm:"type:varchar(64);not null;index;comment:负责人" json:"owner_uuid"`
	ApproverUUID *string        `gorm:"type:varchar(64);comment:审批人" json:"approver_uuid,omitempty"`
	ContactName  string         `gorm:"type:varchar(120);comment:联系人" json:"contact_name,omitempty"`
	ContactPhone string         `gorm:"type:varchar(64);comment:联系电话" json:"contact_phone,omitempty"`
	ContactEmail string         `gorm:"type:varchar(128);comment:联系邮箱" json:"contact_email,omitempty"`
	HealthScore  *int           `gorm:"column:health_score;type:int;comment:健康度" json:"health_score,omitempty"`
	LastSyncAt   *time.Time     `gorm:"column:last_sync_at;comment:最后同步时间" json:"last_sync_at,omitempty"`
	SyncStatus   string         `gorm:"column:sync_status;type:varchar(32);not null;default:'unknown';comment:同步状态" json:"sync_status"`
	CreatedBy    string         `gorm:"column:created_by;type:varchar(64);comment:创建人" json:"created_by,omitempty"`
	UpdatedBy    string         `gorm:"column:updated_by;type:varchar(64);comment:更新人" json:"updated_by,omitempty"`
	Metadata     datatypes.JSON `gorm:"column:metadata;type:jsonb;comment:附加配置" json:"metadata,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName implements gorm.Tabler.
func (ChannelMaster) TableName() string {
	return models.S(models.TableChannelMasters)
}
