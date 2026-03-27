package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RiskRule defines risk matching strategies for logistics waybills.
type RiskRule struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_risk_rule_name,priority:1" json:"tenant_uuid"`
	Name        string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_risk_rule_name,priority:2" json:"name"`
	MatchField  string         `gorm:"column:match_field;type:varchar(32);not null;default:'address';index" json:"match_field"`
	MatchMode   string         `gorm:"column:match_mode;type:varchar(32);not null;default:'contains'" json:"match_mode"`
	Pattern     string         `gorm:"column:pattern;type:varchar(255);not null" json:"pattern"`
	Decision    string         `gorm:"column:decision;type:varchar(32);not null;default:'review';index" json:"decision"`
	RiskLevel   string         `gorm:"column:risk_level;type:varchar(16);not null;default:'medium'" json:"risk_level"`
	Priority    int            `gorm:"column:priority;not null;default:100;index" json:"priority"`
	Enabled     bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Description string         `gorm:"column:description;type:text" json:"description,omitempty"`
	RuleConfig  datatypes.JSON `gorm:"column:rule_config;type:jsonb" json:"rule_config,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskRule) TableName() string { return BaseModel.S(BaseModel.TableLogisticsRiskRules) }
