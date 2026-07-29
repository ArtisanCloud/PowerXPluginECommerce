package product_category

import (
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	SaleSpecStatusActive   = "active"
	SaleSpecStatusDisabled = "disabled"
)

// CategorySaleSpecGroup defines a reusable sales spec dimension for a category.
type CategorySaleSpecGroup struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_category_sale_spec_groups,priority:1" json:"tenant_uuid"`
	CategoryID  string         `gorm:"column:category_id;type:uuid;not null;index;uniqueIndex:uk_category_sale_spec_groups,priority:2" json:"category_id"`
	Code        string         `gorm:"type:varchar(120);not null;uniqueIndex:uk_category_sale_spec_groups,priority:3" json:"code"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	SortOrder   int            `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Required    bool           `gorm:"not null;default:true" json:"required"`
	AllowCustom bool           `gorm:"column:allow_custom;not null;default:false" json:"allow_custom"`
	Status      string         `gorm:"type:varchar(32);not null;default:'active';index" json:"status"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategorySaleSpecGroup) TableName() string {
	return models.S(models.TableProductCategorySaleSpecGroups)
}

func (g *CategorySaleSpecGroup) Normalize() {
	if g == nil {
		return
	}
	g.TenantUUID = strings.TrimSpace(g.TenantUUID)
	g.CategoryID = strings.TrimSpace(g.CategoryID)
	g.Code = strings.TrimSpace(g.Code)
	g.Name = strings.TrimSpace(g.Name)
	g.Status = strings.TrimSpace(strings.ToLower(g.Status))
}

// CategorySaleSpecOption defines an allowed sales spec value under a category spec group.
type CategorySaleSpecOption struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_category_sale_spec_options,priority:1" json:"tenant_uuid"`
	CategoryID string         `gorm:"column:category_id;type:uuid;not null;index" json:"category_id"`
	GroupID    string         `gorm:"column:group_id;type:uuid;not null;index;uniqueIndex:uk_category_sale_spec_options,priority:2" json:"group_id"`
	Code       string         `gorm:"type:varchar(120);not null;uniqueIndex:uk_category_sale_spec_options,priority:3" json:"code"`
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	SortOrder  int            `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Meta       datatypes.JSON `gorm:"type:jsonb" json:"meta,omitempty"`
	Status     string         `gorm:"type:varchar(32);not null;default:'active';index" json:"status"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategorySaleSpecOption) TableName() string {
	return models.S(models.TableProductCategorySaleSpecOptions)
}

func (o *CategorySaleSpecOption) Normalize() {
	if o == nil {
		return
	}
	o.TenantUUID = strings.TrimSpace(o.TenantUUID)
	o.CategoryID = strings.TrimSpace(o.CategoryID)
	o.GroupID = strings.TrimSpace(o.GroupID)
	o.Code = strings.TrimSpace(o.Code)
	o.Name = strings.TrimSpace(o.Name)
	o.Status = strings.TrimSpace(strings.ToLower(o.Status))
}
