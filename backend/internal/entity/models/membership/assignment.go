package membership

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// MembershipAssignment links customer to tier.
type MembershipAssignment struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	CustomerID string         `gorm:"type:uuid;not null;index" json:"customer_id"`
	TierID     string         `gorm:"type:uuid;not null;index" json:"tier_id"`
	Status     string         `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
	ValidFrom  *time.Time     `gorm:"" json:"valid_from,omitempty"`
	ValidTo    *time.Time     `gorm:"" json:"valid_to,omitempty"`
	SourceType string         `gorm:"type:varchar(32);not null" json:"source_type"`
	SourceID   string         `gorm:"type:varchar(128);not null;index" json:"source_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MembershipAssignment) TableName() string {
	return basemodels.S(basemodels.TableMembershipAssignments)
}
