package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RiskHit records risk-matching outcomes and manual release actions.
type RiskHit struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	WaybillID     string         `gorm:"column:waybill_id;type:uuid;not null;default:'';index" json:"waybill_id,omitempty"`
	WaybillNo     string         `gorm:"column:waybill_no;type:varchar(128);not null;default:'';index" json:"waybill_no,omitempty"`
	RuleID        string         `gorm:"column:rule_id;type:uuid;not null;default:'';index" json:"rule_id,omitempty"`
	BlacklistID   string         `gorm:"column:blacklist_id;type:uuid;not null;default:'';index" json:"blacklist_id,omitempty"`
	Source        string         `gorm:"column:source;type:varchar(32);not null;default:'rule';index" json:"source"`
	Decision      string         `gorm:"column:decision;type:varchar(32);not null;default:'review';index" json:"decision"`
	RiskLevel     string         `gorm:"column:risk_level;type:varchar(16);not null;default:'medium'" json:"risk_level"`
	Fingerprint   string         `gorm:"column:fingerprint;type:varchar(255);not null;default:'';index" json:"fingerprint,omitempty"`
	Description   string         `gorm:"column:description;type:text" json:"description,omitempty"`
	Status        string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	ReleasedBy    string         `gorm:"column:released_by;type:varchar(128)" json:"released_by,omitempty"`
	ReleaseReason string         `gorm:"column:release_reason;type:text" json:"release_reason,omitempty"`
	ReleasedAt    *time.Time     `gorm:"column:released_at;index" json:"released_at,omitempty"`
	Payload       datatypes.JSON `gorm:"column:payload;type:jsonb" json:"payload,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskHit) TableName() string { return BaseModel.S(BaseModel.TableLogisticsRiskHits) }
