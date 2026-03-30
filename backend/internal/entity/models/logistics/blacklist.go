package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BlacklistEntry stores explicit denylist constraints for risky recipients/addresses.
type BlacklistEntry struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	EntryType      string         `gorm:"column:entry_type;type:varchar(32);not null;default:'recipient';index" json:"entry_type"`
	RecipientName  string         `gorm:"column:recipient_name;type:varchar(128);index" json:"recipient_name,omitempty"`
	RecipientPhone string         `gorm:"column:recipient_phone;type:varchar(64);index" json:"recipient_phone,omitempty"`
	AddressLine    string         `gorm:"column:address_line;type:varchar(255);index" json:"address_line,omitempty"`
	Reason         string         `gorm:"column:reason;type:text" json:"reason,omitempty"`
	Status         string         `gorm:"column:status;type:varchar(32);not null;default:'active';index" json:"status"`
	ExpiresAt      *time.Time     `gorm:"column:expires_at;index" json:"expires_at,omitempty"`
	Metadata       datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BlacklistEntry) TableName() string { return BaseModel.S(BaseModel.TableLogisticsBlacklists) }
