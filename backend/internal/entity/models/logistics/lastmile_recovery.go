package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// LastmileRecoveryRule defines terminal exception self-healing automation rule.
type LastmileRecoveryRule struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_lastmile_recovery_rule,priority:1" json:"tenant_uuid"`
	Name         string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_lastmile_recovery_rule,priority:2" json:"name"`
	TriggerEvent string         `gorm:"column:trigger_event;type:varchar(64);not null;index" json:"trigger_event"`
	Action       string         `gorm:"column:action;type:varchar(64);not null" json:"action"`
	Priority     int            `gorm:"column:priority;not null;default:100;index" json:"priority"`
	MaxRetries   int            `gorm:"column:max_retries;not null;default:3" json:"max_retries"`
	Enabled      bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Config       datatypes.JSON `gorm:"column:config;type:jsonb" json:"config,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LastmileRecoveryRule) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsLastmileRecoveryRules)
}

// LastmileRecoveryRun records each self-healing execution.
type LastmileRecoveryRun struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_lastmile_recovery_run,priority:1" json:"tenant_uuid"`
	RequestKey   string         `gorm:"column:request_key;type:varchar(128);not null;uniqueIndex:uk_logistics_lastmile_recovery_run,priority:2;index" json:"request_key"`
	RuleID       string         `gorm:"column:rule_id;type:uuid;not null;index" json:"rule_id"`
	WaybillID    string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	WaybillNo    string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no,omitempty"`
	TriggerEvent string         `gorm:"column:trigger_event;type:varchar(64);not null;index" json:"trigger_event"`
	Action       string         `gorm:"column:action;type:varchar(64);not null" json:"action"`
	Status       string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	RetryCount   int            `gorm:"column:retry_count;not null;default:0" json:"retry_count"`
	MaxRetries   int            `gorm:"column:max_retries;not null;default:3" json:"max_retries"`
	Message      string         `gorm:"column:message;type:text" json:"message,omitempty"`
	ManualTaken  bool           `gorm:"column:manual_taken;not null;default:false;index" json:"manual_taken"`
	TakenBy      string         `gorm:"column:taken_by;type:varchar(64)" json:"taken_by,omitempty"`
	TakenReason  string         `gorm:"column:taken_reason;type:text" json:"taken_reason,omitempty"`
	TakenAt      *time.Time     `gorm:"column:taken_at" json:"taken_at,omitempty"`
	Metadata     datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LastmileRecoveryRun) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsLastmileRecoveryRuns)
}
