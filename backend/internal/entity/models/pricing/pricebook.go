package models

import (
	"errors"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/gorm"
)

// Pricebook represents a tenant-scoped pricebook header.
type Pricebook struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();comment:价目表ID" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_pricebook_tenant_code,priority:1;comment:租户" json:"tenant_uuid"`
	Code             string         `gorm:"type:text;not null;uniqueIndex:uk_pricebook_tenant_code,priority:2;comment:业务编码" json:"code"`
	Name             string         `gorm:"type:text;not null;comment:名称" json:"name"`
	Type             string         `gorm:"type:text;not null;default:'sales';index;comment:类型(sales/purchase)" json:"type"`
	Currency         string         `gorm:"type:text;not null;index;comment:币种(ISO4217)" json:"currency"`
	Description      string         `gorm:"type:text;comment:描述" json:"description,omitempty"`
	Status           string         `gorm:"type:text;not null;default:'active';index;comment:状态(active/archived)" json:"status"`
	CurrentVersionID *string        `gorm:"column:current_version_id;type:uuid;index;comment:当前版本" json:"current_version_id,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName implements gorm.Tabler.
func (Pricebook) TableName() string { return models.S(models.TablePricebooks) }

func (pb *Pricebook) BeforeDelete(tx *gorm.DB) error {
	if pb == nil {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(pb.Code), "base") {
		return errors.New("base pricebook cannot be deleted")
	}
	return nil
}
