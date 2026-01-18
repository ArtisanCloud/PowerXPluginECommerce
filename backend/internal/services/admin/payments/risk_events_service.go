package payments

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

var ErrRiskEventServiceUnavailable = errors.New("payment risk event service unavailable")

type RiskEventService struct {
	deps *app.Deps
	repo *paymentrepo.PaymentRiskEventRepository
}

func NewRiskEventService(deps *app.Deps) *RiskEventService {
	if deps == nil || deps.DB == nil {
		return &RiskEventService{deps: deps}
	}
	return &RiskEventService{deps: deps, repo: paymentrepo.NewPaymentRiskEventRepository(deps.DB)}
}

func (s *RiskEventService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *RiskEventService) ListRiskEvents(ctx context.Context, tenantUUID, adminID string, transactionID *uint64) ([]RiskEventDTO, error) {
	if !s.Ready() {
		return nil, ErrRiskEventServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	query := s.deps.DB.WithContext(ctx).Model(&models.PaymentRiskEvent{}).Where("tenant_uuid = ?", tenantUUID)
	if transactionID != nil {
		query = query.Where("transaction_id = ?", *transactionID)
	}
	var rows []models.PaymentRiskEvent
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]RiskEventDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, RiskEventDTO{
			ID:         row.ID,
			RiskType:   row.RiskType,
			RiskScore:  row.RiskScore,
			Action:     row.Action,
			CreatedAt:  row.CreatedAt,
			ResolvedAt: row.ResolvedAt,
		})
	}
	return out, nil
}
