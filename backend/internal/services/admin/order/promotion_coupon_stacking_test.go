package order

import (
	"context"
	"testing"
	"time"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	promotionmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/promotion"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/stretchr/testify/require"
)

func TestServiceCreateOrder_PromotionExcludesCoupon(t *testing.T) {
	ctx := context.Background()
	models.ForceSchemaForTests("")
	db := newTestDB(t)
	createOrderTables(t, db)
	tenant, adminID, customerID, spuID, skuID := "tenant-test", "1001", "cust-1", "spu-1", "sku-1"
	seedCustomer(t, db, tenant, customerID)
	seedSellabilityFixtures(t, db, tenant, spuID, skuID, 10, 0, true)
	now := time.Now().UTC()
	seedPromotion(t, db, promotionmodel.Campaign{
		ID: "promo-no-coupon", TenantUUID: tenant, Code: "NOCOUPON", Name: "排斥券", PromotionType: promotionmodel.TypeAmountOff,
		ConditionRule: mustJSON(map[string]any{"min_order_amount_minor": 1000}),
		ScopeRule:     mustJSON(map[string]any{"scope_type": "all"}),
		ActionRule:    mustJSON(map[string]any{"discount_amount_minor": 100}),
		StackingRule:  mustJSON(map[string]any{"priority": 100, "stackable": true, "stackable_with_coupon": false}),
		ValidFrom:     now.Add(-time.Hour), ValidTo: now.Add(time.Hour), Status: promotionmodel.StatusActive,
	})
	svc := NewService(&app.Deps{DB: db})
	out, err := svc.CreateOrder(ctx, tenant, adminID, "promo-idem-2", CreateOrderRequest{
		CustomerID: customerID, Channel: "official", CouponIDs: []string{"coupon-1"},
		ShippingAddress: &ShippingAddress{RecipientName: "张三", RecipientPhone: "13800138000", Address1: "北京"},
		Items:           []CreateOrderItemInput{{SKUID: skuID, Qty: 1}},
	})
	require.NoError(t, err)
	require.NotNil(t, out.Coupon)
	require.Len(t, out.Coupon.RejectedCoupons, 1)
	require.Equal(t, promotion.ReasonPromotionExcludesCoupon, out.Coupon.RejectedCoupons[0].Reason)
}
