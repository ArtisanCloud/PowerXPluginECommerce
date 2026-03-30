package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GatewayFailureEvent records gateway failure classifications and compensation lifecycle.
type GatewayFailureEvent struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	WaybillID       string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	WaybillNo       string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no,omitempty"`
	Provider        string         `gorm:"column:provider;type:varchar(64);index" json:"provider,omitempty"`
	SourceJobID     string         `gorm:"column:source_job_id;type:uuid;index" json:"source_job_id,omitempty"`
	ErrorClass      string         `gorm:"column:error_class;type:varchar(32);not null;index" json:"error_class"`
	ErrorCode       string         `gorm:"column:error_code;type:varchar(64);index" json:"error_code,omitempty"`
	ErrorMessage    string         `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	Status          string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	RetryCount      int            `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	NextRetryAt     *time.Time     `gorm:"column:next_retry_at;index" json:"next_retry_at,omitempty"`
	CircuitOpenTill *time.Time     `gorm:"column:circuit_open_till" json:"circuit_open_till,omitempty"`
	RecoveredAt     *time.Time     `gorm:"column:recovered_at" json:"recovered_at,omitempty"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GatewayFailureEvent) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsGatewayFailureEvents)
}
