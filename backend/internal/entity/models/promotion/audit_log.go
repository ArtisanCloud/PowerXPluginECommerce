package promotion

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
)

const (
	AuditActionCreate        = "create"
	AuditActionUpdate        = "update"
	AuditActionActivate      = "activate"
	AuditActionPause         = "pause"
	AuditActionClone         = "clone"
	AuditActionQuoteApplied  = "quote_applied"
	AuditActionQuoteRejected = "quote_rejected"
)

// AuditLog records promotion configuration and quote events.
type AuditLog struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	PromotionID  string         `gorm:"column:promotion_id;type:uuid;not null;index" json:"promotion_id"`
	OrderID      *string        `gorm:"column:order_id;type:uuid;index" json:"order_id,omitempty"`
	Action       string         `gorm:"type:varchar(32);not null;index" json:"action"`
	ActionReason string         `gorm:"column:action_reason;type:varchar(128)" json:"action_reason,omitempty"`
	RequestID    string         `gorm:"column:request_id;type:varchar(128)" json:"request_id,omitempty"`
	CreatedBy    string         `gorm:"column:created_by;type:varchar(128)" json:"created_by,omitempty"`
	Payload      datatypes.JSON `gorm:"column:payload;type:jsonb;not null;default:'{}'" json:"payload"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (AuditLog) TableName() string { return models.S(models.TablePromotionAuditLogs) }
