package order

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
)

// OrderItem is an immutable line item snapshot for a given order.
type OrderItem struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_order_item_tenant_order_sku,priority:1" json:"tenant_uuid"`
	OrderID     string    `gorm:"column:order_id;type:uuid;not null;index;uniqueIndex:idx_order_item_tenant_order_sku,priority:2" json:"order_id"`
	SKUID       string    `gorm:"column:sku_id;type:uuid;not null;index;uniqueIndex:idx_order_item_tenant_order_sku,priority:3" json:"sku_id"`
	Qty         int64     `gorm:"not null" json:"qty"`
	UnitPrice   int64     `gorm:"not null" json:"unit_price"`
	LineAmount  int64     `gorm:"not null" json:"line_amount"`
	PriceSource string    `gorm:"type:varchar(64)" json:"price_source,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderItem) TableName() string {
	return basemodels.S(basemodels.TableOrderItems)
}
