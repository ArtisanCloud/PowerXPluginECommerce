package coupon

import (
	"context"
	"errors"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationInput struct {
	TenantUUID string
	OrderID    string
	AssetIDs   []string
	RequestID  string
	CreatedBy  string
}

type ReservationResult struct {
	ReservedAssetIDs []string `json:"reserved_asset_ids"`
}

type ReservationService struct {
	assetRepo *couponrepo.AssetRepository
	logRepo   *couponrepo.UsageLogRepository
}

func NewReservationService(deps *app.Deps) *ReservationService {
	if deps == nil || deps.DB == nil {
		return &ReservationService{}
	}
	return &ReservationService{
		assetRepo: couponrepo.NewAssetRepository(deps.DB),
		logRepo:   couponrepo.NewUsageLogRepository(deps.DB),
	}
}

func (s *ReservationService) Ready() bool {
	return s != nil && s.assetRepo != nil && s.logRepo != nil
}

func (s *ReservationService) ReserveWithTx(ctx context.Context, tx *gorm.DB, in ReservationInput) (*ReservationResult, error) {
	if !s.Ready() {
		return nil, errors.New("coupon reservation service unavailable")
	}
	if tx == nil {
		return nil, errors.New("transaction is required")
	}
	in.TenantUUID = strings.TrimSpace(in.TenantUUID)
	in.OrderID = strings.TrimSpace(in.OrderID)
	if in.TenantUUID == "" || in.OrderID == "" {
		return nil, errors.New("tenant uuid and order id are required")
	}
	trimmed := make([]string, 0, len(in.AssetIDs))
	for _, id := range in.AssetIDs {
		if v := strings.TrimSpace(id); v != "" {
			trimmed = append(trimmed, v)
		}
	}
	if len(trimmed) == 0 {
		return &ReservationResult{ReservedAssetIDs: []string{}}, nil
	}

	assets, err := s.assetRepo.LockByIDsForUpdate(ctx, tx, in.TenantUUID, trimmed)
	if err != nil {
		return nil, err
	}
	assetMap := make(map[string]*couponmodel.CouponAsset, len(assets))
	for i := range assets {
		row := assets[i]
		assetMap[row.ID] = &row
	}
	reservedIDs := make([]string, 0, len(trimmed))
	now := time.Now().UTC()
	for _, id := range trimmed {
		asset, ok := assetMap[id]
		if !ok {
			return nil, NewRuleError(ReasonNotFound, "coupon asset not found")
		}
		if strings.EqualFold(asset.Status, AssetStatusReserved) && asset.ReservedOrderID != nil && strings.TrimSpace(*asset.ReservedOrderID) == in.OrderID {
			reservedIDs = append(reservedIDs, id)
			continue
		}
		if !CanTransitAssetStatus(asset.Status, AssetStatusReserved) {
			return nil, NewRuleError(ReasonOccupied, "coupon asset is not available")
		}
		orderID := in.OrderID
		asset.Status = AssetStatusReserved
		asset.ReservedOrderID = &orderID
		asset.ReservedAt = &now
		if err := s.assetRepo.SaveWithTx(ctx, tx, asset); err != nil {
			return nil, err
		}
		_, err := s.logRepo.CreateWithTxIdempotent(ctx, tx, &couponmodel.CouponUsageLog{
			ID:             uuid.NewString(),
			TenantUUID:     in.TenantUUID,
			AssetID:        asset.ID,
			OrderID:        &orderID,
			Action:         "reserve",
			ActionReason:   "order_submit",
			IdempotencyKey: buildReserveIdempotencyKey(orderID, asset.ID),
			RequestID:      strings.TrimSpace(in.RequestID),
			CreatedBy:      strings.TrimSpace(in.CreatedBy),
		})
		if err != nil {
			return nil, err
		}
		reservedIDs = append(reservedIDs, id)
	}
	return &ReservationResult{ReservedAssetIDs: reservedIDs}, nil
}
