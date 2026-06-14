package promotion

import (
	"errors"
	"strings"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	TypeAmountOff  = "amount_off"
	TypePercentOff = "percent_off"

	StatusDraft   = "draft"
	StatusActive  = "active"
	StatusPaused  = "paused"
	StatusExpired = "expired"
)

// Campaign stores an automatic order promotion rule.
type Campaign struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_promotion_campaign_tenant_code,priority:1" json:"tenant_uuid"`
	Code          string         `gorm:"type:varchar(64);not null;uniqueIndex:uk_promotion_campaign_tenant_code,priority:2" json:"code"`
	Name          string         `gorm:"type:varchar(128);not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	PromotionType string         `gorm:"column:promotion_type;type:varchar(32);not null;index" json:"promotion_type"`
	ConditionRule datatypes.JSON `gorm:"column:condition_rule;type:jsonb;not null;default:'{}'" json:"condition_rule"`
	ScopeRule     datatypes.JSON `gorm:"column:scope_rule;type:jsonb;not null;default:'{}'" json:"scope_rule"`
	ActionRule    datatypes.JSON `gorm:"column:action_rule;type:jsonb;not null;default:'{}'" json:"action_rule"`
	StackingRule  datatypes.JSON `gorm:"column:stacking_rule;type:jsonb;not null;default:'{}'" json:"stacking_rule"`
	ValidFrom     time.Time      `gorm:"column:valid_from;type:timestamptz;not null;index" json:"valid_from"`
	ValidTo       time.Time      `gorm:"column:valid_to;type:timestamptz;not null;index" json:"valid_to"`
	Status        string         `gorm:"type:varchar(16);not null;default:'draft';index" json:"status"`
	CreatedBy     string         `gorm:"column:created_by;type:varchar(128)" json:"created_by,omitempty"`
	UpdatedBy     string         `gorm:"column:updated_by;type:varchar(128)" json:"updated_by,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Campaign) TableName() string { return models.S(models.TablePromotionCampaigns) }

func (m *Campaign) BeforeSave(_ *gorm.DB) error {
	if m == nil {
		return nil
	}
	if strings.TrimSpace(m.TenantUUID) == "" {
		return errors.New("tenant uuid is required")
	}
	if strings.TrimSpace(m.Code) == "" {
		return errors.New("promotion code is required")
	}
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("promotion name is required")
	}
	switch strings.TrimSpace(m.PromotionType) {
	case TypeAmountOff, TypePercentOff:
	default:
		return errors.New("promotion type is invalid")
	}
	if m.ValidFrom.After(m.ValidTo) {
		return errors.New("promotion valid_from must be before or equal to valid_to")
	}
	return nil
}
