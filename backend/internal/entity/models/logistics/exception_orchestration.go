package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ExceptionOrchestrationRule struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_orchestration_rule,priority:1" json:"tenant_uuid"`
	Name         string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_orchestration_rule,priority:2" json:"name"`
	TriggerEvent string         `gorm:"column:trigger_event;type:varchar(64);not null;index" json:"trigger_event"`
	Action       string         `gorm:"column:action;type:varchar(64);not null" json:"action"`
	Priority     int            `gorm:"column:priority;not null;default:100;index" json:"priority"`
	Enabled      bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Config       datatypes.JSON `gorm:"column:config;type:jsonb" json:"config,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ExceptionOrchestrationRule) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsExceptionRules)
}

type ExceptionOrchestrationRun struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	RuleID     string         `gorm:"column:rule_id;type:uuid;not null;index" json:"rule_id"`
	WaybillID  string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	WaybillNo  string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no,omitempty"`
	Trigger    string         `gorm:"column:trigger;type:varchar(64);not null;index" json:"trigger"`
	Result     string         `gorm:"column:result;type:varchar(32);not null;default:'queued';index" json:"result"`
	Message    string         `gorm:"column:message;type:text" json:"message,omitempty"`
	Metadata   datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ExceptionOrchestrationRun) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsExceptionRuns)
}
