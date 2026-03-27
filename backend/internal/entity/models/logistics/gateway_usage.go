package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// GatewayUsage records gateway call/cost/quota consumption per sync job.
type GatewayUsage struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_gateway_usage_job,priority:1" json:"tenant_uuid"`
	CarrierID     string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	Provider      string         `gorm:"column:provider;type:varchar(64);index" json:"provider,omitempty"`
	SourceJobID   string         `gorm:"column:source_job_id;type:uuid;index;uniqueIndex:uk_logistics_gateway_usage_job,priority:2" json:"source_job_id,omitempty"`
	RequestCount  int            `gorm:"column:request_count;not null;default:0" json:"request_count"`
	SuccessCount  int            `gorm:"column:success_count;not null;default:0" json:"success_count"`
	FailedCount   int            `gorm:"column:failed_count;not null;default:0" json:"failed_count"`
	UnitPrice     float64        `gorm:"column:unit_price;type:numeric(18,6);not null;default:0" json:"unit_price"`
	CostAmount    float64        `gorm:"column:cost_amount;type:numeric(18,6);not null;default:0" json:"cost_amount"`
	QuotaConsumed int            `gorm:"column:quota_consumed;not null;default:0" json:"quota_consumed"`
	WindowStartAt time.Time      `gorm:"column:window_start_at;index" json:"window_start_at"`
	WindowEndAt   time.Time      `gorm:"column:window_end_at;index" json:"window_end_at"`
	Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GatewayUsage) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsGatewayUsages)
}
