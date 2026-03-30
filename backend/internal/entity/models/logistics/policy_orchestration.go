package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PolicyOrchestrationFlow defines tenant scoped orchestration workflow metadata.
type PolicyOrchestrationFlow struct {
	ID                 string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID         string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_policy_orchestration_flow_name,priority:1" json:"tenant_uuid"`
	Name               string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_logistics_policy_orchestration_flow_name,priority:2" json:"name"`
	Priority           int            `gorm:"column:priority;not null;default:100;index" json:"priority"`
	Status             string         `gorm:"column:status;type:varchar(32);not null;default:'draft';index" json:"status"`
	FlowDefinition     datatypes.JSON `gorm:"column:flow_definition;type:jsonb" json:"flow_definition,omitempty"`
	ConflictRelations  datatypes.JSON `gorm:"column:conflict_relations;type:jsonb" json:"conflict_relations,omitempty"`
	GrayReleaseConfig  datatypes.JSON `gorm:"column:gray_release_config;type:jsonb" json:"gray_release_config,omitempty"`
	PublishedVersionID string         `gorm:"column:published_version_id;type:varchar(64);index" json:"published_version_id,omitempty"`
	Description        string         `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedBy          string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	UpdatedBy          string         `gorm:"column:updated_by;type:varchar(64)" json:"updated_by,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PolicyOrchestrationFlow) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsPolicyOrchestrationFlows)
}

// PolicyOrchestrationVersion stores immutable snapshots for publish/rollback.
type PolicyOrchestrationVersion struct {
	ID               string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_policy_orchestration_version_no,priority:1;uniqueIndex:uk_logistics_policy_orchestration_request_key,priority:1" json:"tenant_uuid"`
	FlowID           string         `gorm:"column:flow_id;type:uuid;not null;index;uniqueIndex:uk_logistics_policy_orchestration_version_no,priority:2;uniqueIndex:uk_logistics_policy_orchestration_request_key,priority:2" json:"flow_id"`
	VersionNo        int            `gorm:"column:version_no;not null;default:1;uniqueIndex:uk_logistics_policy_orchestration_version_no,priority:3" json:"version_no"`
	RequestKey       string         `gorm:"column:request_key;type:varchar(191);not null;default:'';uniqueIndex:uk_logistics_policy_orchestration_request_key,priority:3" json:"request_key,omitempty"`
	Status           string         `gorm:"column:status;type:varchar(32);not null;default:'published';index" json:"status"`
	ChangeSummary    string         `gorm:"column:change_summary;type:text" json:"change_summary,omitempty"`
	Snapshot         datatypes.JSON `gorm:"column:snapshot;type:jsonb" json:"snapshot,omitempty"`
	ConflictReport   datatypes.JSON `gorm:"column:conflict_report;type:jsonb" json:"conflict_report,omitempty"`
	GrayReleasePlan  datatypes.JSON `gorm:"column:gray_release_plan;type:jsonb" json:"gray_release_plan,omitempty"`
	RolledBackFromID string         `gorm:"column:rolled_back_from_id;type:varchar(64);index" json:"rolled_back_from_id,omitempty"`
	PublishedAt      *time.Time     `gorm:"column:published_at;index" json:"published_at,omitempty"`
	CreatedBy        string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	UpdatedBy        string         `gorm:"column:updated_by;type:varchar(64)" json:"updated_by,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PolicyOrchestrationVersion) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsPolicyOrchestrationVersions)
}
