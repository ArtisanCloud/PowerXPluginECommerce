package payments

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	models "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
	paymentrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrProviderServiceUnavailable = errors.New("payment provider service unavailable")
	ErrProviderNotFound           = errors.New("payment provider not found")
	ErrProviderNameRequired       = errors.New("provider name is required")
)

type ProviderService struct {
	deps *app.Deps
	repo *paymentrepo.PaymentProviderRepository
}

func NewProviderService(deps *app.Deps) *ProviderService {
	if deps == nil || deps.DB == nil {
		return &ProviderService{deps: deps}
	}
	return &ProviderService{deps: deps, repo: paymentrepo.NewPaymentProviderRepository(deps.DB)}
}

func (s *ProviderService) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil && s.repo != nil
}

func (s *ProviderService) ListProviders(ctx context.Context, tenantUUID, adminID string, filter ProviderListFilter) ([]ProviderDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	query := s.deps.DB.WithContext(ctx).Model(&models.PaymentProvider{}).Where("tenant_uuid = ?", tenantUUID)
	if v := strings.TrimSpace(filter.Status); v != "" {
		query = query.Where("status = ?", v)
	}
	if v := strings.TrimSpace(filter.ProviderType); v != "" {
		query = query.Where("provider_type = ?", v)
	}
	if v := strings.TrimSpace(filter.Keyword); v != "" {
		query = query.Where("name ILIKE ?", "%"+v+"%")
	}
	var rows []models.PaymentProvider
	if err := query.Order("updated_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]ProviderDTO, 0, len(rows))
	for _, row := range rows {
		out = append(out, ProviderDTO{
			ID:              row.ID,
			Name:            row.Name,
			ProviderType:    row.ProviderType,
			Status:          row.Status,
			FeeRate:         row.FeeRate,
			SettlementCycle: row.SettlementCycle,
			Currency:        row.Currency,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

func (s *ProviderService) CreateProvider(ctx context.Context, tenantUUID, adminID string, req CreateProviderRequest) (*ProviderDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrProviderNameRequired
	}
	credRaw, _ := json.Marshal(req.Credentials)
	riskRaw, _ := json.Marshal(req.RiskPolicy)
	row := &models.PaymentProvider{
		BaseModel:       models.BaseModel{TenantUuid: strings.TrimSpace(tenantUUID)},
		Name:            strings.TrimSpace(req.Name),
		ProviderType:    strings.TrimSpace(req.Type),
		Status:          strings.TrimSpace(req.Status),
		FeeRate:         req.FeeRate,
		SettlementCycle: strings.TrimSpace(req.SettlementCycle),
		Currency:        strings.TrimSpace(req.Currency),
		Credentials:     datatypes.JSON(credRaw),
		RiskPolicy:      datatypes.JSON(riskRaw),
	}
	var created *models.PaymentProvider
	err := s.repo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ProviderDTO{
		ID:              created.ID,
		Name:            created.Name,
		ProviderType:    created.ProviderType,
		Status:          created.Status,
		FeeRate:         created.FeeRate,
		SettlementCycle: created.SettlementCycle,
		Currency:        created.Currency,
		UpdatedAt:       created.UpdatedAt,
	}, nil
}

func (s *ProviderService) UpdateProvider(ctx context.Context, tenantUUID, adminID string, id uint64, req UpdateProviderRequest) (*ProviderDTO, error) {
	if !s.Ready() {
		return nil, ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return nil, errors.New("admin id is required")
	}
	var updated *models.PaymentProvider
	err := s.repo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		var row models.PaymentProvider
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProviderNotFound
			}
			return err
		}
		if req.Status != nil {
			row.Status = strings.TrimSpace(*req.Status)
		}
		if req.FeeRate != nil {
			row.FeeRate = *req.FeeRate
		}
		if req.SettlementCycle != nil {
			row.SettlementCycle = strings.TrimSpace(*req.SettlementCycle)
		}
		if req.Currency != nil {
			row.Currency = strings.TrimSpace(*req.Currency)
		}
		if req.Credentials != nil {
			credRaw, _ := json.Marshal(req.Credentials)
			row.Credentials = datatypes.JSON(credRaw)
		}
		if req.RiskPolicy != nil {
			riskRaw, _ := json.Marshal(req.RiskPolicy)
			row.RiskPolicy = datatypes.JSON(riskRaw)
		}
		if err := tx.WithContext(ctx).Save(&row).Error; err != nil {
			return err
		}
		updated = &row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ProviderDTO{
		ID:              updated.ID,
		Name:            updated.Name,
		ProviderType:    updated.ProviderType,
		Status:          updated.Status,
		FeeRate:         updated.FeeRate,
		SettlementCycle: updated.SettlementCycle,
		Currency:        updated.Currency,
		UpdatedAt:       updated.UpdatedAt,
	}, nil
}

func (s *ProviderService) TestProvider(ctx context.Context, tenantUUID, adminID string, id uint64) error {
	if !s.Ready() {
		return ErrProviderServiceUnavailable
	}
	if strings.TrimSpace(adminID) == "" {
		return errors.New("admin id is required")
	}
	var row models.PaymentProvider
	if err := s.deps.DB.WithContext(ctx).Where("tenant_uuid = ? AND id = ?", tenantUUID, id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProviderNotFound
		}
		return err
	}
	return nil
}
