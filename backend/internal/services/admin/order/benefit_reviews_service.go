package order

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	ordermodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	orderrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrBenefitReviewServiceUnavailable = errors.New("benefit review service unavailable")
	ErrBenefitReviewNotFound           = errors.New("benefit review not found")
	ErrBenefitReviewInvalidStatus      = errors.New("benefit review status invalid")
	ErrBenefitReviewSameOperator       = errors.New("reviewer must be different from submitter")
	ErrBenefitReviewReasonRequired     = errors.New("review reason is required")
	ErrBenefitReviewOrderNotEditable   = errors.New("order is not editable")
	ErrBenefitReviewCodeRequired       = errors.New("benefit code is required")
	ErrBenefitReviewTypeInvalid        = errors.New("benefit type is invalid")
	ErrBenefitReviewValueInvalid       = errors.New("benefit value is invalid")
	ErrBenefitReviewUsed               = errors.New("benefit code already used")
	ErrBenefitReviewTotalExceeded      = errors.New("benefit amount exceeds order total")
	ErrBenefitReviewStackingNotAllowed = errors.New("benefit stacking not allowed")
)

type BenefitReviewService struct {
	deps       *app.Deps
	reviewRepo *repo.OrderBenefitReviewRepository
	orderRepo  *orderrepo.OrderRepository
	eventRepo  *orderrepo.OrderEventRepository
}

func NewBenefitReviewService(deps *app.Deps) *BenefitReviewService {
	if deps == nil || deps.DB == nil {
		return &BenefitReviewService{deps: deps}
	}
	return &BenefitReviewService{
		deps:       deps,
		reviewRepo: repo.NewOrderBenefitReviewRepository(deps.DB),
		orderRepo:  orderrepo.NewOrderRepository(deps.DB),
		eventRepo:  orderrepo.NewOrderEventRepository(deps.DB),
	}
}

func (s *BenefitReviewService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.reviewRepo != nil && s.orderRepo != nil && s.eventRepo != nil
}

func (s *BenefitReviewService) ListReviews(ctx context.Context, tenantUUID, adminID, orderID string) ([]OrderBenefitReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrBenefitReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	rows, err := s.reviewRepo.ListByOrderID(ctx, tenantUUID, orderID)
	if err != nil {
		return nil, err
	}
	out := make([]OrderBenefitReviewDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toBenefitReviewDTO(&row))
	}
	return out, nil
}

