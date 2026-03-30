package reverse

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// InspectionRule defines decision rules for reverse-waybill inspection.
type InspectionRule struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	Name           string         `gorm:"type:varchar(128);not null" json:"name"`
	Priority       int            `gorm:"type:int;not null;default:100;index" json:"priority"`
	Enabled        bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	ConditionJSON  datatypes.JSON `gorm:"column:condition_json;type:jsonb;not null" json:"condition_json"`
	Decision       string         `gorm:"type:varchar(32);not null" json:"decision"`
	Recommendation string         `gorm:"type:varchar(64);not null" json:"recommendation"`
	Notes          string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InspectionRule) TableName() string {
	return BaseModel.S(BaseModel.TableReverseInspectionRules)
}

// WaybillInspection stores idempotent inspection outcomes for a reverse waybill.
type WaybillInspection struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_reverse_waybill_inspection,priority:1" json:"tenant_uuid"`
	WaybillID      string         `gorm:"column:waybill_id;type:uuid;not null;index;uniqueIndex:uk_reverse_waybill_inspection,priority:2" json:"waybill_id"`
	RuleID         string         `gorm:"column:rule_id;type:uuid;index" json:"rule_id,omitempty"`
	Decision       string         `gorm:"type:varchar(32);not null" json:"decision"`
	Recommendation string         `gorm:"type:varchar(64);not null" json:"recommendation"`
	Reason         string         `gorm:"type:text" json:"reason,omitempty"`
	Attributes     datatypes.JSON `gorm:"type:jsonb;not null" json:"attributes"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WaybillInspection) TableName() string {
	return BaseModel.S(BaseModel.TableReverseWaybillInspections)
}
