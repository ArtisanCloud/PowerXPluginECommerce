package after_sales

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	OrderModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/order"
	AfterSalesRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/after_sales"
	OrderRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/order"
	AdminAfterSales "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrServiceUnavailable = errors.New("after-sales service unavailable")
	ErrCustomerRequired   = errors.New("customer id is required")
	ErrOrderNotFound      = errors.New("order not found")
	ErrCaseNotFound       = errors.New("after-sale case not found")
	ErrOrderItemNotFound  = errors.New("order item not found")
	ErrInvalidCaseType    = errors.New("invalid case type")
	ErrInvalidRequest     = errors.New("invalid after-sales request")
	ErrCaseWindowExpired  = errors.New("after-sales window expired")
)

type DuplicateActiveCaseError struct {
	CaseNo string
}

func (e *DuplicateActiveCaseError) Error() string {
	if e == nil {
		return "active after-sale case already exists"
	}
	if strings.TrimSpace(e.CaseNo) == "" {
		return "active after-sale case already exists"
	}
	return fmt.Sprintf("active after-sale case already exists: %s", e.CaseNo)
}

type CaseService struct {
	deps         *app.Deps
	caseRepo     *AfterSalesRepo.CaseRepository
	timelineRepo *AfterSalesRepo.TimelineRepository
	evidenceRepo *AfterSalesRepo.EvidenceRepository
	orderRepo    *OrderRepo.OrderRepository
}

func NewCaseService(deps *app.Deps) *CaseService {
	if deps == nil || deps.DB == nil {
		return &CaseService{deps: deps}
	}
	return &CaseService{
		deps:         deps,
		caseRepo:     AfterSalesRepo.NewCaseRepository(deps.DB),
		timelineRepo: AfterSalesRepo.NewTimelineRepository(deps.DB),
		evidenceRepo: AfterSalesRepo.NewEvidenceRepository(deps.DB),
		orderRepo:    OrderRepo.NewOrderRepository(deps.DB),
	}
}

func (s *CaseService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.caseRepo != nil && s.timelineRepo != nil && s.evidenceRepo != nil && s.orderRepo != nil
}

