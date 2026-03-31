package after_sales

import (
	"context"
	"strings"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"gorm.io/gorm"
)

// EvidenceRepository persists after-sale evidence rows.
type EvidenceRepository struct {
	*Repository[AfterSalesModel.AfterSaleEvidence]
}

func NewEvidenceRepository(db *gorm.DB) *EvidenceRepository {
	return &EvidenceRepository{Repository: NewRepository[AfterSalesModel.AfterSaleEvidence](db)}
}

func (r *EvidenceRepository) Create(ctx context.Context, row *AfterSalesModel.AfterSaleEvidence) error {
	if row == nil {
		return gorm.ErrInvalidData
	}
	tenantUUID, err := RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(row.TenantUUID) == "" {
		row.TenantUUID = tenantUUID
	}
	return r.DB.WithContext(ctx).Create(row).Error
}
