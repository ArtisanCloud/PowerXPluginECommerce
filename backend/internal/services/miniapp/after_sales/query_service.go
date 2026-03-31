package after_sales

import (
	"context"
	"errors"
	"strings"

	AfterSalesRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/after_sales"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

type QueryService struct {
	deps         *app.Deps
	caseRepo     *AfterSalesRepo.CaseRepository
	timelineRepo *AfterSalesRepo.TimelineRepository
}

func NewQueryService(deps *app.Deps) *QueryService {
	if deps == nil || deps.DB == nil {
		return &QueryService{deps: deps}
	}
	return &QueryService{
		deps:         deps,
		caseRepo:     AfterSalesRepo.NewCaseRepository(deps.DB),
		timelineRepo: AfterSalesRepo.NewTimelineRepository(deps.DB),
	}
}

func (s *QueryService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.caseRepo != nil && s.timelineRepo != nil
}

func (s *QueryService) ListByCustomer(ctx context.Context, tenantUUID, customerID, status string, page, pageSize int) (*CaseListDTO, error) {
	if !s.Ready() {
		return nil, ErrServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	rows, total, err := s.caseRepo.ListByCustomer(ctx, customerID, strings.TrimSpace(status), page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]CaseSummaryDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSummaryDTO(row))
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return &CaseListDTO{Items: items, Meta: CaseListMeta{Total: total, Page: page, PageSize: pageSize}}, nil
}

func (s *QueryService) GetDetailByCustomer(ctx context.Context, tenantUUID, customerID, caseID string) (*CaseDetailDTO, error) {
	if !s.Ready() {
		return nil, ErrServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	row, err := s.caseRepo.GetByIDAndCustomer(ctx, strings.TrimSpace(caseID), customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCaseNotFound
		}
		return nil, err
	}
	timeline, err := s.timelineRepo.ListByCaseID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return &CaseDetailDTO{Case: toSummaryDTO(*row), Timeline: toTimelineDTO(timeline)}, nil
}
