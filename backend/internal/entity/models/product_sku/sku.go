package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProductSKU captures the tenant scoped SKU definition with default logistics and pricing metadata.
type ProductSKU struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:SKU 主键" json:"id"`
	TenantUUID     string         `gorm:"type:uuid;not null;index:idx_product_sku_tenant_code,priority:1;comment:租户" json:"tenant_uuid"`
	SPUID          string         `gorm:"column:spu_id;type:uuid;not null;index;comment:所属 SPU" json:"spu_id"`
	SKUCode        string         `gorm:"column:sku_code;type:varchar(120);not null;index:idx_product_sku_tenant_code,priority:2;comment:SKU 业务编码" json:"sku_code"`
	Barcode        string         `gorm:"type:varchar(64);index;comment:条码" json:"barcode,omitempty"`
	Status         string         `gorm:"type:varchar(32);not null;default:'draft';index;comment:状态" json:"status"`
	LifecyclePhase string         `gorm:"type:varchar(32);default:'concept';comment:生命周期" json:"lifecycle_phase,omitempty"`
	MinOrderQty    int            `gorm:"comment:起订量" json:"min_order_qty,omitempty"`
	SpecValues     datatypes.JSON `gorm:"type:jsonb;comment:规格组合" json:"spec_values,omitempty"`
	DefaultValues  datatypes.JSON `gorm:"type:jsonb;comment:默认字段快照" json:"default_values,omitempty"`
	Logistics      datatypes.JSON `gorm:"type:jsonb;comment:物流属性" json:"logistics,omitempty"`
	PriceRefs      datatypes.JSON `gorm:"type:jsonb;comment:价目指针" json:"price_refs,omitempty"`
	Tags           pq.StringArray `gorm:"type:text[];comment:标签" json:"tags,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSKU) TableName() string { return basemodels.S(basemodels.TableProductSkus) }
