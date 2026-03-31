package after_sales

import (
	"context"
	"strings"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"gorm.io/gorm"
)

// CaseRepository persists after-sale cases.
type CaseRepository struct {
	*Repository[AfterSalesModel.AfterSaleCase]
}

func NewCaseRepository(db *gorm.DB) *CaseRepository {
	return &CaseRepository{Repository: NewRepository[AfterSalesModel.AfterSaleCase](db)}
}

func (r *CaseRepository) Create(ctx context.Context, row *AfterSalesModel.AfterSaleCase) error {
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
