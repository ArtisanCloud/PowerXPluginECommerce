package after_sales

import (
	"context"
	"strings"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	"gorm.io/gorm"
)

// DecisionRepository persists operator decision logs.
type DecisionRepository struct {
	*Repository[AfterSalesModel.AfterSaleDecision]
}

func NewDecisionRepository(db *gorm.DB) *DecisionRepository {
	return &DecisionRepository{Repository: NewRepository[AfterSalesModel.AfterSaleDecision](db)}
}

func (r *DecisionRepository) Create(ctx context.Context, row *AfterSalesModel.AfterSaleDecision) error {
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
