package cart

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Cart represents a customer shopping cart snapshot.
// Hybrid mode: client keeps a local cart and syncs/merges with this server-side cart after login.
type Cart struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index;uniqueIndex:idx_cart_tenant_customer,priority:1" json:"tenant_uuid"`
	CustomerID string         `gorm:"column:customer_id;type:uuid;not null;index;uniqueIndex:idx_cart_tenant_customer,priority:2" json:"customer_id"`
	Items      datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"items"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Cart) TableName() string {
	return basemodels.S(basemodels.TableCarts)
}
