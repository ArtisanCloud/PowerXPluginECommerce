package payments

import (
	"context"
	"errors"
	"strings"
	"time"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/gorm"
)

var (
	ErrTransactionServiceUnavailable = errors.New("payment transaction service unavailable")
	ErrTransactionNotFound           = errors.New("payment transaction not found")
)

type TransactionService struct {
	deps *app.Deps
	repo *paymentrepo.PaymentTransactionRepository
}

func NewTransactionService(deps *app.Deps) *TransactionService {
	if deps == nil || deps.DB == nil {
		return &TransactionService{deps: deps}
	}
	return &TransactionService{deps: deps, repo: paymentrepo.NewPaymentTransactionRepository(deps.DB)}
}

func (s *TransactionService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *TransactionService) ListTransactions(ctx context.Context, tenantUUID, adminID string, filter TransactionListFilter) ([]TransactionDTO, error) {
	if !s.Ready() {
		return nil, ErrTransactionServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	query := s.deps.DB.WithContext(ctx).Model(&models.PaymentTransaction{}).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.Status); v != "" {
		query = query.Where("status = ?", v)
	}
	if filter.ProviderID > 0 {
		query = query.Where("provider_id = ?", filter.ProviderID)
	}
	if v := strings.TrimSpace(filter.OrderID); v != "" {
		query = query.Where("order_id = ?", v)
	}
	if v := strings.TrimSpace(filter.OrderNo); v != "" {
		query = query.Where("order_no = ?", v)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", filter.From.UTC())
	}
	if filter.To != nil {
		query = query.Where("created_at <= ?", filter.To.UTC())
	}
	var rows []models.PaymentTransaction
	if err := query.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]TransactionDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, toTransactionDTO(&row))
	}
	return out, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, tenantUUID, adminID string, id uint64) (*TransactionDTO, error) {
	if !s.Ready() {
		return nil, ErrTransactionServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	var row models.PaymentTransaction
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	out := toTransactionDTO(&row)
	return &out, nil
}

func ParseTimeQuery(raw string) (*time.Time, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func toTransactionDTO(row *models.PaymentTransaction) TransactionDTO {
	if row == nil {
		return TransactionDTO{}
	}
	return TransactionDTO{
		ID:             row.ID,
		TransactionNo:  row.TransactionNo,
		OrderID:        row.OrderID,
		OrderNo:        row.OrderNo,
		ProviderID:     row.ProviderID,
		PayMethod:      row.PayMethod,
		AmountTotal:    row.AmountTotal,
		AmountCurrency: row.AmountCurrency,
		FeeAmount:      row.FeeAmount,
		Status:         row.Status,
		CreatedAt:      row.CreatedAt,
		CompletedAt:    row.CompletedAt,
		FailureReason:  row.FailureReason,
	}
}
