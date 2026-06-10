package coupon

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
)

// CouponUsageLog records immutable asset state actions.
type CouponUsageLog struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string    `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_coupon_usage_idempotency,priority:1" json:"tenant_uuid"`
	AssetID        string    `gorm:"column:asset_id;type:uuid;not null;index" json:"asset_id"`
	OrderID        *string   `gorm:"column:order_id;type:uuid;index" json:"order_id,omitempty"`
	Action         string    `gorm:"type:varchar(16);not null;index;uniqueIndex:uk_coupon_usage_idempotency,priority:2" json:"action"`
	ActionReason   string    `gorm:"column:action_reason;type:varchar(64);not null" json:"action_reason"`
	IdempotencyKey string    `gorm:"column:idempotency_key;type:varchar(128);not null;uniqueIndex:uk_coupon_usage_idempotency,priority:3" json:"idempotency_key"`
	RequestID      string    `gorm:"column:request_id;type:varchar(128)" json:"request_id,omitempty"`
	CreatedBy      string    `gorm:"column:created_by;type:varchar(128)" json:"created_by,omitempty"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (CouponUsageLog) TableName() string { return models.S(models.TableCouponUsageLogs) }
