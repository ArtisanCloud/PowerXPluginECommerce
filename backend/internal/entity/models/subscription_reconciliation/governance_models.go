package subscription_reconciliation

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DeltaTask tracks remediation workflow for a delta item.
type DeltaTask struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_sr_task_fp,priority:1" json:"tenant_uuid"`
	DeltaID          string         `gorm:"column:delta_id;type:uuid;not null;index" json:"delta_id"`
	DeltaFingerprint string         `gorm:"column:delta_fingerprint;type:varchar(128);not null;uniqueIndex:uk_sr_task_fp,priority:2" json:"delta_fingerprint"`
	Assignee         string         `gorm:"column:assignee;type:varchar(64);index" json:"assignee,omitempty"`
	Priority         string         `gorm:"column:priority;type:varchar(24)" json:"priority,omitempty"`
	SLALevel         string         `gorm:"column:sla_level;type:varchar(16);not null" json:"sla_level"`
	SLADeadline      *time.Time     `gorm:"column:sla_deadline" json:"sla_deadline,omitempty"`
	Status           string         `gorm:"column:status;type:varchar(24);not null;default:'pending';index" json:"status"`
	Resolution       string         `gorm:"column:resolution;type:varchar(32)" json:"resolution,omitempty"`
	ResolutionNote   string         `gorm:"column:resolution_note;type:text" json:"resolution_note,omitempty"`
	ClosedAt         *time.Time     `gorm:"column:closed_at" json:"closed_at,omitempty"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeltaTask) TableName() string {
	return BaseModel.S(BaseModel.TableSubscriptionReconciliationDeltaTasks)
}

// RenewalGovernancePolicy configures retry/notify/escalation policies.
type RenewalGovernancePolicy struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	Enabled             bool           `gorm:"column:enabled;not null;default:true" json:"enabled"`
	RetryWindows        datatypes.JSON `gorm:"column:retry_windows;type:jsonb" json:"retry_windows,omitempty"`
	EscalationThreshold int            `gorm:"column:escalation_threshold;not null;default:3" json:"escalation_threshold"`
	NotifyChannels      datatypes.JSON `gorm:"column:notify_channels;type:jsonb" json:"notify_channels,omitempty"`
	Version             int            `gorm:"column:version;not null;default:1;index" json:"version"`
	EffectiveFrom       *time.Time     `gorm:"column:effective_from;index" json:"effective_from,omitempty"`
	EffectiveTo         *time.Time     `gorm:"column:effective_to;index" json:"effective_to,omitempty"`
	Metadata            datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RenewalGovernancePolicy) TableName() string {
	return BaseModel.S(BaseModel.TableSubscriptionReconciliationGovernancePolicies)
}

// RenewalExecutionLog records retry/notify/escalate execution traces.
type RenewalExecutionLog struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	SubscriptionRef string         `gorm:"column:subscription_ref;type:varchar(64);not null;index" json:"subscription_ref"`
	ActionType      string         `gorm:"column:action_type;type:varchar(24);not null;index" json:"action_type"`
	AttemptNo       int            `gorm:"column:attempt_no;not null;default:0" json:"attempt_no"`
	ScheduledAt     *time.Time     `gorm:"column:scheduled_at" json:"scheduled_at,omitempty"`
	ExecutedAt      *time.Time     `gorm:"column:executed_at" json:"executed_at,omitempty"`
	Result          string         `gorm:"column:result;type:varchar(16);not null;index" json:"result"`
	FailureReason   string         `gorm:"column:failure_reason;type:text" json:"failure_reason,omitempty"`
	OperatorType    string         `gorm:"column:operator_type;type:varchar(16);not null;default:'system'" json:"operator_type"`
	OperatorID      string         `gorm:"column:operator_id;type:varchar(64)" json:"operator_id,omitempty"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RenewalExecutionLog) TableName() string {
	return BaseModel.S(BaseModel.TableSubscriptionReconciliationExecutionLogs)
}
