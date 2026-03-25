package fulfillment

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Task represents warehouse execution progress for an order.
type Task struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	OrderID     string         `gorm:"column:order_id;type:uuid;not null;index" json:"order_id"`
	WarehouseID string         `gorm:"column:warehouse_id;type:uuid;not null;index" json:"warehouse_id"`
	Status      string         `gorm:"type:varchar(32);not null;default:'pending';index" json:"status"`
	AssignedTo  string         `gorm:"column:assigned_to;type:varchar(128)" json:"assigned_to,omitempty"`
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Task) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentTasks) }

// TaskLog stores operator-level task status changes.
type TaskLog struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	TaskID     string         `gorm:"column:task_id;type:uuid;not null;index" json:"task_id"`
	Action     string         `gorm:"type:varchar(64);not null" json:"action"`
	OperatorID string         `gorm:"column:operator_id;type:varchar(128)" json:"operator_id,omitempty"`
	Detail     datatypes.JSON `gorm:"type:jsonb" json:"detail,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

func (TaskLog) TableName() string { return BaseModel.S(BaseModel.TableFulfillmentTaskLogs) }
