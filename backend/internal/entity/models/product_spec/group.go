package product_spec

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// ProductSpecGroup represents a spec dimension (e.g. color/size) scoped to an SPU.
type ProductSpecGroup struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"type:uuid;not null;index;uniqueIndex:uk_product_spec_groups,priority:1" json:"tenant_uuid"`
	SPUID      string         `gorm:"column:spu_id;type:uuid;not null;index;uniqueIndex:uk_product_spec_groups,priority:2" json:"spu_id"`
	Code       string         `gorm:"type:varchar(120);not null;uniqueIndex:uk_product_spec_groups,priority:3" json:"code"`
	Name       string         `gorm:"type:varchar(255);not null" json:"name"`
	SortOrder  int            `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	Required   bool           `gorm:"not null;default:true" json:"required"`
	Status     string         `gorm:"type:varchar(32);not null;default:'active';index" json:"status"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductSpecGroup) TableName() string {
	return models.S(models.TableProductSpecGroups)
}

