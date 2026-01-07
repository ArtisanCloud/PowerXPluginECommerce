package product_category

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CategoryTemplateField represents a single field definition inside a template.
type CategoryTemplateField struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_category_template_fields,priority:1" json:"tenant_uuid"`
	TemplateID      string         `gorm:"column:template_id;type:uuid;not null;index;uniqueIndex:uk_category_template_fields,priority:2" json:"template_id"`
	FieldKey        string         `gorm:"column:field_key;type:varchar(120);not null;uniqueIndex:uk_category_template_fields,priority:3" json:"field_key"`
	Group           string         `gorm:"type:varchar(64);comment:字段分组" json:"group,omitempty"`
	FieldType       string         `gorm:"column:field_type;type:varchar(64);not null" json:"field_type"`
	Required        bool           `gorm:"not null;default:false" json:"required"`
	SortOrder       int            `gorm:"column:sort_order;not null;default:0;comment:字段排序" json:"sort_order"`
	ValidationRules datatypes.JSON `gorm:"column:validation_rules;type:jsonb" json:"validation_rules,omitempty"`
	DefaultValue    datatypes.JSON `gorm:"column:default_value;type:jsonb" json:"default_value,omitempty"`
	I18nLabel       datatypes.JSON `gorm:"column:i18n_label;type:jsonb" json:"i18n_label,omitempty"`
	I18nHelp        datatypes.JSON `gorm:"column:i18n_help;type:jsonb" json:"i18n_help,omitempty"`
	Inheritable     bool           `gorm:"not null;default:true" json:"inheritable"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoryTemplateField) TableName() string {
	return models.S(models.TableProductCategoryTemplateFields)
}
