package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AddressValidation struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_address_validation,priority:1" json:"tenant_uuid"`
	RequestKey       string         `gorm:"column:request_key;type:varchar(128);not null;index;uniqueIndex:uk_logistics_address_validation,priority:2" json:"request_key"`
	WaybillID        string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	WaybillNo        string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no,omitempty"`
	RawAddress       string         `gorm:"column:raw_address;type:text;not null" json:"raw_address"`
	Normalized       string         `gorm:"column:normalized;type:text" json:"normalized"`
	Reachable        bool           `gorm:"column:reachable;not null;index" json:"reachable"`
	RiskLevel        string         `gorm:"column:risk_level;type:varchar(32);not null;default:'low';index" json:"risk_level"`
	Suggestion       string         `gorm:"column:suggestion;type:text" json:"suggestion,omitempty"`
	NeedManualReview bool           `gorm:"column:need_manual_review;not null;default:false;index" json:"need_manual_review"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AddressValidation) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsAddressValidations)
}
