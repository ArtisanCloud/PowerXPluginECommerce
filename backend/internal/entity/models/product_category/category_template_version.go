package product_category

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CategoryTemplateVersion stores immutable published snapshots.
type CategoryTemplateVersion struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_category_template_versions,priority:1" json:"tenant_uuid"`
	TemplateID    string         `gorm:"column:template_id;type:uuid;not null;index;uniqueIndex:uk_category_template_versions,priority:2" json:"template_id"`
	VersionNumber int            `gorm:"not null;default:1;uniqueIndex:uk_category_template_versions,priority:3;comment:版本号" json:"version_number"`
	PublishedAt   *time.Time     `gorm:"comment:发布时间" json:"published_at,omitempty"`
	PublishedBy   string         `gorm:"type:varchar(64);comment:发布人" json:"published_by,omitempty"`
	RollbackFrom  *string        `gorm:"column:rollback_from;type:uuid;index;comment:回滚来源版本ID" json:"rollback_from,omitempty"`
	Snapshot      datatypes.JSON `gorm:"type:jsonb;not null;comment:版本快照" json:"snapshot"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoryTemplateVersion) TableName() string {
	return models.S(models.TableProductCategoryTemplateVersions)
}
