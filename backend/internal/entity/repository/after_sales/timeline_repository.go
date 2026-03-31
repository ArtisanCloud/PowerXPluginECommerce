package after_sales

import (
	"context"
	"strings"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"gorm.io/gorm"
)

// TimelineRepository persists after-sale timeline entries.
type TimelineRepository struct {
	*Repository[AfterSalesModel.AfterSaleTimeline]
}

func NewTimelineRepository(db *gorm.DB) *TimelineRepository {
	return &TimelineRepository{Repository: NewRepository[AfterSalesModel.AfterSaleTimeline](db)}
}

func (r *TimelineRepository) Create(ctx context.Context, row *AfterSalesModel.AfterSaleTimeline) error {
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
