package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RedeliveryTask tracks failed-delivery follow-up and redispatch lifecycle.
type RedeliveryTask struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_redelivery_request,priority:1" json:"tenant_uuid"`
	WaybillID       string         `gorm:"column:waybill_id;type:uuid;not null;index" json:"waybill_id"`
	WaybillNo       string         `gorm:"column:waybill_no;type:varchar(128);not null;index" json:"waybill_no"`
	RequestKey      string         `gorm:"column:request_key;type:varchar(191);not null;default:'';uniqueIndex:uk_logistics_redelivery_request,priority:2" json:"request_key,omitempty"`
	Status          string         `gorm:"column:status;type:varchar(32);not null;default:'initiated';index" json:"status"`
	AttemptNo       int            `gorm:"column:attempt_no;not null;default:1" json:"attempt_no"`
	AddressSnapshot datatypes.JSON `gorm:"column:address_snapshot;type:jsonb" json:"address_snapshot,omitempty"`
	LastReason      string         `gorm:"column:last_reason;type:text" json:"last_reason,omitempty"`
	OperatorID      string         `gorm:"column:operator_id;type:varchar(128)" json:"operator_id,omitempty"`
	ClosedAt        *time.Time     `gorm:"column:closed_at" json:"closed_at,omitempty"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RedeliveryTask) TableName() string { return BaseModel.S(BaseModel.TableLogisticsRedeliveryTasks) }
