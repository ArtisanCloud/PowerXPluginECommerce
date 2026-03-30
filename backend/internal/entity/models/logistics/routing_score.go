package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RoutingScoreProfile stores multi-objective routing weights.
type RoutingScoreProfile struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_routing_score_profile,priority:1" json:"tenant_uuid"`
	Name             string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_routing_score_profile,priority:2" json:"name"`
	TimelinessWeight float64        `gorm:"column:timeliness_weight;not null;default:0.4" json:"timeliness_weight"`
	CostWeight       float64        `gorm:"column:cost_weight;not null;default:0.3" json:"cost_weight"`
	QuotaWeight      float64        `gorm:"column:quota_weight;not null;default:0.2" json:"quota_weight"`
	RiskWeight       float64        `gorm:"column:risk_weight;not null;default:0.1" json:"risk_weight"`
	FallbackStrategy string         `gorm:"column:fallback_strategy;type:varchar(32);not null;default:'highest_timeliness'" json:"fallback_strategy"`
	Enabled          bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Config           datatypes.JSON `gorm:"column:config;type:jsonb" json:"config,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RoutingScoreProfile) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsRoutingScoreProfiles)
}

// RoutingScoreSimulation stores optimizer simulation snapshots.
type RoutingScoreSimulation struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	RequestKey     string         `gorm:"column:request_key;type:varchar(128);not null;index" json:"request_key"`
	ProfileID      string         `gorm:"column:profile_id;type:uuid;index" json:"profile_id,omitempty"`
	PreferredID    string         `gorm:"column:preferred_carrier_id;type:uuid;index" json:"preferred_carrier_id,omitempty"`
	ChosenID       string         `gorm:"column:chosen_carrier_id;type:uuid;index" json:"chosen_carrier_id,omitempty"`
	ChosenName     string         `gorm:"column:chosen_carrier_name;type:varchar(128)" json:"chosen_carrier_name,omitempty"`
	Strategy       string         `gorm:"column:strategy;type:varchar(32);not null;default:'optimizer'" json:"strategy"`
	Degraded       bool           `gorm:"column:degraded;not null;default:false" json:"degraded"`
	Reason         string         `gorm:"column:reason;type:text" json:"reason,omitempty"`
	InputPayload   datatypes.JSON `gorm:"column:input_payload;type:jsonb" json:"input_payload,omitempty"`
	Candidates     datatypes.JSON `gorm:"column:candidates;type:jsonb" json:"candidates,omitempty"`
	ExplainPayload datatypes.JSON `gorm:"column:explain_payload;type:jsonb" json:"explain_payload,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RoutingScoreSimulation) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsRoutingScoreSimulations)
}
