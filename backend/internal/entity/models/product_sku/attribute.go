package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ProductSKUAttribute links SKU to individual spec values for quick filtering.
type ProductSKUAttribute struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SKUId        string         `gorm:"column:sku_id;type:uuid;not null;index:idx_sku_spec_unique,priority:1" json:"sku_id"`
	SpecID       string         `gorm:"type:uuid;not null;index:idx_sku_spec_unique,priority:2" json:"spec_id"`
	SpecValueID  string         `gorm:"type:uuid;not null;index:idx_sku_spec_unique,priority:3" json:"spec_value_id"`
	SpecName     string         `gorm:"type:varchar(128);comment:规格名称" json:"spec_name"`
	ValueName    string         `gorm:"type:varchar(128);comment:规格值" json:"value_name"`
	DisplayOrder int            `gorm:"type:int;default:0" json:"display_order"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSKUAttribute) TableName() string {
	return basemodels.S(basemodels.TableProductSkuAttributes)
}
