package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ProductSKUInventory captures realtime inventory snapshots per warehouse.
type ProductSKUInventory struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SKUId          string         `gorm:"column:sku_id;type:uuid;not null;uniqueIndex:idx_sku_wh_unique,priority:1" json:"sku_id"`
	WarehouseID    string         `gorm:"type:varchar(64);not null;uniqueIndex:idx_sku_wh_unique,priority:2" json:"warehouse_id"`
	AvailableQty   int64          `gorm:"not null;default:0" json:"available_qty"`
	LockedQty      int64          `gorm:"not null;default:0" json:"locked_qty"`
	InTransitQty   int64          `gorm:"not null;default:0" json:"in_transit_qty"`
	SafetyStock    int64          `gorm:"not null;default:0" json:"safety_stock"`
	AlertThreshold int64          `gorm:"not null;default:0" json:"alert_threshold"`
	LastSyncedAt   *time.Time     `json:"last_synced_at,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSKUInventory) TableName() string {
	return basemodels.S(basemodels.TableProductSkuInventories)
}
