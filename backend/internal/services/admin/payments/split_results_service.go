package payments

import (
	"context"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

var ErrSplitResultServiceUnavailable = errors.New("payment split result service unavailable")

type SplitResultService struct {
	deps *app.Deps
	repo *paymentrepo.PaymentSplitResultRepository
}

func NewSplitResultService(deps *app.Deps) *SplitResultService {
	if deps == nil || deps.DB == nil {
		return &SplitResultService{deps: deps}
	}
	return &SplitResultService{deps: deps, repo: paymentrepo.NewPaymentSplitResultRepository(deps.DB)}
}

func (s *SplitResultService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *SplitResultService) ListSplitResults(ctx context.Context, tenantUUID, adminID string, transactionID *uint64) ([]SplitResultDTO, error) {
	if !s.Ready() {
		return nil, ErrSplitResultServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	query := s.deps.DB.WithContext(ctx).Model(&models.PaymentSplitResult{}).Where("tenant_uuid = ?", tenantUUID)
	if transactionID != nil {
		query = query.Where("transaction_id = ?", *transactionID)
	}
	var rows []models.PaymentSplitResult
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]SplitResultDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, SplitResultDTO{
			ID:          row.ID,
			RuleID:      row.RuleID,
			Participant: row.Participant,
			Amount:      row.Amount,
			Status:      row.Status,
			CreatedAt:   row.CreatedAt,
		})
	}
	return out, nil
}
