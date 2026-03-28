package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ReconciliationBatch stores triple-check reconciliation execution batch.
type ReconciliationBatch struct {
	ID                 string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID         string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_reconciliation_batch_no,priority:1" json:"tenant_uuid"`
	BatchNo            string         `gorm:"column:batch_no;type:varchar(64);not null;uniqueIndex:uk_logistics_reconciliation_batch_no,priority:2" json:"batch_no"`
	CarrierID          string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id"`
	Status             string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	RecordCount        int            `gorm:"column:record_count;not null;default:0" json:"record_count"`
	MatchedCount       int            `gorm:"column:matched_count;not null;default:0" json:"matched_count"`
	ExceptionCount     int            `gorm:"column:exception_count;not null;default:0" json:"exception_count"`
	TotalBillAmount    float64        `gorm:"column:total_bill_amount;not null;default:0" json:"total_bill_amount"`
	TotalBankAmount    float64        `gorm:"column:total_bank_amount;not null;default:0" json:"total_bank_amount"`
	TotalInvoiceAmount float64        `gorm:"column:total_invoice_amount;not null;default:0" json:"total_invoice_amount"`
	Summary            datatypes.JSON `gorm:"column:summary;type:jsonb" json:"summary,omitempty"`
	Metadata           datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	ExecutedAt         *time.Time     `gorm:"column:executed_at" json:"executed_at,omitempty"`
	ConfirmedAt        *time.Time     `gorm:"column:confirmed_at" json:"confirmed_at,omitempty"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReconciliationBatch) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsReconciliationBatches)
}

// ReconciliationRecord stores matching result of bill/bank/invoice row.
type ReconciliationRecord struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID    string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_reconciliation_record,priority:1" json:"tenant_uuid"`
	BatchID       string         `gorm:"column:batch_id;type:uuid;not null;index;uniqueIndex:uk_logistics_reconciliation_record,priority:2" json:"batch_id"`
	WaybillID     string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id"`
	WaybillNo     string         `gorm:"column:waybill_no;type:varchar(128);index;uniqueIndex:uk_logistics_reconciliation_record,priority:3" json:"waybill_no"`
	CarrierID     string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id"`
	BillAmount    float64        `gorm:"column:bill_amount;not null;default:0" json:"bill_amount"`
	BankAmount    float64        `gorm:"column:bank_amount;not null;default:0" json:"bank_amount"`
	InvoiceAmount float64        `gorm:"column:invoice_amount;not null;default:0" json:"invoice_amount"`
	DiffAmount    float64        `gorm:"column:diff_amount;not null;default:0" json:"diff_amount"`
	MatchType     string         `gorm:"column:match_type;type:varchar(64);not null;default:'matched';index" json:"match_type"`
	Suggestion    string         `gorm:"column:suggestion;type:varchar(64);not null;default:'auto_archive';index" json:"suggestion"`
	Status        string         `gorm:"column:status;type:varchar(32);not null;default:'resolved';index" json:"status"`
	CaseID        string         `gorm:"column:case_id;type:uuid;index" json:"case_id,omitempty"`
	HandledAction string         `gorm:"column:handled_action;type:varchar(32)" json:"handled_action,omitempty"`
	HandledNote   string         `gorm:"column:handled_note;type:text" json:"handled_note,omitempty"`
	HandledBy     string         `gorm:"column:handled_by;type:varchar(64)" json:"handled_by,omitempty"`
	HandledAt     *time.Time     `gorm:"column:handled_at" json:"handled_at,omitempty"`
	Metadata      datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReconciliationRecord) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsReconciliationRecords)
}

// ReconciliationCase stores exception workorder for manual handling.
type ReconciliationCase struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_reconciliation_case_no,priority:1" json:"tenant_uuid"`
	CaseNo     string         `gorm:"column:case_no;type:varchar(64);not null;uniqueIndex:uk_logistics_reconciliation_case_no,priority:2" json:"case_no"`
	BatchID    string         `gorm:"column:batch_id;type:uuid;not null;index" json:"batch_id"`
	RecordID   string         `gorm:"column:record_id;type:uuid;not null;index" json:"record_id"`
	CarrierID  string         `gorm:"column:carrier_id;type:uuid;index" json:"carrier_id"`
	WaybillNo  string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no"`
	Status     string         `gorm:"column:status;type:varchar(32);not null;default:'open';index" json:"status"`
	Reason     string         `gorm:"column:reason;type:varchar(128);not null;default:'amount_mismatch'" json:"reason"`
	Suggestion string         `gorm:"column:suggestion;type:varchar(64);not null;default:'manual_review'" json:"suggestion"`
	ActionNote string         `gorm:"column:action_note;type:text" json:"action_note,omitempty"`
	HandledBy  string         `gorm:"column:handled_by;type:varchar(64)" json:"handled_by,omitempty"`
	HandledAt  *time.Time     `gorm:"column:handled_at" json:"handled_at,omitempty"`
	Metadata   datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReconciliationCase) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsReconciliationCases)
}
