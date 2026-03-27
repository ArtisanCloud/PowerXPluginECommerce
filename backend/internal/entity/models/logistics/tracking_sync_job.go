package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TrackingSyncJob records batch provider-pull tracking synchronization jobs.
type TrackingSyncJob struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	Provider        string         `gorm:"column:provider;type:varchar(64);index" json:"provider,omitempty"`
	WaybillStatus   string         `gorm:"column:waybill_status;type:varchar(32);index" json:"waybill_status,omitempty"`
	Status          string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	BatchLimit      int            `gorm:"column:batch_limit;not null;default:20" json:"batch_limit"`
	EventLimit      int            `gorm:"column:event_limit;not null;default:20" json:"event_limit"`
	TotalWaybills   int            `gorm:"column:total_waybills;not null;default:0" json:"total_waybills"`
	SuccessCount    int            `gorm:"column:success_count;not null;default:0" json:"success_count"`
	FailedCount     int            `gorm:"column:failed_count;not null;default:0" json:"failed_count"`
	AppendedCount   int            `gorm:"column:appended_count;not null;default:0" json:"appended_count"`
	ReplayedCount   int            `gorm:"column:replayed_count;not null;default:0" json:"replayed_count"`
	P95LatencyMS    int            `gorm:"column:p95_latency_ms;not null;default:0" json:"p95_latency_ms"`
	LastError       string         `gorm:"column:last_error;type:text" json:"last_error,omitempty"`
	CancelledReason string         `gorm:"column:cancelled_reason;type:text" json:"cancelled_reason,omitempty"`
	StartedAt       *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	FinishedAt      *time.Time     `gorm:"column:finished_at" json:"finished_at,omitempty"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TrackingSyncJob) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsTrackingSyncJobs)
}
