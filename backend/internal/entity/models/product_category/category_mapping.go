package product_category

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	MappingStrategyManual = "manual"
	MappingStrategyAuto   = "auto"

	MappingSyncPending  = "pending"
	MappingSyncSynced   = "synced"
	MappingSyncFailed   = "failed"
	MappingSyncDisabled = "disabled"
)

// CategoryMapping represents external channel category mapping.
type CategoryMapping struct {
	ID                 string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID         string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_category_mapping,priority:1" json:"tenant_uuid"`
	CategoryID         string         `gorm:"column:category_id;type:uuid;not null;index;uniqueIndex:uk_category_mapping,priority:2" json:"category_id"`
	Channel            string         `gorm:"type:varchar(64);not null;index;uniqueIndex:uk_category_mapping,priority:3" json:"channel"`
	PlatformCategoryID string         `gorm:"column:platform_category_id;type:varchar(128);not null;index;comment:平台类目ID" json:"platform_category_id"`
	Strategy           string         `gorm:"type:varchar(32);not null;default:'manual'" json:"strategy"`
	SyncStatus         string         `gorm:"column:sync_status;type:varchar(32);not null;default:'pending'" json:"sync_status"`
	Metadata           datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoryMapping) TableName() string { return models.S(models.TableProductCategoryMappings) }

func (m *CategoryMapping) Normalize() {
	if m == nil {
		return
	}
	m.TenantUUID = strings.TrimSpace(m.TenantUUID)
	m.CategoryID = strings.TrimSpace(m.CategoryID)
	m.Channel = strings.TrimSpace(m.Channel)
	m.PlatformCategoryID = strings.TrimSpace(m.PlatformCategoryID)
	m.Strategy = strings.TrimSpace(strings.ToLower(m.Strategy))
	m.SyncStatus = strings.TrimSpace(strings.ToLower(m.SyncStatus))
}
