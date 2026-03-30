package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OpsAutomationPolicy defines retry/circuit-breaker/suppression/escalation strategies.
type OpsAutomationPolicy struct {
	ID                     string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID             string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_ops_automation_policy,priority:1" json:"tenant_uuid"`
	Name                   string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_ops_automation_policy,priority:2" json:"name"`
	CarrierID              string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	RetryStrategy          datatypes.JSON `gorm:"column:retry_strategy;type:jsonb" json:"retry_strategy,omitempty"`
	CircuitBreakerStrategy datatypes.JSON `gorm:"column:circuit_breaker_strategy;type:jsonb" json:"circuit_breaker_strategy,omitempty"`
	SuppressionRule        datatypes.JSON `gorm:"column:suppression_rule;type:jsonb" json:"suppression_rule,omitempty"`
	EscalationChain        datatypes.JSON `gorm:"column:escalation_chain;type:jsonb" json:"escalation_chain,omitempty"`
	Enabled                bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	LastEvaluatedAt        *time.Time     `gorm:"column:last_evaluated_at;index" json:"last_evaluated_at,omitempty"`
	CreatedAt              time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OpsAutomationPolicy) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsOpsAutomationPolicies)
}

// OpsAutomationRun records suppression hit, auto recovery result and escalation trace.
type OpsAutomationRun struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_ops_automation_run_scope,priority:1" json:"tenant_uuid"`
	PolicyID       string         `gorm:"column:policy_id;type:uuid;not null;index;index:idx_logistics_ops_automation_run_scope,priority:2" json:"policy_id"`
	CarrierID      string         `gorm:"column:carrier_id;type:uuid;index;index:idx_logistics_ops_automation_run_scope,priority:3" json:"carrier_id,omitempty"`
	TriggerSource  string         `gorm:"column:trigger_source;type:varchar(32);not null;default:'manual_evaluate'" json:"trigger_source"`
	TriggerKey     string         `gorm:"column:trigger_key;type:varchar(128);index" json:"trigger_key,omitempty"`
	Status         string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	Suppressed     bool           `gorm:"column:suppressed;not null;default:false" json:"suppressed"`
	AutoRecovered  bool           `gorm:"column:auto_recovered;not null;default:false" json:"auto_recovered"`
	Escalated      bool           `gorm:"column:escalated;not null;default:false" json:"escalated"`
	TakeoverBy     string         `gorm:"column:takeover_by;type:varchar(64)" json:"takeover_by,omitempty"`
	TakeoverReason string         `gorm:"column:takeover_reason;type:text" json:"takeover_reason,omitempty"`
	Payload        datatypes.JSON `gorm:"column:payload;type:jsonb" json:"payload,omitempty"`
	StartedAt      *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	FinishedAt     *time.Time     `gorm:"column:finished_at" json:"finished_at,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OpsAutomationRun) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsOpsAutomationRuns)
}
