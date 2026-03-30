package membership

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// Entitlement represents granted benefit items.
type Entitlement struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"type:uuid;not null;index;uniqueIndex:uk_entitlement_grant,priority:1" json:"tenant_uuid"`
	CustomerID  string         `gorm:"type:uuid;not null;index;uniqueIndex:uk_entitlement_grant,priority:2" json:"customer_id"`
	ServiceCode string         `gorm:"type:varchar(120);not null;index;uniqueIndex:uk_entitlement_grant,priority:5" json:"service_code"`
	Quantity    int64          `gorm:"not null" json:"quantity"`
	ValidFrom   *time.Time     `gorm:"" json:"valid_from,omitempty"`
	ValidTo     *time.Time     `gorm:"" json:"valid_to,omitempty"`
	StackPolicy string         `gorm:"type:varchar(16);not null;default:'stack'" json:"stack_policy"`
	SourceType  string         `gorm:"type:varchar(32);not null;uniqueIndex:uk_entitlement_grant,priority:3" json:"source_type"`
	SourceID    string         `gorm:"type:varchar(128);not null;uniqueIndex:uk_entitlement_grant,priority:4" json:"source_id"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Entitlement) TableName() string { return basemodels.S(basemodels.TableEntitlements) }
