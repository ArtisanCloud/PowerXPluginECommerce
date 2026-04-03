package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

type RedeemInput struct {
	TenantUUID string
	OrderID    string
	RequestID  string
	Operator   string
	Reason     string
}

type RedeemResult struct {
	RedeemedAssetIDs []string `json:"redeemed_asset_ids"`
}

type RedeemService struct {
	assetRepo *couponrepo.AssetRepository
	logSvc    *UsageLogService
}

func NewRedeemService(deps *app.Deps) *RedeemService {
	if deps == nil || deps.DB == nil {
		return &RedeemService{}
	}
	logRepo := couponrepo.NewUsageLogRepository(deps.DB)
	return &RedeemService{
		assetRepo: couponrepo.NewAssetRepository(deps.DB),
		logSvc:    NewUsageLogServiceWithRepo(logRepo),
	}
}

func (s *RedeemService) Ready() bool {
	return s != nil && s.assetRepo != nil && s.logSvc != nil && s.logSvc.Ready()
}

func (s *RedeemService) RedeemWithTx(ctx context.Context, tx *gorm.DB, in RedeemInput) (*RedeemResult, error) {
	if !s.Ready() {
		return nil, errors.New("coupon redeem service unavailable")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	in.TenantUUID = strings.TrimSpace(in.TenantUUID)
	in.OrderID = strings.TrimSpace(in.OrderID)
	if in.TenantUUID == "" || in.OrderID == "" {
		return nil, errors.New("tenant uuid and order id are required")
	}
	if strings.TrimSpace(in.Reason) == "" {
		in.Reason = "payment_success"
	}
	assets, err := s.assetRepo.LockReservedByOrderForUpdate(ctx, tx, in.TenantUUID, in.OrderID)
	if err != nil {
		return nil, err
	}
	result := &RedeemResult{RedeemedAssetIDs: []string{}}
	now := time.Now().UTC()
	for _, row := range assets {
		if strings.EqualFold(strings.TrimSpace(row.Status), AssetStatusRedeemed) {
			result.RedeemedAssetIDs = append(result.RedeemedAssetIDs, row.ID)
			continue
		}
		if !CanTransitAssetStatus(row.Status, AssetStatusRedeemed) {
			continue
		}
		asset := row
		asset.Status = AssetStatusRedeemed
		asset.RedeemedAt = &now
		if err := s.assetRepo.SaveWithTx(ctx, tx, &asset); err != nil {
			return nil, err
		}
		_, err := s.logSvc.WriteActionWithTx(ctx, tx, in.TenantUUID, asset.ID, in.OrderID, "redeem", in.Reason, in.RequestID, in.Operator, BuildActionIdempotencyKey("redeem", in.OrderID, asset.ID))
		if err != nil {
			return nil, err
		}
		result.RedeemedAssetIDs = append(result.RedeemedAssetIDs, asset.ID)
	}
	return result, nil
}
