package after_sales

import (
	"context"
	"fmt"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

type DashboardService struct {
	deps *app.Deps
}

func NewDashboardService(deps *app.Deps) *DashboardService {
	return &DashboardService{deps: deps}
}

func (s *DashboardService) Snapshot(ctx context.Context, tenantUUID string) (*DashboardSnapshot, error) {
	if s == nil || s.deps == nil || s.deps.DB == nil {
		return nil, ErrAdminServiceUnavailable
	}
	ctx = withTenantContext(ctx, tenantUUID)
	table := models.S(models.TableAfterSalesCases)
	query := func(statuses []string) (int64, error) {
		var count int64
		if err := s.deps.DB.WithContext(ctx).
			Table(table).
			Where("tenant_uuid = ? AND status IN ?", strings.TrimSpace(tenantUUID), statuses).
			Count(&count).Error; err != nil {
			return 0, err
		}
		return count, nil
	}

	pendingCount, err := query([]string{CaseStatusPending})
	if err != nil {
		return nil, fmt.Errorf("count pending failed: %w", err)
	}
	processingCount, err := query([]string{CaseStatusAccepted, CaseStatusReviewing, CaseStatusApproved})
	if err != nil {
		return nil, fmt.Errorf("count processing failed: %w", err)
	}
	completedCount, err := query([]string{CaseStatusCompleted, CaseStatusClosed})
	if err != nil {
		return nil, fmt.Errorf("count completed failed: %w", err)
	}
	rejectedCount, err := query([]string{CaseStatusRejected})
	if err != nil {
		return nil, fmt.Errorf("count rejected failed: %w", err)
	}

	return &DashboardSnapshot{
		PendingCount:    pendingCount,
		ProcessingCount: processingCount,
		CompletedCount:  completedCount,
		RejectedCount:   rejectedCount,
	}, nil
}
