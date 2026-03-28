package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SettlementBatch stores reconciliation batch metadata.
type SettlementBatch struct {
	ID                string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID        string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_settlement_batch_no,priority:1" json:"tenant_uuid"`
	BatchNo           string         `gorm:"column:batch_no;type:varchar(64);not null;uniqueIndex:uk_logistics_settlement_batch_no,priority:2" json:"batch_no"`
	CarrierID         string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id"`
	Status            string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	WaybillCount      int            `gorm:"column:waybill_count;not null;default:0" json:"waybill_count"`
	DiffCount         int            `gorm:"column:diff_count;not null;default:0" json:"diff_count"`
	TotalExpectedFee  float64        `gorm:"column:total_expected_fee;not null;default:0" json:"total_expected_fee"`
	TotalActualFee    float64        `gorm:"column:total_actual_fee;not null;default:0" json:"total_actual_fee"`
	TotalDiffAmount   float64        `gorm:"column:total_diff_amount;not null;default:0" json:"total_diff_amount"`
	SuggestionSummary datatypes.JSON `gorm:"column:suggestion_summary;type:jsonb" json:"suggestion_summary,omitempty"`
	Metadata          datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	ConfirmedAt       *time.Time     `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SettlementBatch) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsSettlementBatches)
}

// SettlementDiff stores per-waybill reconciliation attribution.
type SettlementDiff struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_settlement_diff,priority:1" json:"tenant_uuid"`
	BatchID       string         `gorm:"column:batch_id;type:uuid;not null;index;uniqueIndex:uk_logistics_settlement_diff,priority:2" json:"batch_id"`
	WaybillID     string         `gorm:"column:waybill_id;type:uuid;not null;index;uniqueIndex:uk_logistics_settlement_diff,priority:3" json:"waybill_id"`
	WaybillNo     string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no"`
	CarrierID     string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id"`
	ExpectedFee   float64        `gorm:"column:expected_fee;not null;default:0" json:"expected_fee"`
	ActualFee     float64        `gorm:"column:actual_fee;not null;default:0" json:"actual_fee"`
	DiffAmount    float64        `gorm:"column:diff_amount;not null;default:0" json:"diff_amount"`
	Attribution   string         `gorm:"column:attribution;type:varchar(64);not null;default:'matched';index" json:"attribution"`
	Suggestion    string         `gorm:"column:suggestion;type:varchar(64);not null;default:'accept';index" json:"suggestion"`
	Status        string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	HandledAction string         `gorm:"column:handled_action;type:varchar(32)" json:"handled_action,omitempty"`
	HandledNote   string         `gorm:"column:handled_note;type:text" json:"handled_note,omitempty"`
	HandledBy     string         `gorm:"column:handled_by;type:varchar(64)" json:"handled_by,omitempty"`
	HandledAt     *time.Time     `gorm:"column:handled_at" json:"handled_at,omitempty"`
	Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SettlementDiff) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsSettlementDiffs)
}
