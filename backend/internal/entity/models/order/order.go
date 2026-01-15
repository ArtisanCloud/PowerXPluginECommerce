package order

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Order represents a first-party order created by mini-app or admin.
type Order struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"type:uuid;not null;index;uniqueIndex:idx_order_tenant_order_no,priority:1" json:"tenant_uuid"`
	OrderNo             string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_order_tenant_order_no,priority:2;index" json:"order_no"`
	CustomerID          string         `gorm:"column:customer_id;type:uuid;not null;index" json:"customer_id"`
	Channel             string         `gorm:"type:varchar(64);not null;index" json:"channel"`
	Status              string         `gorm:"type:varchar(32);not null;index" json:"status"`
	Currency            string         `gorm:"type:varchar(8);not null" json:"currency"`
	SubtotalAmount      int64          `gorm:"not null;default:0" json:"subtotal_amount"`
	TotalAmount         int64          `gorm:"not null;default:0" json:"total_amount"`
	ShippingAddressID   string         `gorm:"column:shipping_address_id;type:uuid;index" json:"shipping_address_id,omitempty"`
	ShippingAddressSnap datatypes.JSON `gorm:"column:shipping_address_snapshot;type:jsonb" json:"shipping_address_snapshot,omitempty"`
	PriceSnapshot       datatypes.JSON `gorm:"type:jsonb" json:"price_snapshot,omitempty"`
	SellabilitySnap     datatypes.JSON `gorm:"column:sellability_snapshot;type:jsonb" json:"sellability_snapshot,omitempty"`
	CreatedByType       string         `gorm:"type:varchar(16);not null;default:'system'" json:"created_by_type"`
	CreatedBy           string         `gorm:"type:varchar(128)" json:"created_by,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Order) TableName() string {
	return basemodels.S(basemodels.TableOrders)
}
