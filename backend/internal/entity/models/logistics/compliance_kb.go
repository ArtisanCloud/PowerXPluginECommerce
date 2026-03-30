package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ComplianceKBVersion stores crossborder compliance policy snapshots.
type ComplianceKBVersion struct {
	ID                 string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID         string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_compliance_kb_version,priority:1" json:"tenant_uuid"`
	CountryCode        string         `gorm:"column:country_code;type:varchar(16);not null;index;uniqueIndex:uk_logistics_compliance_kb_version,priority:2" json:"country_code"`
	PolicyVersion      string         `gorm:"column:policy_version;type:varchar(64);not null;uniqueIndex:uk_logistics_compliance_kb_version,priority:3" json:"policy_version"`
	SourcePackID       string         `gorm:"column:source_pack_id;type:uuid;index" json:"source_pack_id,omitempty"`
	SourceVersionID    string         `gorm:"column:source_version_id;type:uuid;index" json:"source_version_id,omitempty"`
	SourceVersionNo    int            `gorm:"column:source_version_no;not null;default:0" json:"source_version_no"`
	CountryRuleMapping datatypes.JSON `gorm:"column:country_rule_mapping;type:jsonb;not null" json:"country_rule_mapping"`
	EffectiveFrom      *time.Time     `gorm:"column:effective_from;index" json:"effective_from,omitempty"`
	EffectiveTo        *time.Time     `gorm:"column:effective_to;index" json:"effective_to,omitempty"`
	RolloutScope       datatypes.JSON `gorm:"column:rollout_scope;type:jsonb" json:"rollout_scope,omitempty"`
	Status             string         `gorm:"column:status;type:varchar(32);not null;default:'draft';index" json:"status"`
	Notes              string         `gorm:"column:notes;type:text" json:"notes,omitempty"`
	PublishedBy        string         `gorm:"column:published_by;type:varchar(64)" json:"published_by,omitempty"`
	PublishedAt        *time.Time     `gorm:"column:published_at;index" json:"published_at,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ComplianceKBVersion) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsComplianceKBVersions)
}
