package coupon

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// UsageLog is a phase-1 skeleton for coupon action logs.
type UsageLog struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UsageLog) TableName() string { return models.S(models.TableCouponUsageLogs) }
