package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// PricebookScope defines the applicability dimension/value for a pricebook.
type PricebookScope struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:Scope ID" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_pricebook_scope,priority:1;comment:租户" json:"tenant_uuid"`
	PricebookID string         `gorm:"column:pricebook_id;type:uuid;not null;index;uniqueIndex:uk_pricebook_scope,priority:2;comment:价目表ID" json:"pricebook_id"`
	Dimension   string         `gorm:"type:text;not null;uniqueIndex:uk_pricebook_scope,priority:3;comment:维度(channel/customer_group/supplier)" json:"dimension"`
	DimensionID string         `gorm:"column:dimension_id;type:uuid;not null;uniqueIndex:uk_pricebook_scope,priority:4;comment:维度ID" json:"dimension_id"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName implements gorm.Tabler.
func (PricebookScope) TableName() string { return models.S(models.TablePricebookScopes) }
