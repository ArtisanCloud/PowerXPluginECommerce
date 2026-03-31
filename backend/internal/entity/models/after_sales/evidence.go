package after_sales

import (
	"time"

	BaseModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
)

// AfterSaleEvidence stores customer/operator submitted evidence references.
type AfterSaleEvidence struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantUUID   string    `gorm:"column:tenant_uuid;type:uuid;not null;index" json:"tenant_uuid"`
	CaseID       string    `gorm:"column:case_id;type:uuid;not null;index" json:"case_id"`
	UploaderType string    `gorm:"column:uploader_type;type:varchar(32);not null" json:"uploader_type"`
	UploaderID   string    `gorm:"column:uploader_id;type:varchar(128)" json:"uploader_id,omitempty"`
	EvidenceType string    `gorm:"column:evidence_type;type:varchar(32);not null" json:"evidence_type"`
	ContentRef   string    `gorm:"column:content_ref;type:text;not null" json:"content_ref"`
	Description  string    `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AfterSaleEvidence) TableName() string {
	return BaseModel.S(BaseModel.TableAfterSalesEvidences)
}
