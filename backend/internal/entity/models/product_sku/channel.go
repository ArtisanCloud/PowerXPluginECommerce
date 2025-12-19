package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
)

// ProductSKUChannel stores per-channel SKU identifiers and publish metadata.
type ProductSKUChannel struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SKUId         string         `gorm:"column:sku_id;type:uuid;not null;index:idx_sku_channel_unique,priority:1" json:"sku_id"`
	ChannelCode   string         `gorm:"type:varchar(64);not null;index:idx_sku_channel_unique,priority:2" json:"channel_code"`
	ChannelSkuID  string         `gorm:"type:varchar(120);not null;comment:渠道 SKU" json:"channel_sku_id"`
	Status        string         `gorm:"type:varchar(32);not null;default:'pending';index" json:"status"`
	SyncMode      string         `gorm:"type:varchar(16);default:'push'" json:"sync_mode"`
	PublishTime   *time.Time     `json:"publish_time,omitempty"`
	PublishTaskID *string        `gorm:"type:varchar(64);comment:任务 ID" json:"publish_task_id,omitempty"`
	LastError     string         `gorm:"type:text" json:"last_error,omitempty"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProductSKUChannel) TableName() string {
	return basemodels.S(basemodels.TableProductSkuChannels)
}
