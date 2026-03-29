package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// QualityAuditReport stores periodic fulfillment quality retrospective snapshots.
type QualityAuditReport struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:uuid;not null;index;index:idx_logistics_quality_audit_scope,priority:1" json:"tenant_uuid"`
	ReportPeriodFrom    time.Time      `gorm:"column:report_period_from;not null;index" json:"report_period_from"`
	ReportPeriodTo      time.Time      `gorm:"column:report_period_to;not null;index" json:"report_period_to"`
	WindowHours         int            `gorm:"column:window_hours;not null;default:24;index" json:"window_hours"`
	CarrierID           string         `gorm:"column:carrier_id;type:uuid;index;index:idx_logistics_quality_audit_scope,priority:2" json:"carrier_id,omitempty"`
	WarehouseID         string         `gorm:"column:warehouse_id;type:varchar(64);index;index:idx_logistics_quality_audit_scope,priority:3" json:"warehouse_id,omitempty"`
	DestinationZone     string         `gorm:"column:destination_zone;type:varchar(64);index;index:idx_logistics_quality_audit_scope,priority:4" json:"destination_zone,omitempty"`
	TotalWaybills       int            `gorm:"column:total_waybills;not null;default:0" json:"total_waybills"`
	DeliveredCount      int            `gorm:"column:delivered_count;not null;default:0" json:"delivered_count"`
	ExceptionCount      int            `gorm:"column:exception_count;not null;default:0" json:"exception_count"`
	TimeoutCount        int            `gorm:"column:timeout_count;not null;default:0" json:"timeout_count"`
	OnTimeRate          float64        `gorm:"column:on_time_rate;type:numeric(8,2);not null;default:0" json:"on_time_rate"`
	DeliverySuccessRate float64        `gorm:"column:delivery_success_rate;type:numeric(8,2);not null;default:0" json:"delivery_success_rate"`
	TotalCost           float64        `gorm:"column:total_cost;type:numeric(12,2);not null;default:0" json:"total_cost"`
	AvgCost             float64        `gorm:"column:avg_cost;type:numeric(10,2);not null;default:0" json:"avg_cost"`
	ForecastCount       int            `gorm:"column:forecast_count;not null;default:0" json:"forecast_count"`
	RootCauseCount      int            `gorm:"column:root_cause_count;not null;default:0" json:"root_cause_count"`
	InterwarehouseCount int            `gorm:"column:interwarehouse_count;not null;default:0" json:"interwarehouse_count"`
	Conclusion          string         `gorm:"column:conclusion;type:text" json:"conclusion,omitempty"`
	ActionItems         datatypes.JSON `gorm:"column:action_items;type:jsonb" json:"action_items,omitempty"`
	Status              string         `gorm:"column:status;type:varchar(32);not null;default:'generated';index" json:"status"`
	GeneratedBy         string         `gorm:"column:generated_by;type:varchar(64)" json:"generated_by,omitempty"`
	GeneratedAt         *time.Time     `gorm:"column:generated_at;index" json:"generated_at,omitempty"`
	Metadata            datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (QualityAuditReport) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsQualityAuditReports)
}
