package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SLOGuardPolicy stores tenant-level SLO guard strategies and throttle actions.
type SLOGuardPolicy struct {
	ID                string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID        string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_slo_guard_policy,priority:1" json:"tenant_uuid"`
	Name              string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_slo_guard_policy,priority:2" json:"name"`
	CarrierID         string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	WindowHours       int            `gorm:"column:window_hours;not null;default:24" json:"window_hours"`
	MinSuccessRate    float64        `gorm:"column:min_success_rate;type:numeric(7,2);not null;default:95" json:"min_success_rate"`
	MaxP95LatencyMS   int            `gorm:"column:max_p95_latency_ms;not null;default:2000" json:"max_p95_latency_ms"`
	MaxFailedRequests int            `gorm:"column:max_failed_requests;not null;default:10" json:"max_failed_requests"`
	Action            string         `gorm:"column:action;type:varchar(32);not null;default:'throttle'" json:"action"`
	ThrottleRatio     int            `gorm:"column:throttle_ratio;not null;default:50" json:"throttle_ratio"`
	Enabled           bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Metadata          datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SLOGuardPolicy) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsSLOGuardPolicies)
}

// SLOGuardState records throttle lifecycle and manual release operations.
type SLOGuardState struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_slo_guard_state_scope,priority:1" json:"tenant_uuid"`
	PolicyID         string         `gorm:"column:policy_id;type:uuid;not null;index;index:idx_logistics_slo_guard_state_scope,priority:2" json:"policy_id"`
	CarrierID        string         `gorm:"column:carrier_id;type:uuid;index;index:idx_logistics_slo_guard_state_scope,priority:3" json:"carrier_id,omitempty"`
	Status           string         `gorm:"column:status;type:varchar(32);not null;default:'normal';index" json:"status"`
	Action           string         `gorm:"column:action;type:varchar(32);not null;default:'throttle'" json:"action"`
	ThrottleRatio    int            `gorm:"column:throttle_ratio;not null;default:0" json:"throttle_ratio"`
	ReasonCode       string         `gorm:"column:reason_code;type:varchar(64);index" json:"reason_code,omitempty"`
	ReasonMessage    string         `gorm:"column:reason_message;type:text" json:"reason_message,omitempty"`
	TriggerMetrics   datatypes.JSON `gorm:"column:trigger_metrics;type:jsonb" json:"trigger_metrics,omitempty"`
	ActivatedAt      *time.Time     `gorm:"column:activated_at" json:"activated_at,omitempty"`
	RecoveredAt      *time.Time     `gorm:"column:recovered_at" json:"recovered_at,omitempty"`
	ManualReleasedBy string         `gorm:"column:manual_released_by;type:varchar(64)" json:"manual_released_by,omitempty"`
	ManualReason     string         `gorm:"column:manual_reason;type:text" json:"manual_reason,omitempty"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SLOGuardState) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsSLOGuardStates)
}
