package models

import (
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelConfig stores per-channel strategy & financial configuration.
type ChannelConfig struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"type:uuid;not null;index:idx_channel_config_tenant_channel,priority:1" json:"tenant_uuid"`
	ChannelID           string         `gorm:"type:uuid;not null;index:idx_channel_config_tenant_channel,priority:2" json:"channel_id"`
	PricebookID         string         `gorm:"type:varchar(128)" json:"pricebook_id,omitempty"`
	InventoryStrategyID string         `gorm:"type:varchar(128)" json:"inventory_strategy_id,omitempty"`
	LogisticsStrategyID string         `gorm:"type:varchar(128)" json:"logistics_strategy_id,omitempty"`
	CSSLAID             string         `gorm:"column:cs_sla_id;type:varchar(128)" json:"cs_sla_id,omitempty"`
	FeeRate             float64        `gorm:"type:numeric(6,3);default:0" json:"fee_rate"`
	SettlementCycle     string         `gorm:"type:varchar(64)" json:"settlement_cycle,omitempty"`
	PaymentTerms        string         `gorm:"type:varchar(128)" json:"payment_terms,omitempty"`
	Notes               string         `gorm:"type:text" json:"notes,omitempty"`
	EffectiveAt         *time.Time     `json:"effective_at,omitempty"`
	DeprecatedAt        *time.Time     `json:"deprecated_at,omitempty"`
	TeamMetadata        datatypes.JSON `gorm:"type:jsonb" json:"team_metadata,omitempty"`
	CreatedBy           string         `gorm:"type:varchar(64)" json:"created_by,omitempty"`
	UpdatedBy           string         `gorm:"type:varchar(64)" json:"updated_by,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the fully qualified table name.
func (ChannelConfig) TableName() string {
	return models.S(models.TableChannelConfigs)
}
