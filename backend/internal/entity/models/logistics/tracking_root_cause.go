package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// TrackingRootCause stores waybill anomaly attribution and remediation lifecycle.
type TrackingRootCause struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID      string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_root_cause_scope,priority:1;uniqueIndex:uk_logistics_root_cause_waybill_type,priority:1" json:"tenant_uuid"`
	WaybillID       string         `gorm:"column:waybill_id;type:uuid;not null;index;index:idx_logistics_root_cause_scope,priority:2;uniqueIndex:uk_logistics_root_cause_waybill_type,priority:2" json:"waybill_id"`
	WaybillNo       string         `gorm:"column:waybill_no;type:varchar(128);not null;index" json:"waybill_no"`
	CarrierID       string         `gorm:"column:carrier_id;type:uuid;index;index:idx_logistics_root_cause_scope,priority:3" json:"carrier_id,omitempty"`
	WarehouseID     string         `gorm:"column:warehouse_id;type:varchar(64);index;index:idx_logistics_root_cause_scope,priority:4" json:"warehouse_id,omitempty"`
	DestinationZone string         `gorm:"column:destination_zone;type:varchar(64);index;index:idx_logistics_root_cause_scope,priority:5" json:"destination_zone,omitempty"`
	AnomalyType     string         `gorm:"column:anomaly_type;type:varchar(64);not null;index;uniqueIndex:uk_logistics_root_cause_waybill_type,priority:3" json:"anomaly_type"`
	OwnerType       string         `gorm:"column:owner_type;type:varchar(32);not null;default:'unknown';index" json:"owner_type"`
	Severity        string         `gorm:"column:severity;type:varchar(32);not null;default:'medium';index" json:"severity"`
	SuggestedAction string         `gorm:"column:suggested_action;type:varchar(64);not null;default:'manual_review'" json:"suggested_action"`
	SelectedAction  string         `gorm:"column:selected_action;type:varchar(64)" json:"selected_action,omitempty"`
	Status          string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	EvidenceChain   datatypes.JSON `gorm:"column:evidence_chain;type:jsonb" json:"evidence_chain,omitempty"`
	ResultNote      string         `gorm:"column:result_note;type:text" json:"result_note,omitempty"`
	HandledBy       string         `gorm:"column:handled_by;type:varchar(64)" json:"handled_by,omitempty"`
	DetectedAt      *time.Time     `gorm:"column:detected_at;index" json:"detected_at,omitempty"`
	HandledAt       *time.Time     `gorm:"column:handled_at" json:"handled_at,omitempty"`
	Metadata        datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TrackingRootCause) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsTrackingRootCauses)
}
