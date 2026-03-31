package subscription_reconciliation

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ReconciliationBatch records one reconciliation execution for a billing cycle.
type ReconciliationBatch struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_sr_batch,priority:1" json:"tenant_uuid"`
	BillingCycle        string         `gorm:"column:billing_cycle;type:varchar(32);not null;uniqueIndex:uk_sr_batch,priority:2" json:"billing_cycle"`
	RunType             string         `gorm:"column:run_type;type:varchar(16);not null;default:'daily';uniqueIndex:uk_sr_batch,priority:3" json:"run_type"`
	ExpectedAmountMinor int64          `gorm:"column:expected_amount_minor;not null;default:0" json:"expected_amount_minor"`
	ActualAmountMinor   int64          `gorm:"column:actual_amount_minor;not null;default:0" json:"actual_amount_minor"`
	DeltaAmountMinor    int64          `gorm:"column:delta_amount_minor;not null;default:0" json:"delta_amount_minor"`
	DeltaCount          int            `gorm:"column:delta_count;not null;default:0" json:"delta_count"`
	Status              string         `gorm:"column:status;type:varchar(24);not null;default:'running';index" json:"status"`
	StartedAt           *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	FinishedAt          *time.Time     `gorm:"column:finished_at" json:"finished_at,omitempty"`
	CreatedBy           string         `gorm:"column:created_by;type:varchar(64)" json:"created_by,omitempty"`
	Metadata            datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReconciliationBatch) TableName() string {
	return BaseModel.S(BaseModel.TableSubscriptionReconciliationBatches)
}

// ReconciliationDelta stores one mismatch item produced by a batch.
type ReconciliationDelta struct {
	ID                  string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID          string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_sr_delta_fp,priority:1" json:"tenant_uuid"`
	BatchID             string         `gorm:"column:batch_id;type:uuid;not null;index" json:"batch_id"`
	SubscriptionRef     string         `gorm:"column:subscription_ref;type:varchar(64);index" json:"subscription_ref,omitempty"`
	BillRef             string         `gorm:"column:bill_ref;type:varchar(64);index" json:"bill_ref,omitempty"`
	PaymentRef          string         `gorm:"column:payment_ref;type:varchar(64);index" json:"payment_ref,omitempty"`
	DeltaType           string         `gorm:"column:delta_type;type:varchar(48);not null;index" json:"delta_type"`
	RiskLevel           string         `gorm:"column:risk_level;type:varchar(16);not null;index" json:"risk_level"`
	ExpectedAmountMinor int64          `gorm:"column:expected_amount_minor;not null;default:0" json:"expected_amount_minor"`
	ActualAmountMinor   int64          `gorm:"column:actual_amount_minor;not null;default:0" json:"actual_amount_minor"`
	DeltaAmountMinor    int64          `gorm:"column:delta_amount_minor;not null;default:0" json:"delta_amount_minor"`
	ReasonCode          string         `gorm:"column:reason_code;type:varchar(64);index" json:"reason_code,omitempty"`
	Status              string         `gorm:"column:status;type:varchar(24);not null;default:'open';index" json:"status"`
	DeltaFingerprint    string         `gorm:"column:delta_fingerprint;type:varchar(128);not null;uniqueIndex:uk_sr_delta_fp,priority:2" json:"delta_fingerprint"`
	DetectedAt          *time.Time     `gorm:"column:detected_at" json:"detected_at,omitempty"`
	Metadata            datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ReconciliationDelta) TableName() string {
	return BaseModel.S(BaseModel.TableSubscriptionReconciliationDeltas)
}
