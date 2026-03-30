package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AllocationDecision stores automatic/manual carrier allocation records.
type AllocationDecision struct {
	ID                string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID        string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_allocation_request,priority:1" json:"tenant_uuid"`
	RequestKey        string         `gorm:"column:request_key;type:varchar(128);not null;uniqueIndex:uk_logistics_allocation_request,priority:2" json:"request_key"`
	WaybillID         string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	OrderID           string         `gorm:"column:order_id;type:uuid;index" json:"order_id,omitempty"`
	CarrierID         string         `gorm:"column:carrier_id;type:uuid;not null;index" json:"carrier_id"`
	WarehouseID       string         `gorm:"column:warehouse_id;type:varchar(64);index" json:"warehouse_id,omitempty"`
	DestinationZone   string         `gorm:"column:destination_zone;type:varchar(64);index" json:"destination_zone,omitempty"`
	Strategy          string         `gorm:"column:strategy;type:varchar(64);not null;default:'capacity_first'" json:"strategy"`
	Reason            string         `gorm:"column:reason;type:varchar(255)" json:"reason,omitempty"`
	ManualOverride    bool           `gorm:"column:manual_override;not null;default:false;index" json:"manual_override"`
	PreviousCarrierID string         `gorm:"column:previous_carrier_id;type:uuid;index" json:"previous_carrier_id,omitempty"`
	Candidates        datatypes.JSON `gorm:"column:candidates;type:jsonb" json:"candidates,omitempty"`
	Metadata          datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedBy         string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AllocationDecision) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsAllocationDecisions)
}
