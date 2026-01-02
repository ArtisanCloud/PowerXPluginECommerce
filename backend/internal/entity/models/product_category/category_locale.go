package product_category

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// CategoryLocale stores localized category labels.
type CategoryLocale struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CategoryID  string         `gorm:"column:category_id;type:uuid;not null;index:idx_category_locale_unique,priority:1" json:"category_id"`
	Locale      string         `gorm:"type:varchar(16);not null;index:idx_category_locale_unique,priority:2" json:"locale"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoryLocale) TableName() string { return models.S(models.TableProductCategoryLocales) }
