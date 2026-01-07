package product_category

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

const (
	TemplateStatusEnabled  = "enabled"
	TemplateStatusDisabled = "disabled"
)

// CategoryTemplate represents a reusable template bound to categories.
type CategoryTemplate struct {
	ID                        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID                string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	Name                      string         `gorm:"type:varchar(255);not null;comment:模板名称" json:"name"`
	Status                    string         `gorm:"type:varchar(32);not null;default:'enabled';index" json:"status"`
	ApplicableLevel           *int           `gorm:"column:applicable_level;comment:适用层级" json:"applicable_level,omitempty"`
	Notes                     string         `gorm:"type:text" json:"notes,omitempty"`
	CurrentPublishedVersionID *string        `gorm:"column:current_published_version_id;type:uuid;index;comment:当前已发布版本ID" json:"current_published_version_id,omitempty"`
	CurrentPublishedVersionNo int            `gorm:"column:current_published_version_no;not null;default:0;comment:当前已发布版本号" json:"current_published_version_no"`
	LastPublishedAt           *time.Time     `gorm:"column:last_published_at;comment:最近发布时间" json:"last_published_at,omitempty"`
	CreatedAt                 time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt                 gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoryTemplate) TableName() string { return models.S(models.TableProductCategoryTemplates) }

func (t *CategoryTemplate) Normalize() {
	if t == nil {
		return
	}
	t.TenantUUID = strings.TrimSpace(t.TenantUUID)
	t.Name = strings.TrimSpace(t.Name)
	t.Status = strings.TrimSpace(strings.ToLower(t.Status))
	t.Notes = strings.TrimSpace(t.Notes)
}
