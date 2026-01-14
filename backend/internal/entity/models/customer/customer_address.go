package customer

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CustomerAddress represents a shipping address record for a customer (address book).
// It is tenant-scoped and guarded by RLS.
type CustomerAddress struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	CustomerID     string         `gorm:"column:customer_id;type:uuid;not null;index" json:"customer_id"`
	IsDefault      bool           `gorm:"column:is_default;not null;default:false;index" json:"is_default"`
	Label          string         `gorm:"type:varchar(64)" json:"label,omitempty"`
	RecipientName  string         `gorm:"column:recipient_name;type:varchar(64);not null" json:"recipient_name"`
	RecipientPhone string         `gorm:"column:recipient_phone;type:varchar(32);not null" json:"recipient_phone"`
	CountryCode    string         `gorm:"column:country_code;type:varchar(8)" json:"country_code,omitempty"`
	Province       string         `gorm:"type:varchar(64)" json:"province,omitempty"`
	City           string         `gorm:"type:varchar(64)" json:"city,omitempty"`
	District       string         `gorm:"type:varchar(64)" json:"district,omitempty"`
	Address1       string         `gorm:"column:address1;type:varchar(256);not null" json:"address1"`
	Address2       string         `gorm:"column:address2;type:varchar(256)" json:"address2,omitempty"`
	PostalCode     string         `gorm:"column:postal_code;type:varchar(16)" json:"postal_code,omitempty"`
	Metadata       datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomerAddress) TableName() string {
	return basemodels.S(basemodels.TableCustomerAddresses)
}
