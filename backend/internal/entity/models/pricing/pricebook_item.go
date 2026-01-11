package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PricebookItem stores SKU pricing fields under a specific version.
type PricebookItem struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:条目ID" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_pricebook_item_version_sku,priority:1;comment:租户" json:"tenant_uuid"`
	PricebookID string         `gorm:"column:pricebook_id;type:uuid;not null;index;comment:价目表ID" json:"pricebook_id"`
	VersionID   string         `gorm:"column:version_id;type:uuid;not null;index;uniqueIndex:uk_pricebook_item_version_sku,priority:2;comment:版本ID" json:"version_id"`
	SKUID       string         `gorm:"column:sku_id;type:uuid;not null;index;uniqueIndex:uk_pricebook_item_version_sku,priority:3;comment:SKU ID" json:"sku_id"`
	BaseAmount  *int64         `gorm:"column:base_amount_minor;type:bigint;comment:基准价(最小货币单位)" json:"base_amount_minor,omitempty"`
	SaleAmount  *int64         `gorm:"column:sale_amount_minor;type:bigint;comment:销售价(最小货币单位)" json:"sale_amount_minor,omitempty"`
	MsrpAmount  *int64         `gorm:"column:msrp_amount_minor;type:bigint;comment:建议零售价(最小货币单位)" json:"msrp_amount_minor,omitempty"`
	CostAmount  *int64         `gorm:"column:cost_amount_minor;type:bigint;comment:成本价(最小货币单位)" json:"cost_amount_minor,omitempty"`
	MinAmount   *int64         `gorm:"column:min_amount_minor;type:bigint;comment:最低价(最小货币单位)" json:"min_amount_minor,omitempty"`
	MaxAmount   *int64         `gorm:"column:max_amount_minor;type:bigint;comment:最高价(最小货币单位)" json:"max_amount_minor,omitempty"`
	TaxIncluded bool           `gorm:"column:tax_included;type:boolean;not null;default:false;comment:含税" json:"tax_included"`
	Meta        datatypes.JSON `gorm:"column:meta;type:jsonb;comment:扩展字段" json:"meta,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName implements gorm.Tabler.
func (PricebookItem) TableName() string { return models.S(models.TablePricebookItems) }
