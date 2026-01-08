package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
)

// PricebookAuditLog captures key lifecycle events for pricebooks and their versions.
type PricebookAuditLog struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:审计ID" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index;comment:租户" json:"tenant_uuid"`
	ResourceType string         `gorm:"column:resource_type;type:text;not null;index;comment:资源类型" json:"resource_type"`
	ResourceID   string         `gorm:"column:resource_id;type:uuid;not null;index;comment:资源ID" json:"resource_id"`
	Action       string         `gorm:"type:text;not null;index;comment:动作" json:"action"`
	Actor        string         `gorm:"type:text;comment:操作者" json:"actor,omitempty"`
	Payload      datatypes.JSON `gorm:"type:jsonb;comment:摘要" json:"payload,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// TableName implements gorm.Tabler.
func (PricebookAuditLog) TableName() string { return models.S(models.TablePricebookAuditLogs) }
