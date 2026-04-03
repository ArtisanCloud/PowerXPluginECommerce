package jobs

import (
	"context"
	"strings"
	"time"

	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	couponsvc "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

// OrderTimeoutReleaseJob releases reserved coupons for timeout-closed orders.
type OrderTimeoutReleaseJob struct {
	deps       *app.Deps
	releaseSvc *couponsvc.ReleaseService
}

func NewOrderTimeoutReleaseJob(deps *app.Deps) *OrderTimeoutReleaseJob {
	return &OrderTimeoutReleaseJob{deps: deps, releaseSvc: couponsvc.NewReleaseService(deps)}
}

func (j *OrderTimeoutReleaseJob) RunOnce(ctx context.Context, tenantUUID, orderID string) error {
	if j == nil || j.deps == nil || j.deps.DB == nil || j.releaseSvc == nil || !j.releaseSvc.Ready() {
		return nil
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	orderID = strings.TrimSpace(orderID)
	if tenantUUID == "" || orderID == "" {
		return nil
	}
	return j.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.WithContext(ctx).
			Model(&ordermodel.Order{}).
			Where("tenant_uuid = ? AND id = ? AND status = ?", tenantUUID, orderID, "pending_payment").
			Updates(map[string]any{"status": "cancelled", "updated_at": now}).Error; err != nil {
			return err
		}
		_, err := j.releaseSvc.ReleaseWithTx(ctx, tx, couponsvc.ReleaseInput{
			TenantUUID: tenantUUID,
			OrderID:    orderID,
			Reason:     "order_timeout",
			Operator:   "timeout_job",
		})
		return err
	})
}
