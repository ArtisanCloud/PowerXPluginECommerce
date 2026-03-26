package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// LabelPrintTask records batch label print/reprint execution for a waybill.
type LabelPrintTask struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_label_print_request,priority:1" json:"tenant_uuid"`
	RequestKey    string         `gorm:"column:request_key;type:varchar(160);not null;uniqueIndex:uk_logistics_label_print_request,priority:2" json:"request_key"`
	WaybillID     string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	WaybillNo     string         `gorm:"column:waybill_no;type:varchar(128);not null;index" json:"waybill_no"`
	Status        string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	AttemptCount  int            `gorm:"column:attempt_count;not null;default:0" json:"attempt_count"`
	MaxAttempts   int            `gorm:"column:max_attempts;not null;default:3" json:"max_attempts"`
	LastError     string         `gorm:"column:last_error;type:text" json:"last_error,omitempty"`
	RetryQueuedAt *time.Time     `gorm:"column:retry_queued_at" json:"retry_queued_at,omitempty"`
	PrintedAt     *time.Time     `gorm:"column:printed_at" json:"printed_at,omitempty"`
	Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LabelPrintTask) TableName() string { return BaseModel.S(BaseModel.TableLogisticsLabelPrintTasks) }
