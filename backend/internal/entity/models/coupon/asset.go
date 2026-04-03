package coupon

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CouponAsset is a user-owned coupon instance.
type CouponAsset struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_coupon_asset_code,priority:1" json:"tenant_uuid"`
	TemplateID      string         `gorm:"column:template_id;type:uuid;not null;index" json:"template_id"`
	UserID          string         `gorm:"column:user_id;type:varchar(64);not null;index:idx_coupon_asset_user_status,priority:1" json:"user_id"`
	CouponCode      string         `gorm:"column:coupon_code;type:varchar(64);not null;uniqueIndex:uk_coupon_asset_code,priority:2" json:"coupon_code"`
	Status          string         `gorm:"type:varchar(16);not null;default:'available';index;index:idx_coupon_asset_user_status,priority:2" json:"status"`
	ReservedOrderID *string        `gorm:"column:reserved_order_id;type:uuid;index" json:"reserved_order_id,omitempty"`
	ReservedAt      *time.Time     `gorm:"column:reserved_at;type:timestamptz" json:"reserved_at,omitempty"`
	RedeemedAt      *time.Time     `gorm:"column:redeemed_at;type:timestamptz" json:"redeemed_at,omitempty"`
	RefundedAt      *time.Time     `gorm:"column:refunded_at;type:timestamptz" json:"refunded_at,omitempty"`
	ExpiredAt       *time.Time     `gorm:"column:expired_at;type:timestamptz" json:"expired_at,omitempty"`
	ValidFrom       *time.Time     `gorm:"column:valid_from;type:timestamptz;index" json:"valid_from,omitempty"`
	ValidTo         *time.Time     `gorm:"column:valid_to;type:timestamptz;index:idx_coupon_asset_user_status,priority:3" json:"valid_to,omitempty"`
	Meta            datatypes.JSON `gorm:"column:meta;type:jsonb;not null;default:'{}'" json:"meta,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CouponAsset) TableName() string { return models.S(models.TableCouponAssets) }
