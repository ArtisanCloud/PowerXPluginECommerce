package after_sales

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	AfterSalesModel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/after_sales"
	AfterSalesRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrAdminServiceUnavailable = errors.New("after-sales admin service unavailable")
	ErrAdminCaseNotFound       = errors.New("after-sales case not found")
	ErrInvalidAction           = errors.New("invalid after-sales action")
	ErrCoreFieldsFrozen        = errors.New("core fields are read-only after case reaches terminal status")
)

var terminalStatuses = map[string]struct{}{
	CaseStatusRejected:  {},
	CaseStatusCompleted: {},
	CaseStatusClosed:    {},
}

type CaseService struct {
	deps         *app.Deps
	caseRepo     *AfterSalesRepo.CaseRepository
	timelineRepo *AfterSalesRepo.TimelineRepository
	orderSync    *OrderSyncService
}

func NewCaseService(deps *app.Deps) *CaseService {
	if deps == nil || deps.DB == nil {
		return &CaseService{deps: deps}
	}
	return &CaseService{
		deps:         deps,
		caseRepo:     AfterSalesRepo.NewCaseRepository(deps.DB),
		timelineRepo: AfterSalesRepo.NewTimelineRepository(deps.DB),
		orderSync:    NewOrderSyncService(deps),
	}
}

func (s *CaseService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.caseRepo != nil && s.timelineRepo != nil
}

func (s *CaseService) List(ctx context.Context, tenantUUID string, query CaseQuery) (*CaseListDTO, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	rows, total, err := s.caseRepo.ListForAdmin(ctx, AfterSalesRepo.AdminCaseListFilter{
		Status:   strings.TrimSpace(query.Status),
		CaseType: strings.TrimSpace(query.CaseType),
		OrderID:  strings.TrimSpace(query.OrderID),
		Keyword:  strings.TrimSpace(query.Keyword),
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		return nil, err
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	items := make([]CaseSummaryDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSummaryDTO(row))
	}
	if err := s.hydrateReverseLinkSummary(ctx, tenantUUID, items); err != nil {
		return nil, err
	}
	resp := &CaseListDTO{Items: items}
	resp.Meta.Total = total
	resp.Meta.Page = query.Page
	resp.Meta.PageSize = query.PageSize
	return resp, nil
}

func (s *CaseService) Detail(ctx context.Context, tenantUUID, caseID string) (*CaseDetailDTO, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	row, err := s.caseRepo.GetByID(ctx, caseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminCaseNotFound
		}
		return nil, err
	}
	timeline, err := s.timelineRepo.ListByCaseID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	summary := toSummaryDTO(*row)
	one := []CaseSummaryDTO{summary}
	if err := s.hydrateReverseLinkSummary(ctx, tenantUUID, one); err != nil {
		return nil, err
	}
	if len(one) > 0 {
		summary = one[0]
	}
	return &CaseDetailDTO{Case: summary, Timeline: toTimelineDTO(timeline)}, nil
}

func (s *CaseService) Transition(ctx context.Context, tenantUUID, caseID, action, operatorID, note string) (*AfterSalesModel.AfterSaleCase, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
	}
	toStatus, err := statusByAction(action)
	if err != nil {
		return nil, err
	}
	ctx = withTenantContext(ctx, tenantUUID)

	var updated *AfterSalesModel.AfterSaleCase
	err = s.caseRepo.WithTenantTx(ctx, func(tx *gorm.DB) error {
		row, err := s.lockCaseByID(ctx, tx, tenantUUID, caseID)
		if err != nil {
			return err
		}
		if err := ValidateCaseStatusTransition(row.Status, toStatus); err != nil {
			return err
		}
		now := time.Now().UTC()
		fromStatus := row.Status
		row.Status = toStatus
		if strings.EqualFold(strings.TrimSpace(row.CaseType), "exchange") &&
			(strings.EqualFold(strings.TrimSpace(action), "approve") || strings.EqualFold(strings.TrimSpace(action), "complete")) {
			note = strings.TrimSpace(strings.Join([]string{
				note,
				"[exchange-boundary] auto reship disabled; auto inventory reserve disabled",
			}, " "))
		}
		if toStatus == CaseStatusClosed {
			row.ClosedAt = &now
		}
		if err := tx.WithContext(ctx).Model(&AfterSalesModel.AfterSaleCase{}).
			Where("tenant_uuid = ? AND id = ?", strings.TrimSpace(tenantUUID), strings.TrimSpace(caseID)).
			Updates(map[string]any{"status": row.Status, "closed_at": row.ClosedAt}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Create(&AfterSalesModel.AfterSaleTimeline{
			TenantUUID:   strings.TrimSpace(tenantUUID),
			CaseID:       row.ID,
			Action:       strings.TrimSpace(strings.ToLower(action)),
			FromStatus:   fromStatus,
			ToStatus:     row.Status,
			OperatorType: "operator",
			OperatorID:   strings.TrimSpace(operatorID),
			Note:         strings.TrimSpace(note),
		}).Error; err != nil {
			return err
		}
		if s.orderSync != nil {
			if err := s.orderSync.SyncCaseTransitionWithTx(ctx, tx, tenantUUID, action, operatorID, note, row); err != nil && !errors.Is(err, ErrOrderSyncUnavailable) {
				return err
			}
		}
		updated = row
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminCaseNotFound
		}
		return nil, err
	}
	return updated, nil
}

