package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// PricebookVersion represents an immutable snapshot of a pricebook at a point in time.
type PricebookVersion struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:价目表版本ID" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_pricebook_version,priority:1;comment:租户" json:"tenant_uuid"`
	PricebookID string         `gorm:"column:pricebook_id;type:uuid;not null;index;uniqueIndex:uk_pricebook_version,priority:2;comment:价目表ID" json:"pricebook_id"`
	Version     int            `gorm:"type:int;not null;uniqueIndex:uk_pricebook_version,priority:3;comment:版本号" json:"version"`
	State       string         `gorm:"type:text;not null;default:'draft';index;comment:状态(draft/active/archived/expired)" json:"state"`
	EffectiveAt time.Time      `gorm:"column:effective_at;type:timestamptz;not null;index;comment:生效时间" json:"effective_at"`
	ExpiresAt   *time.Time     `gorm:"column:expires_at;type:timestamptz;index;comment:失效时间" json:"expires_at,omitempty"`
	PublishedAt *time.Time     `gorm:"column:published_at;type:timestamptz;index;comment:发布时间" json:"published_at,omitempty"`
	PublishedBy string         `gorm:"column:published_by;type:text;comment:发布人" json:"published_by,omitempty"`
	Note        string         `gorm:"type:text;comment:发布说明" json:"note,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName implements gorm.Tabler.
func (PricebookVersion) TableName() string { return models.S(models.TablePricebookVersions) }
