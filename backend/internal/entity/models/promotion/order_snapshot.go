package promotion

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OrderSnapshot stores the promotion calculation result at order creation time.
type OrderSnapshot struct {
	ID                       string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID               string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_order_promotion_snapshot,priority:1" json:"tenant_uuid"`
	OrderID                  string         `gorm:"column:order_id;type:uuid;not null;index;uniqueIndex:uk_order_promotion_snapshot,priority:2" json:"order_id"`
	Currency                 string         `gorm:"type:varchar(8);not null" json:"currency"`
	BaseTotalMinor           int64          `gorm:"column:base_total_minor;not null;default:0" json:"base_total_minor"`
	PromotionDiscountMinor   int64          `gorm:"column:promotion_discount_minor;not null;default:0" json:"promotion_discount_minor"`
	AfterPromotionTotalMinor int64          `gorm:"column:after_promotion_total_minor;not null;default:0" json:"after_promotion_total_minor"`
	AppliedPromotions        datatypes.JSON `gorm:"column:applied_promotions;type:jsonb;not null;default:'[]'" json:"applied_promotions"`
	RejectedPromotions       datatypes.JSON `gorm:"column:rejected_promotions;type:jsonb;not null;default:'[]'" json:"rejected_promotions"`
	LineAllocations          datatypes.JSON `gorm:"column:line_allocations;type:jsonb;not null;default:'[]'" json:"line_allocations"`
	PricedAt                 time.Time      `gorm:"column:priced_at;type:timestamptz;not null;index" json:"priced_at"`
	CreatedAt                time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OrderSnapshot) TableName() string { return models.S(models.TableOrderPromotionSnapshots) }
