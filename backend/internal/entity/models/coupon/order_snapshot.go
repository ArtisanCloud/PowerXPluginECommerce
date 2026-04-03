package coupon

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OrderCouponSnapshot stores priced coupon data for an order.
type OrderCouponSnapshot struct {
	ID                 string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID         string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_order_coupon_snapshot,priority:1" json:"tenant_uuid"`
	OrderID            string         `gorm:"column:order_id;type:uuid;not null;index;uniqueIndex:uk_order_coupon_snapshot,priority:2" json:"order_id"`
	Currency           string         `gorm:"type:varchar(8);not null" json:"currency"`
	BaseTotalMinor     int64          `gorm:"column:base_total_minor;not null;default:0" json:"base_total_minor"`
	DiscountTotalMinor int64          `gorm:"column:discount_total_minor;not null;default:0" json:"discount_total_minor"`
	PayableTotalMinor  int64          `gorm:"column:payable_total_minor;not null;default:0" json:"payable_total_minor"`
	LineAllocations    datatypes.JSON `gorm:"column:line_allocations;type:jsonb;not null;default:'[]'" json:"line_allocations"`
	AppliedCoupons     datatypes.JSON `gorm:"column:applied_coupons;type:jsonb;not null;default:'[]'" json:"applied_coupons"`
	RejectedCoupons    datatypes.JSON `gorm:"column:rejected_coupons;type:jsonb;not null;default:'[]'" json:"rejected_coupons"`
	PricedAt           time.Time      `gorm:"column:priced_at;type:timestamptz;not null;index" json:"priced_at"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OrderCouponSnapshot) TableName() string { return models.S(models.TableOrderCouponSnapshots) }
