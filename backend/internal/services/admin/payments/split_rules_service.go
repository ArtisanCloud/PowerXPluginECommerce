package payments

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

var ErrSplitRuleServiceUnavailable = errors.New("payment split rule service unavailable")

type SplitRuleService struct {
	deps *app.Deps
	repo *paymentrepo.PaymentSplitRuleRepository
}

func NewSplitRuleService(deps *app.Deps) *SplitRuleService {
	if deps == nil || deps.DB == nil {
		return &SplitRuleService{deps: deps}
	}
	return &SplitRuleService{deps: deps, repo: paymentrepo.NewPaymentSplitRuleRepository(deps.DB)}
}

func (s *SplitRuleService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *SplitRuleService) ListSplitRules(ctx context.Context, tenantUUID, adminID string) ([]SplitRuleDTO, error) {
	if !s.Ready() {
		return nil, ErrSplitRuleServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	var rows []models.PaymentSplitRule
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantUUID).Order("updated_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]SplitRuleDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, SplitRuleDTO{
			ID:        row.ID,
			Name:      row.Name,
			Status:    row.Status,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}
