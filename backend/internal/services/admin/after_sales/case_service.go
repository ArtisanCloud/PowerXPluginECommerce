package after_sales

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
}

func NewCaseService(deps *app.Deps) *CaseService {
	if deps == nil || deps.DB == nil {
		return &CaseService{deps: deps}
	}
	return &CaseService{
		deps:         deps,
		caseRepo:     AfterSalesRepo.NewCaseRepository(deps.DB),
		timelineRepo: AfterSalesRepo.NewTimelineRepository(deps.DB),
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
	return &CaseDetailDTO{Case: toSummaryDTO(*row), Timeline: toTimelineDTO(timeline)}, nil
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
