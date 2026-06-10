package coupon

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	couponmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/coupon"
	couponrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/coupon"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

type RefundInput struct {
	TenantUUID string
	OrderID    string
	RefundNo   string
	RequestID  string
	Operator   string
	Reason     string
}

type RefundResult struct {
	RefundedAssetIDs []string `json:"refunded_asset_ids"`
	SkippedAssetIDs  []string `json:"skipped_asset_ids"`
}

type RefundService struct {
	assetRepo    *couponrepo.AssetRepository
	templateRepo *couponrepo.TemplateRepository
	logSvc       *UsageLogService
}

func NewRefundService(deps *app.Deps) *RefundService {
	if deps == nil || deps.DB == nil {
		return &RefundService{}
	}
	return &RefundService{
		assetRepo:    couponrepo.NewAssetRepository(deps.DB),
		templateRepo: couponrepo.NewTemplateRepository(deps.DB),
		logSvc:       NewUsageLogServiceWithRepo(couponrepo.NewUsageLogRepository(deps.DB)),
	}
}

func (s *RefundService) Ready() bool {
	return s != nil && s.assetRepo != nil && s.templateRepo != nil && s.logSvc != nil && s.logSvc.Ready()
}

func (s *RefundService) RefundWithTx(ctx context.Context, tx *gorm.DB, in RefundInput) (*RefundResult, error) {
	if !s.Ready() {
		return nil, errors.New("coupon refund service unavailable")
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
		in.Reason = "payment_refund"
	}
	assets, err := s.assetRepo.LockRedeemedByOrderForUpdate(ctx, tx, in.TenantUUID, in.OrderID)
	if err != nil {
		return nil, err
	}
	result := &RefundResult{RefundedAssetIDs: []string{}, SkippedAssetIDs: []string{}}
	if len(assets) == 0 {
		return result, nil
	}
	templateIDs := make([]string, 0, len(assets))
	for _, row := range assets {
		templateIDs = append(templateIDs, row.TemplateID)
	}
	templates, err := s.templateRepo.MapByIDAnyWithTx(ctx, tx, in.TenantUUID, templateIDs)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for _, row := range assets {
		tpl, ok := templates[strings.TrimSpace(row.TemplateID)]
		if !ok || !refundRuleReturnsCoupon(tpl) {
			result.SkippedAssetIDs = append(result.SkippedAssetIDs, row.ID)
			continue
		}
		if !CanTransitAssetStatus(row.Status, AssetStatusRefunded) {
			result.SkippedAssetIDs = append(result.SkippedAssetIDs, row.ID)
			continue
		}
		asset := row
		asset.Status = AssetStatusRefunded
		asset.RefundedAt = &now
		if err := s.assetRepo.SaveWithTx(ctx, tx, &asset); err != nil {
			return nil, err
		}
		idempotencyKey := BuildActionIdempotencyKey("refund", in.OrderID, asset.ID)
		if v := strings.TrimSpace(in.RefundNo); v != "" {
			idempotencyKey = idempotencyKey + ":" + v
		}
		_, err := s.logSvc.WriteActionWithTx(ctx, tx, in.TenantUUID, asset.ID, in.OrderID, "refund", in.Reason, in.RequestID, in.Operator, idempotencyKey)
		if err != nil {
			return nil, err
		}
		result.RefundedAssetIDs = append(result.RefundedAssetIDs, asset.ID)
	}
	return result, nil
}

func refundRuleReturnsCoupon(tpl couponmodel.CouponTemplate) bool {
	var rule map[string]any
	if len(tpl.RefundRule) == 0 {
		return false
	}
	if err := json.Unmarshal(tpl.RefundRule, &rule); err != nil {
		return false
	}
	if v, ok := rule["return_coupon"].(bool); ok && v {
		return true
	}
	policy := strings.TrimSpace(strings.ToLower(stringFromAny(rule["policy"])))
	switch policy {
	case "return_coupon", "full_refund", "proportional_refund":
		return true
	default:
		return false
	}
}

func stringFromAny(v any) string {
	switch value := v.(type) {
	case string:
		return value
	default:
		return ""
	}
}
