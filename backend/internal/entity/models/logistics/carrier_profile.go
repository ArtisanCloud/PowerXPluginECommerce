package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CarrierProfile aggregates service quality and governance status per carrier.
type CarrierProfile struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID     string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_carrier_profile_carrier,priority:1" json:"tenant_uuid"`
	CarrierID      string         `gorm:"column:carrier_id;type:uuid;not null;index;uniqueIndex:uk_logistics_carrier_profile_carrier,priority:2" json:"carrier_id"`
	StabilityScore float64        `gorm:"column:stability_score;type:numeric(6,2);not null;default:0" json:"stability_score"`
	CostScore      float64        `gorm:"column:cost_score;type:numeric(6,2);not null;default:0" json:"cost_score"`
	ServiceRating  string         `gorm:"column:service_rating;type:varchar(16);not null;default:'B';index" json:"service_rating"`
	CompositeScore float64        `gorm:"column:composite_score;type:numeric(6,2);not null;default:0;index" json:"composite_score"`
	Status         string         `gorm:"column:status;type:varchar(32);not null;default:'active';index" json:"status"`
	RetireReason   string         `gorm:"column:retire_reason;type:text" json:"retire_reason,omitempty"`
	ScoreTrend     datatypes.JSON `gorm:"column:score_trend;type:jsonb" json:"score_trend,omitempty"`
	Suggestion     string         `gorm:"column:suggestion;type:text" json:"suggestion,omitempty"`
	ConfirmedBy    string         `gorm:"column:confirmed_by;type:varchar(64)" json:"confirmed_by,omitempty"`
	ConfirmedAt    *time.Time     `gorm:"column:confirmed_at;index" json:"confirmed_at,omitempty"`
	EvaluatedAt    *time.Time     `gorm:"column:evaluated_at;index" json:"evaluated_at,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CarrierProfile) TableName() string { return BaseModel.S(BaseModel.TableLogisticsCarrierProfiles) }