// T030: terminal statuses freeze core business fields.
func (s *CaseService) UpdateCoreFields(ctx context.Context, tenantUUID, caseID string, req CoreFieldsUpdateRequest) (*AfterSalesModel.AfterSaleCase, error) {
	if !s.Ready() {
		return nil, ErrAdminServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	var updated *AfterSalesModel.AfterSaleCase
	err := s.caseRepo.WithTenantTx(ctx, func(tx *gorm.DB) error {
		row, err := s.lockCaseByID(ctx, tx, tenantUUID, caseID)
		if err != nil {
			return err
		}
		if isTerminalStatus(row.Status) {
			return ErrCoreFieldsFrozen
		}
		row.CaseType = strings.TrimSpace(strings.ToLower(req.CaseType))
		row.ReasonCode = strings.TrimSpace(req.ReasonCode)
		row.ReasonDetail = strings.TrimSpace(req.ReasonDetail)
		if req.RequestedQty > 0 {
			row.RequestedQty = req.RequestedQty
		}
		if req.RequestedAmountMinor >= 0 {
			row.RequestedAmountMinor = req.RequestedAmountMinor
		}
		if strings.TrimSpace(req.Currency) != "" {
			row.Currency = strings.TrimSpace(req.Currency)
		}
		if err := tx.WithContext(ctx).Model(&AfterSalesModel.AfterSaleCase{}).
			Where("tenant_uuid = ? AND id = ?", strings.TrimSpace(tenantUUID), strings.TrimSpace(caseID)).
			Updates(map[string]any{
				"case_type":              row.CaseType,
				"reason_code":            row.ReasonCode,
				"reason_detail":          row.ReasonDetail,
				"requested_qty":          row.RequestedQty,
				"requested_amount_minor": row.RequestedAmountMinor,
				"currency":               row.Currency,
			}).Error; err != nil {
			return err
		}
		updated = row
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminCaseNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (s *CaseService) lockCaseByID(ctx context.Context, tx *gorm.DB, tenantUUID, caseID string) (*AfterSalesModel.AfterSaleCase, error) {
	query := tx.WithContext(ctx)
	if query.Dialector != nil && query.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var row AfterSalesModel.AfterSaleCase
	if err := query.Where("tenant_uuid = ? AND id = ?", strings.TrimSpace(tenantUUID), strings.TrimSpace(caseID)).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func statusByAction(action string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "accept":
		return CaseStatusAccepted, nil
	case "review":
		return CaseStatusReviewing, nil
	case "approve":
		return CaseStatusApproved, nil
	case "reject":
		return CaseStatusRejected, nil
	case "complete":
		return CaseStatusCompleted, nil
	case "close":
		return CaseStatusClosed, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidAction, action)
	}
}

func isTerminalStatus(status string) bool {
	_, ok := terminalStatuses[NormalizeCaseStatus(status)]
	return ok
}

type reverseLinkSnapshot struct {
	CaseID           string
	ReverseWaybillNo string
	ReceiveStatus    string
	CreatedAt        time.Time
}

func (s *CaseService) hydrateReverseLinkSummary(ctx context.Context, tenantUUID string, items []CaseSummaryDTO) error {
	if s == nil || s.deps == nil || s.deps.DB == nil || len(items) == 0 {
		return nil
	}
	ids := make([]string, 0, len(items))
	index := make(map[string]int, len(items))
	for i := range items {
		id := strings.TrimSpace(items[i].ID)
		if id == "" {
			continue
		}
		ids = append(ids, id)
		index[id] = i
	}
	if len(ids) == 0 {
		return nil
	}
	table := models.S(models.TableAfterSalesReverseLogisticsLinks)
	var rows []reverseLinkSnapshot
	if err := s.deps.DB.WithContext(ctx).
		Table(table).
		Select("case_id, reverse_waybill_no, receive_status, created_at").
		Where("tenant_uuid = ? AND case_id IN ? AND deleted_at IS NULL", strings.TrimSpace(tenantUUID), ids).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.CaseID]; ok {
			continue
		}
		seen[row.CaseID] = struct{}{}
		i, ok := index[row.CaseID]
		if !ok {
			continue
		}
		items[i].ReverseWaybillNo = strings.TrimSpace(row.ReverseWaybillNo)
		items[i].ReverseReceiveStatus = strings.TrimSpace(row.ReceiveStatus)
		linkedAt := row.CreatedAt
		items[i].ReverseLinkedAt = &linkedAt
	}
	return nil
}
