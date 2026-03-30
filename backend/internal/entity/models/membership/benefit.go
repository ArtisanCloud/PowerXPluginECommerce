package membership

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// MembershipBenefit defines a benefit or a bundle of benefit items.
type MembershipBenefit struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	Name       string         `gorm:"type:varchar(120);not null" json:"name"`
	Type       string         `gorm:"type:varchar(16);not null;default:'single'" json:"type"`
	Items      datatypes.JSON `gorm:"type:jsonb" json:"items,omitempty"`
	Status     string         `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MembershipBenefit) TableName() string { return basemodels.S(basemodels.TableMembershipBenefits) }
