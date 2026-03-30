package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FulfillmentSandboxScenario stores what-if scenario parameters and strategy packs.
type FulfillmentSandboxScenario struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_sandbox_scenario,priority:1" json:"tenant_uuid"`
	Name            string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_sandbox_scenario,priority:2" json:"name"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	WarehouseID     string         `gorm:"column:warehouse_id;type:varchar(64);index" json:"warehouse_id,omitempty"`
	DestinationZone string         `gorm:"column:destination_zone;type:varchar(64);index" json:"destination_zone,omitempty"`
	BaselineConfig  datatypes.JSON `gorm:"column:baseline_config;type:jsonb" json:"baseline_config,omitempty"`
	StrategyConfig  datatypes.JSON `gorm:"column:strategy_config;type:jsonb" json:"strategy_config,omitempty"`
	Status          string         `gorm:"column:status;type:varchar(32);not null;default:'active';index" json:"status"`
	Description     string         `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedBy       string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	UpdatedBy       string         `gorm:"column:updated_by;type:varchar(64)" json:"updated_by,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FulfillmentSandboxScenario) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsFulfillmentSandboxScenarios)
}

// FulfillmentSandboxRun stores each simulation run and its metrics snapshot.
type FulfillmentSandboxRun struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_sandbox_run_scope,priority:1" json:"tenant_uuid"`
	ScenarioID     string         `gorm:"column:scenario_id;type:uuid;not null;index;index:idx_logistics_sandbox_run_scope,priority:2" json:"scenario_id"`
	RequestKey     string         `gorm:"column:request_key;type:varchar(128);not null;index;uniqueIndex:uk_logistics_sandbox_run_request,priority:2" json:"request_key"`
	WindowDays     int            `gorm:"column:window_days;not null;default:7" json:"window_days"`
	Strategy       string         `gorm:"column:strategy;type:varchar(32);not null;default:'balanced';index" json:"strategy"`
	TimelinessRate float64        `gorm:"column:timeliness_rate;type:numeric(7,2);not null;default:0" json:"timeliness_rate"`
	CostIndex      float64        `gorm:"column:cost_index;type:numeric(12,4);not null;default:0" json:"cost_index"`
	ExceptionRate  float64        `gorm:"column:exception_rate;type:numeric(7,2);not null;default:0" json:"exception_rate"`
	RecoveryHours  float64        `gorm:"column:recovery_hours;type:numeric(9,2);not null;default:0" json:"recovery_hours"`
	Score          float64        `gorm:"column:score;type:numeric(9,2);not null;default:0" json:"score"`
	Snapshot       datatypes.JSON `gorm:"column:snapshot;type:jsonb" json:"snapshot,omitempty"`
	Recommendation string         `gorm:"column:recommendation;type:text" json:"recommendation,omitempty"`
	Status         string         `gorm:"column:status;type:varchar(32);not null;default:'completed';index" json:"status"`
	CreatedBy      string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FulfillmentSandboxRun) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsFulfillmentSandboxRuns)
}
