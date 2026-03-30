package logistics

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CrossborderDocument struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID  string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_crossborder_doc,priority:1" json:"tenant_uuid"`
	WaybillID   string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	WaybillNo   string         `gorm:"column:waybill_no;type:varchar(128);not null;index;uniqueIndex:uk_logistics_crossborder_doc,priority:2" json:"waybill_no"`
	DocType     string         `gorm:"column:doc_type;type:varchar(64);not null;uniqueIndex:uk_logistics_crossborder_doc,priority:3" json:"doc_type"`
	DocNo       string         `gorm:"column:doc_no;type:varchar(128);not null" json:"doc_no"`
	CountryFrom string         `gorm:"column:country_from;type:varchar(32);not null;default:''" json:"country_from"`
	CountryTo   string         `gorm:"column:country_to;type:varchar(32);not null;default:''" json:"country_to"`
	Status      string         `gorm:"column:status;type:varchar(32);not null;default:'pending';index" json:"status"`
	ValidatedAt *time.Time     `gorm:"column:validated_at" json:"validated_at,omitempty"`
	Metadata    datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CrossborderDocument) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCrossborderDocuments)
}

type CrossborderTaxQuote struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_crossborder_tax_quote,priority:1" json:"tenant_uuid"`
	RequestKey       string         `gorm:"column:request_key;type:varchar(128);not null;uniqueIndex:uk_logistics_crossborder_tax_quote,priority:2" json:"request_key"`
	WaybillID        string         `gorm:"column:waybill_id;type:uuid;index" json:"waybill_id,omitempty"`
	WaybillNo        string         `gorm:"column:waybill_no;type:varchar(128);index" json:"waybill_no,omitempty"`
	Destination      string         `gorm:"column:destination_country;type:varchar(32);not null;default:'';index" json:"destination_country"`
	Currency         string         `gorm:"column:currency;type:varchar(16);not null;default:'USD'" json:"currency"`
	DeclaredValue    float64        `gorm:"column:declared_value;type:numeric(12,2);not null;default:0" json:"declared_value"`
	ShippingFee      float64        `gorm:"column:shipping_fee;type:numeric(12,2);not null;default:0" json:"shipping_fee"`
	InsuranceFee     float64        `gorm:"column:insurance_fee;type:numeric(12,2);not null;default:0" json:"insurance_fee"`
	ExemptionAmount  float64        `gorm:"column:exemption_amount;type:numeric(12,2);not null;default:0" json:"exemption_amount"`
	DutyRate         float64        `gorm:"column:duty_rate;type:numeric(8,4);not null;default:0" json:"duty_rate"`
	VatRate          float64        `gorm:"column:vat_rate;type:numeric(8,4);not null;default:0" json:"vat_rate"`
	DutyAmount       float64        `gorm:"column:duty_amount;type:numeric(12,2);not null;default:0" json:"duty_amount"`
	VatAmount        float64        `gorm:"column:vat_amount;type:numeric(12,2);not null;default:0" json:"vat_amount"`
	TotalTaxAmount   float64        `gorm:"column:total_tax_amount;type:numeric(12,2);not null;default:0" json:"total_tax_amount"`
	QuoteProvider    string         `gorm:"column:quote_provider;type:varchar(64);not null;default:'rule-engine'" json:"quote_provider"`
	NormalizedStatus string         `gorm:"column:normalized_status;type:varchar(32);not null;default:'estimated'" json:"normalized_status"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CrossborderTaxQuote) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCrossborderTaxQuotes)
}

type CrossborderTrackingMap struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID       string         `gorm:"column:tenant_uuid;type:uuid;not null;index;uniqueIndex:uk_logistics_crossborder_tracking_map,priority:1" json:"tenant_uuid"`
	Provider         string         `gorm:"column:provider;type:varchar(64);not null;index;uniqueIndex:uk_logistics_crossborder_tracking_map,priority:2" json:"provider"`
	ProviderStatus   string         `gorm:"column:provider_status;type:varchar(64);not null;uniqueIndex:uk_logistics_crossborder_tracking_map,priority:3" json:"provider_status"`
	NormalizedStatus string         `gorm:"column:normalized_status;type:varchar(32);not null;index" json:"normalized_status"`
	Description      string         `gorm:"column:description;type:varchar(255)" json:"description,omitempty"`
	Priority         int            `gorm:"column:priority;not null;default:100;index" json:"priority"`
	Enabled          bool           `gorm:"column:enabled;not null;default:true;index" json:"enabled"`
	Metadata         datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CrossborderTrackingMap) TableName() string {
	return BaseModel.S(BaseModel.TableLogisticsCrossborderTrackingMaps)
}
