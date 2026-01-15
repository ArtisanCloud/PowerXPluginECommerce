package order

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
)

// OrderEvent records important order lifecycle events for audit and tracing.
type OrderEvent struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	OrderID      string         `gorm:"column:order_id;type:uuid;not null;index" json:"order_id"`
	EventType    string         `gorm:"type:varchar(64);not null;index" json:"event_type"`
	OperatorType string         `gorm:"type:varchar(16);not null" json:"operator_type"`
	Operator     string         `gorm:"type:varchar(128)" json:"operator,omitempty"`
	Payload      datatypes.JSON `gorm:"type:jsonb" json:"payload,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderEvent) TableName() string {
	return basemodels.S(basemodels.TableOrderEvents)
}
