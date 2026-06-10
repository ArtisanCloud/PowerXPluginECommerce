package coupon

import (
	"context"
	"errors"
	"fmt"
	"strings"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UsageLogService struct {
	repo *couponrepo.UsageLogRepository
}

func NewUsageLogService(deps *app.Deps) *UsageLogService {
	if deps == nil || deps.DB == nil {
		return &UsageLogService{}
	}
	return &UsageLogService{repo: couponrepo.NewUsageLogRepository(deps.DB)}
}

func NewUsageLogServiceWithRepo(repo *couponrepo.UsageLogRepository) *UsageLogService {
	return &UsageLogService{repo: repo}
}

func (s *UsageLogService) Ready() bool {
	return s != nil && s.repo != nil
}

func (s *UsageLogService) WriteActionWithTx(ctx context.Context, tx *gorm.DB, tenantUUID, assetID, orderID, action, reason, requestID, actor, idempotencyKey string) (bool, error) {
	if !s.Ready() {
		return false, errors.New("coupon usage log service unavailable")
	}
	if tx == nil {
		return false, errors.New("transaction is required")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	assetID = strings.TrimSpace(assetID)
	action = strings.TrimSpace(strings.ToLower(action))
	reason = strings.TrimSpace(reason)
	if tenantUUID == "" || assetID == "" || action == "" {
		return false, errors.New("tenant uuid, asset id and action are required")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey = BuildActionIdempotencyKey(action, orderID, assetID)
	}
	var orderPtr *string
	if v := strings.TrimSpace(orderID); v != "" {
		orderPtr = &v
	}
	return s.repo.CreateWithTxIdempotent(ctx, tx, &couponmodel.CouponUsageLog{
		ID:             uuid.NewString(),
		TenantUUID:     tenantUUID,
		AssetID:        assetID,
		OrderID:        orderPtr,
		Action:         action,
		ActionReason:   reason,
		IdempotencyKey: idempotencyKey,
		RequestID:      strings.TrimSpace(requestID),
		CreatedBy:      strings.TrimSpace(actor),
	})
}

func BuildActionIdempotencyKey(action, orderID, assetID string) string {
	return fmt.Sprintf("%s:%s:%s", strings.TrimSpace(strings.ToLower(action)), strings.TrimSpace(orderID), strings.TrimSpace(assetID))
}