func (s *CaseService) Create(ctx context.Context, tenantUUID, customerID string, req CreateCaseRequest) (*CaseDetailDTO, error) {
	if !s.Ready() {
		return nil, ErrServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}

	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	order, err := s.orderRepo.GetByID(ctx, strings.TrimSpace(tenantUUID), strings.TrimSpace(req.OrderID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(order.CustomerID), customerID) {
		return nil, ErrOrderNotFound
	}

	if err := s.verifyOrderItemBelongs(ctx, tenantUUID, req.OrderID, req.OrderItemID); err != nil {
		return nil, err
	}
	if err := validateCaseWindow(req.CaseType, order); err != nil {
		return nil, err
	}

	active, err := s.caseRepo.FindActiveByOrderItem(ctx, req.OrderItemID)
	if err == nil && active != nil {
		return nil, &DuplicateActiveCaseError{CaseNo: active.CaseNo}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	caseRow := &AfterSalesModel.AfterSaleCase{
		ID:                   utils.NewUUID(),
		TenantUUID:           strings.TrimSpace(tenantUUID),
		CaseNo:               generateCaseNo(now),
		OrderID:              strings.TrimSpace(req.OrderID),
		OrderItemID:          strings.TrimSpace(req.OrderItemID),
		CustomerID:           customerID,
		CaseType:             strings.TrimSpace(strings.ToLower(req.CaseType)),
		Status:               AdminAfterSales.CaseStatusPending,
		ReasonCode:           strings.TrimSpace(req.ReasonCode),
		ReasonDetail:         strings.TrimSpace(req.ReasonDetail),
		RequestedQty:         maxInt(req.RequestedQty, 1),
		RequestedAmountMinor: maxInt64(req.RequestedAmountMinor, 0),
		Currency:             defaultString(req.Currency, "CNY"),
		SourceChannel:        "miniapp",
	}

	initialTimeline := &AfterSalesModel.AfterSaleTimeline{
		ID:           utils.NewUUID(),
		TenantUUID:   strings.TrimSpace(tenantUUID),
		CaseID:       caseRow.ID,
		Action:       "create",
		FromStatus:   "",
		ToStatus:     AdminAfterSales.CaseStatusPending,
		OperatorType: "customer",
		OperatorID:   customerID,
		Note:         strings.TrimSpace(req.ReasonDetail),
	}

	if err := s.caseRepo.WithTenantTx(ctx, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(caseRow).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Create(initialTimeline).Error; err != nil {
			return err
		}
		if len(req.Evidences) == 0 {
			return nil
		}
		evidences := make([]AfterSalesModel.AfterSaleEvidence, 0, len(req.Evidences))
		for _, item := range req.Evidences {
			if strings.TrimSpace(item.ContentRef) == "" {
				continue
			}
			evidences = append(evidences, AfterSalesModel.AfterSaleEvidence{
				ID:           utils.NewUUID(),
				TenantUUID:   strings.TrimSpace(tenantUUID),
				CaseID:       caseRow.ID,
				UploaderType: "customer",
				UploaderID:   customerID,
				EvidenceType: defaultString(item.EvidenceType, "other"),
				ContentRef:   strings.TrimSpace(item.ContentRef),
				Description:  strings.TrimSpace(item.Description),
			})
		}
		if len(evidences) == 0 {
			return nil
		}
		return tx.WithContext(ctx).Create(&evidences).Error
	}); err != nil {
		return nil, err
	}

	return &CaseDetailDTO{
		Case:     toSummaryDTO(*caseRow),
		Timeline: toTimelineDTO([]AfterSalesModel.AfterSaleTimeline{*initialTimeline}),
	}, nil
}

func (s *CaseService) verifyOrderItemBelongs(ctx context.Context, tenantUUID, orderID, orderItemID string) error {
	var item OrderModel.OrderItem
	err := s.deps.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND order_id = ? AND id = ?", strings.TrimSpace(tenantUUID), strings.TrimSpace(orderID), strings.TrimSpace(orderItemID)).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderItemNotFound
		}
		return err
	}
	return nil
}

func validateCreateRequest(req CreateCaseRequest) error {
	if strings.TrimSpace(req.OrderID) == "" || strings.TrimSpace(req.OrderItemID) == "" || strings.TrimSpace(req.ReasonCode) == "" {
		return ErrInvalidRequest
	}
	if AdminAfterSales.NormalizeCaseStatus(AdminAfterSales.CaseStatusPending) == "" {
		return ErrInvalidRequest
	}
	caseType := strings.TrimSpace(strings.ToLower(req.CaseType))
	if caseType != "refund_only" && caseType != "return_refund" && caseType != "exchange" {
		return ErrInvalidCaseType
	}
	return nil
}

func validateCaseWindow(caseType string, order *OrderModel.Order) error {
	if order == nil {
		return ErrOrderNotFound
	}
	now := time.Now().UTC()
	caseType = strings.TrimSpace(strings.ToLower(caseType))
	since := now.Sub(order.CreatedAt.UTC())

	switch caseType {
	case "refund_only":
		if since > 24*time.Hour {
			return ErrCaseWindowExpired
		}
		status := strings.TrimSpace(strings.ToLower(order.Status))
		if strings.Contains(status, "shipped") || strings.Contains(status, "delivered") || strings.Contains(status, "completed") {
			return ErrCaseWindowExpired
		}
	case "return_refund":
		if since > 7*24*time.Hour {
			return ErrCaseWindowExpired
		}
	case "exchange":
		if since > 15*24*time.Hour {
			return ErrCaseWindowExpired
		}
	default:
		return ErrInvalidCaseType
	}
	return nil
}

func generateCaseNo(now time.Time) string {
	return fmt.Sprintf("RMA%s%03d", now.UTC().Format("20060102150405"), now.Nanosecond()%1000)
}

func defaultString(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func maxInt(v int, fallback int) int {
	if v <= 0 {
		return fallback
	}
	return v
}

func maxInt64(v int64, fallback int64) int64 {
	if v < 0 {
		return fallback
	}
	return v
}
