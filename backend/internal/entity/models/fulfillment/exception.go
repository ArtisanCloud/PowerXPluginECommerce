package fulfillment

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Exception records fulfillment failures and escalation lifecycle.
type Exception struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	TaskID        string         `gorm:"column:task_id;type:uuid;not null;index" json:"task_id"`
	WaybillID     string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	Type          string         `gorm:"type:varchar(64);not null;index" json:"type"`
	Status        string         `gorm:"type:varchar(32);not null;default:'open';index" json:"status"`
	Reason        string         `gorm:"type:text" json:"reason,omitempty"`
	FirstActionAt *time.Time     `gorm:"column:first_action_at" json:"first_action_at,omitempty"`
	EscalatedAt   *time.Time     `gorm:"column:escalated_at" json:"escalated_at,omitempty"`
	Metadata      datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Exception) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentExceptions) }
