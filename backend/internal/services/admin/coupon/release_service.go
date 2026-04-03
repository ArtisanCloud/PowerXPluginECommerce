package coupon

import (
	"context"
	"errors"
	"strings"

	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

type ReleaseInput struct {
	TenantUUID string
	OrderID    string
	RequestID  string
	Operator   string
	Reason     string
}

type ReleaseResult struct {
	ReleasedAssetIDs []string `json:"released_asset_ids"`
}

type ReleaseService struct {
	assetRepo *couponrepo.AssetRepository
	logSvc    *UsageLogService
}

func NewReleaseService(deps *app.Deps) *ReleaseService {
	if deps == nil || deps.DB == nil {
		return &ReleaseService{}
	}
	logRepo := couponrepo.NewUsageLogRepository(deps.DB)
	return &ReleaseService{
		assetRepo: couponrepo.NewAssetRepository(deps.DB),
		logSvc:    NewUsageLogServiceWithRepo(logRepo),
	}
}

func (s *ReleaseService) Ready() bool {
	return s != nil && s.assetRepo != nil && s.logSvc != nil && s.logSvc.Ready()
}

func (s *ReleaseService) ReleaseWithTx(ctx context.Context, tx *gorm.DB, in ReleaseInput) (*ReleaseResult, error) {
	if !s.Ready() {
		return nil, errors.New("coupon release service unavailable")
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
		in.Reason = "order_closed"
	}
	assets, err := s.assetRepo.LockReservedByOrderForUpdate(ctx, tx, in.TenantUUID, in.OrderID)
	if err != nil {
		return nil, err
	}
	result := &ReleaseResult{ReleasedAssetIDs: []string{}}
	for _, row := range assets {
		if !CanTransitAssetStatus(row.Status, AssetStatusAvailable) {
			continue
		}
		asset := row
		asset.Status = AssetStatusAvailable
		asset.ReservedOrderID = nil
		asset.ReservedAt = nil
		if err := s.assetRepo.SaveWithTx(ctx, tx, &asset); err != nil {
			return nil, err
		}
		_, err := s.logSvc.WriteActionWithTx(ctx, tx, in.TenantUUID, asset.ID, in.OrderID, "release", in.Reason, in.RequestID, in.Operator, BuildActionIdempotencyKey("release", in.OrderID, asset.ID))
		if err != nil {
			return nil, err
		}
		result.ReleasedAssetIDs = append(result.ReleasedAssetIDs, asset.ID)
	}
	return result, nil
}
