package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CustomsRulePack stores country-level customs governance rule pack.
type CustomsRulePack struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_customs_rule_pack,priority:1" json:"tenant_uuid"`
	Name             string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_customs_rule_pack,priority:2" json:"name"`
	CountryCode      string         `gorm:"column:country_code;type:varchar(8);not null;index" json:"country_code"`
	Status           string         `gorm:"column:status;type:varchar(32);not null;default:'draft';index" json:"status"`
	Strategy         string         `gorm:"column:strategy;type:varchar(32);not null;default:'first_hit'" json:"strategy"`
	DefaultRiskLevel string         `gorm:"column:default_risk_level;type:varchar(16);not null;default:'low'" json:"default_risk_level"`
	Description      string         `gorm:"column:description;type:text" json:"description,omitempty"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomsRulePack) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCustomsRulePacks)
}

// CustomsRuleVersion stores versioned rules under a customs rule pack.
type CustomsRuleVersion struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_customs_rule_version,priority:1" json:"tenant_uuid"`
	PackID       string         `gorm:"column:pack_id;type:uuid;not null;index;uniqueIndex:uk_logistics_customs_rule_version,priority:2" json:"pack_id"`
	VersionNo    int            `gorm:"column:version_no;not null;default:1;uniqueIndex:uk_logistics_customs_rule_version,priority:3" json:"version_no"`
	Status       string         `gorm:"column:status;type:varchar(32);not null;default:'draft';index" json:"status"`
	Rules        datatypes.JSON `gorm:"column:rules;type:jsonb;not null" json:"rules"`
	HitStrategy  string         `gorm:"column:hit_strategy;type:varchar(32);not null;default:'first_hit'" json:"hit_strategy"`
	RiskSnapshot datatypes.JSON `gorm:"column:risk_snapshot;type:jsonb" json:"risk_snapshot,omitempty"`
	PublishedAt  *time.Time     `gorm:"column:published_at" json:"published_at,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CustomsRuleVersion) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCustomsRuleVersions)
}
