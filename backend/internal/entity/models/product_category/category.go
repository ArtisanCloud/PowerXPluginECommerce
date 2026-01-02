package product_category

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

const (
	CategoryStatusEnabled  = "enabled"
	CategoryStatusDisabled = "disabled"
)

// ProductCategory represents a tenant-scoped product category node.
type ProductCategory struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:类目主键" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_product_categories_tenant_code,priority:1;uniqueIndex:uk_product_categories_tenant_alias_slug,priority:1;uniqueIndex:uk_product_categories_tenant_parent_display_name,priority:1;index:idx_product_categories_tenant_path,priority:1;comment:租户" json:"tenant_uuid"`
	ParentID       *string        `gorm:"column:parent_id;type:uuid;index;uniqueIndex:uk_product_categories_tenant_parent_display_name,priority:2;comment:父类目ID" json:"parent_id,omitempty"`
	Code           string         `gorm:"type:varchar(120);not null;uniqueIndex:uk_product_categories_tenant_code,priority:2;comment:类目编码" json:"code"`
	DisplayName    string         `gorm:"column:display_name;type:varchar(255);not null;uniqueIndex:uk_product_categories_tenant_parent_display_name,priority:3;comment:显示名称" json:"display_name"`
	AliasSlug      string         `gorm:"column:alias_slug;type:varchar(255);not null;uniqueIndex:uk_product_categories_tenant_alias_slug,priority:2;comment:别名/slug" json:"alias_slug"`
	Path           string         `gorm:"type:text;not null;index:idx_product_categories_tenant_path,priority:2;comment:类目路径（形如 /<id>/<id>/ ）" json:"path"`
	Level          int            `gorm:"not null;default:0;comment:层级深度" json:"level"`
	SortOrder      int            `gorm:"not null;default:0;comment:同级排序" json:"sort_order"`
	Status         string         `gorm:"type:varchar(32);not null;default:'enabled';index;comment:状态" json:"status"`
	TemplateID     *string        `gorm:"column:template_id;type:uuid;index;comment:绑定模板ID" json:"template_id,omitempty"`
	SEOTitle       string         `gorm:"column:seo_title;type:varchar(255);comment:SEO 标题" json:"seo_title,omitempty"`
	SEODescription string         `gorm:"column:seo_description;type:text;comment:SEO 描述" json:"seo_description,omitempty"`
	SEOKeywords    string         `gorm:"column:seo_keywords;type:text;comment:SEO 关键词" json:"seo_keywords,omitempty"`
	ImageURL       string         `gorm:"column:image_url;type:text;comment:展示图" json:"image_url,omitempty"`
	IsFeatured     bool           `gorm:"column:is_featured;not null;default:false;comment:是否推荐位" json:"is_featured"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductCategory) TableName() string { return models.S(models.TableProductCategories) }

func (c *ProductCategory) Normalize() {
	if c == nil {
		return
	}
	c.TenantUUID = strings.TrimSpace(c.TenantUUID)
	c.Code = strings.TrimSpace(c.Code)
	c.DisplayName = strings.TrimSpace(c.DisplayName)
	c.AliasSlug = strings.TrimSpace(c.AliasSlug)
	c.Path = strings.TrimSpace(c.Path)
	c.Status = strings.TrimSpace(strings.ToLower(c.Status))
}

func (c *ProductCategory) IsEnabled() bool {
	if c == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(c.Status), CategoryStatusEnabled)
}
