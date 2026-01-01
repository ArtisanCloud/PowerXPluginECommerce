package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ProductSKUMedia stores SKU specific media assets for color/size variations.
type ProductSKUMedia struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SKUId       string         `gorm:"column:sku_id;type:uuid;not null;index" json:"sku_id"`
	MediaType   string         `gorm:"type:varchar(16);not null" json:"media_type"`
	URL         string         `gorm:"type:text;not null" json:"url"`
	IsPrimary   bool           `gorm:"not null;default:false" json:"is_primary"`
	ChannelCode string         `gorm:"type:varchar(64);comment:渠道覆盖" json:"channel_code,omitempty"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Metadata    string         `gorm:"type:text" json:"metadata,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSKUMedia) TableName() string {
	return basemodels.S(basemodels.TableProductSkuMedia)
}
