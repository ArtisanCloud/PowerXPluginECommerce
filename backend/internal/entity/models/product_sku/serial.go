package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
)

// ProductSKUSerialRecord stores optional serial numbers / batch attributes for compliance.
type ProductSKUSerialRecord struct {
	ID         string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string     `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	SKUId      string     `gorm:"column:sku_id;type:uuid;not null;index" json:"sku_id"`
	SerialNo   string     `gorm:"type:varchar(128);not null;index:idx_serial_unique,priority:1" json:"serial_no"`
	BatchNo    string     `gorm:"type:varchar(64);index:idx_serial_unique,priority:2" json:"batch_no,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Status     string     `gorm:"type:varchar(32);not null;default:'available'" json:"status"`
	AuditLogID string     `gorm:"type:uuid" json:"audit_log_id,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductSKUSerialRecord) TableName() string {
	return basemodels.S(basemodels.TableProductSkuSerials)
}
