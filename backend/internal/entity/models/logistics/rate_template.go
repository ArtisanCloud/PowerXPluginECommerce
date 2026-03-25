package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RateTemplate is the top-level freight template.
type RateTemplate struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_rate_template,priority:1" json:"tenant_uuid"`
	Name       string         `gorm:"type:varchar(128);not null;uniqueIndex:uk_logistics_rate_template,priority:2" json:"name"`
	Currency   string         `gorm:"type:varchar(8);not null;default:'CNY'" json:"currency"`
	Status     string         `gorm:"type:varchar(32);not null;default:'draft';index" json:"status"`
	Version    int            `gorm:"not null;default:1" json:"version"`
	Channels   datatypes.JSON `gorm:"type:jsonb" json:"channels,omitempty"`
	Rules      datatypes.JSON `gorm:"type:jsonb" json:"rules,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RateTemplate) TableName() string { return BaseModel.S(BaseModel.TableLogisticsRateTemplates) }

// RateZone stores zone-level pricing details for a template.
type RateZone struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	TemplateID       string         `gorm:"column:template_id;type:uuid;not null;index" json:"template_id"`
	Region           string         `gorm:"type:varchar(128);not null" json:"region"`
	FirstWeight      float64        `gorm:"column:first_weight;type:numeric(12,3);not null;default:0" json:"first_weight"`
	FirstFee         float64        `gorm:"column:first_fee;type:numeric(12,2);not null;default:0" json:"first_fee"`
	AdditionalWeight float64        `gorm:"column:additional_weight;type:numeric(12,3);not null;default:0" json:"additional_weight"`
	AdditionalFee    float64        `gorm:"column:additional_fee;type:numeric(12,2);not null;default:0" json:"additional_fee"`
	FreeThreshold    float64        `gorm:"column:free_threshold;type:numeric(12,2);not null;default:0" json:"free_threshold"`
	Metadata         datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RateZone) TableName() string { return BaseModel.S(BaseModel.TableLogisticsRateZones) }
