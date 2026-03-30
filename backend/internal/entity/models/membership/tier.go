package membership

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// MembershipTier defines tier configuration.
type MembershipTier struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	Name       string         `gorm:"type:varchar(120);not null" json:"name"`
	Code       string         `gorm:"type:varchar(64);not null;index" json:"code"`
	Status     string         `gorm:"type:varchar(32);not null;default:'draft'" json:"status"`
	Rules      datatypes.JSON `gorm:"type:jsonb" json:"rules,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MembershipTier) TableName() string { return basemodels.S(basemodels.TableMembershipTiers) }
