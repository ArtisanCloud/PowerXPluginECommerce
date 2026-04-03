package coupon

import (
	"errors"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CouponTemplate defines coupon rules.
type CouponTemplate struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_coupon_template_tenant_code,priority:1" json:"tenant_uuid"`
	Code          string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_coupon_template_tenant_code,priority:2" json:"code"`
	Name          string         `gorm:"type:varchar(128);not null" json:"name"`
	CouponType    string         `gorm:"column:coupon_type;type:varchar(16);not null;index" json:"coupon_type"`
	ThresholdRule datatypes.JSON `gorm:"column:threshold_rule;type:jsonb;not null;default:'{}'" json:"threshold_rule"`
	ScopeRule     datatypes.JSON `gorm:"column:scope_rule;type:jsonb;not null;default:'{}'" json:"scope_rule"`
	StackingRule  datatypes.JSON `gorm:"column:stacking_rule;type:jsonb;not null;default:'{}'" json:"stacking_rule"`
	RefundRule    datatypes.JSON `gorm:"column:refund_rule;type:jsonb;not null;default:'{}'" json:"refund_rule"`
	ValidFrom     time.Time      `gorm:"column:valid_from;type:timestamptz;not null;index" json:"valid_from"`
	ValidTo       time.Time      `gorm:"column:valid_to;type:timestamptz;not null;index" json:"valid_to"`
	Status        string         `gorm:"type:varchar(16);not null;default:'draft';index" json:"status"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CouponTemplate) TableName() string { return models.S(models.TableCouponTemplates) }

func (m *CouponTemplate) BeforeSave(_ *gorm.DB) error {
	if m == nil {
		return nil
	}
	if strings.TrimSpace(m.Code) == "" {
		return errors.New("coupon template code is required")
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("coupon template name is required")
	}
	if strings.TrimSpace(m.CouponType) == "" {
		return errors.New("coupon template type is required")
	}
	if m.ValidFrom.After(m.ValidTo) {
		return errors.New("coupon template valid_from must be before or equal to valid_to")
	}
	return nil
}
