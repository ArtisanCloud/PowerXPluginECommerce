package product_sku

import (
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ProductSKUAuditLog captures immutable events for SKU operations (including inventory adjustments).
type ProductSKUAuditLog struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SKUId            string         `gorm:"column:sku_id;type:uuid;not null;index" json:"sku_id"`
	WarehouseID      string         `gorm:"type:varchar(64);not null" json:"warehouse_id"`
	Action           string         `gorm:"type:text;not null" json:"action"`
	Actor            string         `gorm:"type:text" json:"actor,omitempty"`
	Delta            int64          `gorm:"not null" json:"delta"`
	BeforeAvailable  int64          `gorm:"column:before_available_qty;not null" json:"before_available_qty"`
	AfterAvailable   int64          `gorm:"column:after_available_qty;not null" json:"after_available_qty"`
	RequestID        string         `gorm:"type:text" json:"request_id,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSKUAuditLog) TableName() string { return models.S(models.TableProductSkuAuditLogs) }
