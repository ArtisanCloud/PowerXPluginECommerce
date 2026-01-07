package product_category

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// CategoryPermission scopes which principals can operate on categories.
type CategoryPermission struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	PrincipalType string         `gorm:"column:principal_type;type:varchar(32);not null;index" json:"principal_type"`
	PrincipalID   string         `gorm:"column:principal_id;type:varchar(64);not null;index" json:"principal_id"`
	CategoryID    string         `gorm:"column:category_id;type:uuid;not null;index" json:"category_id"`
	Action        string         `gorm:"type:varchar(32);not null;comment:read/manage/template/mapping/import" json:"action"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CategoryPermission) TableName() string { return models.S(models.TableProductCategoryPermissions) }
