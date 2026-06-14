package promotion

import (
	"context"
	"encoding/json"
	"strings"

	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	promotionrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SnapshotService struct {
	deps *app.Deps
	repo *promotionrepo.OrderSnapshotRepository
}

func NewSnapshotService(deps *app.Deps) *SnapshotService {
	if deps == nil || deps.DB == nil {
		return &SnapshotService{deps: deps}
	}
	return &SnapshotService{deps: deps, repo: promotionrepo.NewOrderSnapshotRepository(deps.DB)}
}

func (s *SnapshotService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *SnapshotService) Save(ctx context.Context, db *gorm.DB, tenantUUID, orderID string, result *QuoteResult) (*promotionmodel.OrderSnapshot, error) {
	if !s.Ready() || result == nil {
		return nil, ErrPromotionServiceUnavailable
	}
	if db == nil {
		db = s.deps.DB
	}
	applied, _ := json.Marshal(result.AppliedPromotions)
	rejected, _ := json.Marshal(result.RejectedPromotions)
	allocations, _ := json.Marshal(result.LineAllocations)
	row := &promotionmodel.OrderSnapshot{
		TenantUUID: strings.TrimSpace(tenantUUID), OrderID: strings.TrimSpace(orderID), Currency: result.Currency,
		BaseTotalMinor: result.BaseTotalMinor, PromotionDiscountMinor: result.PromotionDiscountMinor,
		AfterPromotionTotalMinor: result.AfterPromotionTotalMinor, AppliedPromotions: datatypes.JSON(applied),
		RejectedPromotions: datatypes.JSON(rejected), LineAllocations: datatypes.JSON(allocations), PricedAt: result.PricedAt,
	}
	return row, db.WithContext(ctx).Create(row).Error
}

func (s *SnapshotService) FindByOrderID(ctx context.Context, tenantUUID, orderID string) (*promotionmodel.OrderSnapshot, error) {
	if !s.Ready() {
		return nil, ErrPromotionServiceUnavailable
	}
	return s.repo.FindByOrderID(ctx, tenantUUID, orderID)
}
