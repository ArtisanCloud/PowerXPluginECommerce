package after_sales

import (
	"context"
	"strings"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"gorm.io/gorm"
)

// ReverseLogisticsLinkRepository persists reverse-waybill links.
type ReverseLogisticsLinkRepository struct {
	*Repository[AfterSalesModel.ReturnLogisticsLink]
}

func NewReverseLogisticsLinkRepository(db *gorm.DB) *ReverseLogisticsLinkRepository {
	return &ReverseLogisticsLinkRepository{Repository: NewRepository[AfterSalesModel.ReturnLogisticsLink](db)}
}

func (r *ReverseLogisticsLinkRepository) Create(ctx context.Context, row *AfterSalesModel.ReturnLogisticsLink) error {
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