func (s *BenefitReviewService) CreateReview(ctx context.Context, tenantUUID, adminID, orderID string, req BenefitReviewCreateRequest) (*OrderBenefitReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrBenefitReviewServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	adminID = strings.TrimSpace(adminID)
	orderID = strings.TrimSpace(orderID)
	req.BenefitType = strings.TrimSpace(req.BenefitType)
	req.BenefitCode = strings.TrimSpace(req.BenefitCode)
	req.ValueType = strings.TrimSpace(req.ValueType)
	req.Currency = strings.TrimSpace(req.Currency)
	req.Note = strings.TrimSpace(req.Note)

	if adminID == "" {
		return nil, errors.New("admin id is required")
	}
	if orderID == "" {
		return nil, errors.New("order id is required")
	}
	if req.BenefitCode == "" {
		return nil, ErrBenefitReviewCodeRequired
	}
	if !isValidBenefitType(req.BenefitType) {
		return nil, ErrBenefitReviewTypeInvalid
	}
	if !isValidValueType(req.BenefitType, req.ValueType) {
		return nil, ErrBenefitReviewValueInvalid
	}
	if req.Value <= 0 {
		return nil, ErrBenefitReviewValueInvalid
	}

	var created *models.OrderBenefitReview
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		orderRow, err := s.orderRepo.LockByID(ctx, tx, tenantUUID, orderID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if !isOrderEditableStatus(orderRow.Status) {
			return ErrBenefitReviewOrderNotEditable
		}
		usedCount, err := s.reviewRepo.CountApprovedByCode(ctx, tx, tenantUUID, req.BenefitType, req.BenefitCode, 0)
		if err != nil {
			return err
		}
		if usedCount > 0 {
			return ErrBenefitReviewUsed
		}
		if !req.StackingAllowed {
			activeCount, err := s.reviewRepo.CountActiveByOrder(ctx, tx, tenantUUID, orderID)
			if err != nil {
				return err
			}
			if activeCount > 0 {
				return ErrBenefitReviewStackingNotAllowed
			}
		}

		value, amountMinor, err := computeBenefitAmounts(orderRow.TotalAmount, req.ValueType, req.Value)
		if err != nil {
			return err
		}
		if amountMinor > orderRow.TotalAmount {
			return ErrBenefitReviewTotalExceeded
		}
		currency := req.Currency
		if currency == "" {
			currency = orderRow.Currency
		}

		now := time.Now().UTC()
		row := &models.OrderBenefitReview{
			BaseModel:       models.BaseModel{TenantUuid: tenantUUID},
			OrderID:         orderID,
			OrderNo:         orderRow.OrderNo,
			BenefitType:     req.BenefitType,
			BenefitCode:     req.BenefitCode,
			ValueType:       req.ValueType,
			Value:           value,
			AmountMinor:     amountMinor,
			Currency:        currency,
			StackingAllowed: req.StackingAllowed,
			Status:          "pending_review",
			SubmittedBy:     adminID,
			SubmittedAt:     now,
			Note:            req.Note,
		}
		if err := tx.WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
		eventPayload, _ := jsonBenefitReviewPayload(ctx, row, adminID, "created", "")
		event := &ordermodel.OrderEvent{
			ID:           uuid.NewString(),
			TenantUUID:   tenantUUID,
			OrderID:      orderID,
			EventType:    "order.benefit.review.created",
			OperatorType: "admin",
			Operator:     adminID,
			Payload:      datatypes.JSON(eventPayload),
		}
		if err := s.eventRepo.CreateWithTx(ctx, tx, event); err != nil {
			return err
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	resp := toBenefitReviewDTO(created)
	return &resp, nil
}

func (s *BenefitReviewService) ApproveReviews(ctx context.Context, tenantUUID, adminID string, reviewIDs []uint64, req BenefitReviewDecisionRequest) ([]OrderBenefitReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrBenefitReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if len(reviewIDs) == 0 {
		return nil, errors.New("review ids are required")
	}
	updated := make([]OrderBenefitReviewDTO, 0, len(reviewIDs))
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		for _, reviewID := range reviewIDs {
			row, err := s.reviewRepo.LockByID(ctx, tx, tenantUUID, reviewID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrBenefitReviewNotFound
				}
				return err
			}
			if strings.TrimSpace(row.Status) != "pending_review" {
				return ErrBenefitReviewInvalidStatus
			}
			if row.SubmittedBy == adminID && !isRootOperator(ctx) {
				return ErrBenefitReviewSameOperator
			}
			orderRow, err := s.orderRepo.LockByID(ctx, tx, tenantUUID, row.OrderID)
			if err != nil {
				return err
			}
			if !isOrderEditableStatus(orderRow.Status) {
				return ErrBenefitReviewOrderNotEditable
			}
			usedCount, err := s.reviewRepo.CountApprovedByCode(ctx, tx, tenantUUID, row.BenefitType, row.BenefitCode, row.ID)
			if err != nil {
				return err
			}
			if usedCount > 0 {
				return ErrBenefitReviewUsed
			}
			if !row.StackingAllowed {
				activeCount, err := s.reviewRepo.CountActiveByOrder(ctx, tx, tenantUUID, row.OrderID)
				if err != nil {
					return err
				}
				if activeCount > 1 {
					return ErrBenefitReviewStackingNotAllowed
				}
			}

			now := time.Now().UTC()
			updates := map[string]any{
				"status":        "approved",
				"reviewed_by":   adminID,
				"reviewed_at":   now,
				"review_reason": strings.TrimSpace(req.Reason),
			}
			if err := tx.WithContext(ctx).
				Model(&models.OrderBenefitReview{}).
				Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).
				Updates(updates).Error; err != nil {
				return err
			}
			eventPayload, _ := jsonBenefitReviewPayload(ctx, row, adminID, "approved", req.Reason)
			event := &ordermodel.OrderEvent{
				ID:           uuid.NewString(),
				TenantUUID:   tenantUUID,
				OrderID:      row.OrderID,
				EventType:    "order.benefit.review.approved",
				OperatorType: "admin",
				Operator:     adminID,
				Payload:      datatypes.JSON(eventPayload),
			}
			if err := s.eventRepo.CreateWithTx(ctx, tx, event); err != nil {
				return err
			}
			if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).First(&row).Error; err != nil {
				return err
			}
			updated = append(updated, toBenefitReviewDTO(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *BenefitReviewService) RejectReviews(ctx context.Context, tenantUUID, adminID string, reviewIDs []uint64, req BenefitReviewDecisionRequest) ([]OrderBenefitReviewDTO, error) {
	if !s.Ready() {
		return nil, ErrBenefitReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if len(reviewIDs) == 0 {
		return nil, errors.New("review ids are required")
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, ErrBenefitReviewReasonRequired
	}
	updated := make([]OrderBenefitReviewDTO, 0, len(reviewIDs))
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		for _, reviewID := range reviewIDs {
			row, err := s.reviewRepo.LockByID(ctx, tx, tenantUUID, reviewID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrBenefitReviewNotFound
				}
				return err
			}
			if strings.TrimSpace(row.Status) != "pending_review" {
				return ErrBenefitReviewInvalidStatus
			}
			if row.SubmittedBy == adminID && !isRootOperator(ctx) {
				return ErrBenefitReviewSameOperator
			}
			now := time.Now().UTC()
			updates := map[string]any{
				"status":        "rejected",
				"reviewed_by":   adminID,
				"reviewed_at":   now,
				"review_reason": reason,
			}
			if err := tx.WithContext(ctx).
				Model(&models.OrderBenefitReview{}).
				Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).
				Updates(updates).Error; err != nil {
				return err
			}
			eventPayload, _ := jsonBenefitReviewPayload(ctx, row, adminID, "rejected", reason)
			event := &ordermodel.OrderEvent{
				ID:           uuid.NewString(),
				TenantUUID:   tenantUUID,
				OrderID:      row.OrderID,
				EventType:    "order.benefit.review.rejected",
				OperatorType: "admin",
				Operator:     adminID,
				Payload:      datatypes.JSON(eventPayload),
			}
			if err := s.eventRepo.CreateWithTx(ctx, tx, event); err != nil {
				return err
			}
			if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, row.ID).First(&row).Error; err != nil {
				return err
			}
			updated = append(updated, toBenefitReviewDTO(row))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *BenefitReviewService) SearchBenefitCodes(ctx context.Context, tenantUUID, adminID, benefitType, keyword string) ([]map[string]string, error) {
	if !s.Ready() {
		return nil, ErrBenefitReviewServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	benefitType = strings.TrimSpace(benefitType)
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []map[string]string{}, nil
	}
	if !isValidBenefitType(benefitType) {
		return nil, ErrBenefitReviewTypeInvalid
	}
	var available bool
	err := s.reviewRepo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		usedCount, err := s.reviewRepo.CountApprovedByCode(ctx, tx, tenantUUID, benefitType, keyword, 0)
		if err != nil {
			return err
		}
		available = usedCount == 0
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !available {
		return []map[string]string{}, nil
	}
	return []map[string]string{{"label": keyword, "value": keyword}}, nil
}

func toBenefitReviewDTO(row *models.OrderBenefitReview) OrderBenefitReviewDTO {
	if row == nil {
		return OrderBenefitReviewDTO{}
	}
	return OrderBenefitReviewDTO{
		ID:              row.ID,
		OrderID:         row.OrderID,
		OrderNo:         row.OrderNo,
		BenefitType:     row.BenefitType,
		BenefitCode:     row.BenefitCode,
		ValueType:       row.ValueType,
		Value:           row.Value,
		AmountMinor:     row.AmountMinor,
		Currency:        row.Currency,
		StackingAllowed: row.StackingAllowed,
		Status:          row.Status,
		SubmittedBy:     row.SubmittedBy,
		SubmittedAt:     row.SubmittedAt,
		ReviewedBy:      row.ReviewedBy,
		ReviewedAt:      row.ReviewedAt,
		ReviewReason:    row.ReviewReason,
		Note:            row.Note,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func computeBenefitAmounts(orderTotal int64, valueType string, value float64) (int64, int64, error) {
	switch valueType {
	case "amount", "balance":
		minor := int64(math.Round(value * 100))
		if minor <= 0 {
			return 0, 0, ErrBenefitReviewValueInvalid
		}
		return minor, minor, nil
	case "percent":
		if value <= 0 || value > 100 {
			return 0, 0, ErrBenefitReviewValueInvalid
		}
		amount := int64(math.Round(float64(orderTotal) * value / 100))
		if amount <= 0 {
			return 0, 0, ErrBenefitReviewValueInvalid
		}
		bps := int64(math.Round(value * 100))
		return bps, amount, nil
	default:
		return 0, 0, ErrBenefitReviewValueInvalid
	}
}

func isValidBenefitType(raw string) bool {
	switch strings.TrimSpace(raw) {
	case "coupon", "giftcard":
		return true
	default:
		return false
	}
}

func isValidValueType(benefitType, valueType string) bool {
	valueType = strings.TrimSpace(valueType)
	switch strings.TrimSpace(benefitType) {
	case "coupon":
		return valueType == "amount" || valueType == "percent"
	case "giftcard":
		return valueType == "balance"
	default:
		return false
	}
}

func jsonBenefitReviewPayload(ctx context.Context, review *models.OrderBenefitReview, reviewer, action, reason string) ([]byte, error) {
	requestID, _ := authx.RequestIDFromContext(ctx)
	payload := map[string]any{
		"requestId":       requestID,
		"reviewId":        review.ID,
		"orderId":         review.OrderID,
		"orderNo":         review.OrderNo,
		"benefitType":     review.BenefitType,
		"benefitCode":     review.BenefitCode,
		"valueType":       review.ValueType,
		"value":           review.Value,
		"amount":          review.AmountMinor,
		"currency":        review.Currency,
		"stackingAllowed": review.StackingAllowed,
		"note":            review.Note,
		"reviewer":        reviewer,
		"action":          action,
		"reason":          reason,
		"submittedBy":     review.SubmittedBy,
	}
	return json.Marshal(payload)
}

func isRootOperator(ctx context.Context) bool {
	tc, ok := authx.TenantContextFromContext(ctx)
	if !ok || len(tc.Roles) == 0 {
		return false
	}
	for _, role := range tc.Roles {
		switch strings.ToLower(strings.TrimSpace(role)) {
		case "superadmin", "system.admin", "root":
			return true
		}
	}
	return false
}
