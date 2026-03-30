package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// InterwarehouseAllocation stores candidate and confirmed inter-warehouse transfer decisions.
type InterwarehouseAllocation struct {
	ID                string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID        string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_interwarehouse_candidate,priority:1" json:"tenant_uuid"`
	RequestKey        string         `gorm:"column:request_key;type:varchar(128);not null;index;uniqueIndex:uk_logistics_interwarehouse_candidate,priority:2" json:"request_key"`
	WaybillID         string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	OrderID           string         `gorm:"column:order_id;type:uuid;index" json:"order_id,omitempty"`
	CarrierID         string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id,omitempty"`
	SourceWarehouseID string         `gorm:"column:source_warehouse_id;type:varchar(64);not null;index" json:"source_warehouse_id"`
	TargetWarehouseID string         `gorm:"column:target_warehouse_id;type:varchar(64);not null;index;uniqueIndex:uk_logistics_interwarehouse_candidate,priority:3" json:"target_warehouse_id"`
	DestinationZone   string         `gorm:"column:destination_zone;type:varchar(64);index" json:"destination_zone,omitempty"`
	TransferQty       int            `gorm:"column:transfer_qty;not null;default:0" json:"transfer_qty"`
	SourceAvailable   int            `gorm:"column:source_available;not null;default:0" json:"source_available"`
	TargetAvailable   int            `gorm:"column:target_available;not null;default:0" json:"target_available"`
	TransferCost      float64        `gorm:"column:transfer_cost;type:numeric(12,2);not null;default:0" json:"transfer_cost"`
	ETAImpactHours    float64        `gorm:"column:eta_impact_hours;type:numeric(8,2);not null;default:0" json:"eta_impact_hours"`
	Score             float64        `gorm:"column:score;type:numeric(8,4);not null;default:0;index" json:"score"`
	Strategy          string         `gorm:"column:strategy;type:varchar(64);not null;default:'collaboration_first'" json:"strategy"`
	Status            string         `gorm:"column:status;type:varchar(32);not null;default:'suggested';index" json:"status"`
	Reason            string         `gorm:"column:reason;type:text" json:"reason,omitempty"`
	ConfirmedBy       string         `gorm:"column:confirmed_by;type:varchar(64)" json:"confirmed_by,omitempty"`
	ConfirmedAt       *time.Time     `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	Metadata          datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (InterwarehouseAllocation) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsInterwarehouseAllocations)
}
