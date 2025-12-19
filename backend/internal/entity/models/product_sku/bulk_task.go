package product_sku

import (
	"time"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
)

// ProductSKUBulkTask records asynchronous bulk adjustments (price/inventory/export).
type ProductSKUBulkTask struct {
	TaskID           string         `gorm:"column:task_id;type:varchar(64);primaryKey" json:"task_id"`
	TenantUUID       string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	TaskType         string         `gorm:"type:varchar(32);not null" json:"task_type"`
	Scope            datatypes.JSON `gorm:"type:jsonb;comment:SKU 范围" json:"scope,omitempty"`
	Operation        datatypes.JSON `gorm:"type:jsonb;comment:操作内容" json:"operation,omitempty"`
	ApprovalRequired bool           `gorm:"not null;default:false" json:"approval_required"`
	ApprovalState    string         `gorm:"type:varchar(32);default:'pending'" json:"approval_state"`
	Status           string         `gorm:"type:varchar(32);not null;default:'pending';index" json:"status"`
	SubmittedBy      string         `gorm:"type:varchar(64)" json:"submitted_by,omitempty"`
	ApprovedBy       string         `gorm:"type:varchar(64)" json:"approved_by,omitempty"`
	ApprovedAt       *time.Time     `json:"approved_at,omitempty"`
	ErrorReportURL   string         `gorm:"type:text" json:"error_report_url,omitempty"`
	Result           datatypes.JSON `gorm:"type:jsonb" json:"result,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProductSKUBulkTask) TableName() string {
	return basemodels.S(basemodels.TableProductSkuBulkTasks)
}

// ProductSKUBulkTaskItem stores per SKU execution status.
type ProductSKUBulkTaskItem struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index" json:"tenant_uuid"`
	TaskID     string         `gorm:"type:varchar(64);not null;index" json:"task_id"`
	SKUId      string         `gorm:"column:sku_id;type:uuid;not null;index" json:"sku_id"`
	Status     string         `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`
	Message    string         `gorm:"type:text" json:"message,omitempty"`
	Diff       datatypes.JSON `gorm:"type:jsonb" json:"diff,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProductSKUBulkTaskItem) TableName() string {
	return basemodels.S(basemodels.TableProductSkuBulkTaskItems)
}
