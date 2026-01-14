package cart

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	cartmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/cart"
	cartrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/cart"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrCartServiceUnavailable = errors.New("cart service unavailable")
	ErrCustomerRequired       = errors.New("customer id is required")
	ErrItemsRequired          = errors.New("items are required")
	ErrInvalidQty             = errors.New("qty must be positive")
)

type Service struct {
	db   *gorm.DB
	repo *cartrepo.Repository
}

func NewService(db *gorm.DB) *Service {
	if db == nil {
		return &Service{}
	}
	return &Service{
		db:   db,
		repo: cartrepo.NewRepository(db),
	}
}

func (s *Service) Ready() bool {
	return s != nil && s.db != nil && s.repo != nil
}

func (s *Service) GetCart(ctx context.Context, tenantUUID, customerID string) (*CartDTO, error) {
	if !s.Ready() {
		return nil, ErrCartServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}

	row, err := s.repo.FindByCustomerID(ctx, tenantUUID, customerID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return &CartDTO{Items: []CartItemInput{}, UpdatedAt: time.Time{}}, nil
	}

	items, err := decodeItems(row.Items)
	if err != nil {
		return nil, err
	}
	return &CartDTO{Items: items, UpdatedAt: row.UpdatedAt}, nil
}

func (s *Service) SyncCart(ctx context.Context, tenantUUID, customerID string, req CartSyncRequest) (*CartDTO, error) {
	if !s.Ready() {
		return nil, ErrCartServiceUnavailable
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, ErrCustomerRequired
	}
	normalized, err := normalizeItems(req.Items)
	if err != nil {
		return nil, err
	}

	strategy := strings.ToLower(strings.TrimSpace(req.Strategy))
	if strategy == "" {
		strategy = "max"
	}
	if strategy != "max" && strategy != "overwrite" {
		return nil, errors.New("invalid strategy")
	}

	var dto *CartDTO
	err = s.repo.WithTenantTx(ctx, tenantUUID, func(tx *gorm.DB) error {
		existing, err := s.repo.LockByCustomerID(ctx, tx, tenantUUID, customerID)
		if err != nil {
			return err
		}

		next := normalized
		if existing != nil && strategy == "max" {
			serverItems, err := decodeItems(existing.Items)
			if err != nil {
				return err
			}
			next = mergeItemsMax(serverItems, normalized)
		}

		raw, _ := json.Marshal(next)
		now := time.Now().UTC()
		if existing == nil {
			row := &cartmodel.Cart{
				ID:         uuid.NewString(),
				TenantUUID: tenantUUID,
				CustomerID: customerID,
				Items:      datatypes.JSON(raw),
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			if err := tx.WithContext(ctx).Create(row).Error; err != nil {
				return err
			}
			dto = &CartDTO{Items: next, UpdatedAt: row.UpdatedAt}
			return nil
		}

		existing.Items = datatypes.JSON(raw)
		existing.UpdatedAt = now
		if err := s.repo.SaveWithTx(ctx, tx, existing); err != nil {
			return err
		}
		dto = &CartDTO{Items: next, UpdatedAt: existing.UpdatedAt}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dto, nil
}

func normalizeItems(items []CartItemInput) ([]CartItemInput, error) {
	if len(items) == 0 {
		return []CartItemInput{}, nil
	}
	merged := map[string]int64{}
	for _, it := range items {
		sku := strings.TrimSpace(it.SKUID)
		if sku == "" {
			return nil, errors.New("skuId is required")
		}
		if it.Qty <= 0 {
			return nil, ErrInvalidQty
		}
		merged[sku] += it.Qty
	}
	out := make([]CartItemInput, 0, len(merged))
	for sku, qty := range merged {
		out = append(out, CartItemInput{SKUID: sku, Qty: qty})
	}
	return out, nil
}

func mergeItemsMax(server, client []CartItemInput) []CartItemInput {
	m := map[string]int64{}
	for _, it := range server {
		sku := strings.TrimSpace(it.SKUID)
		if sku == "" || it.Qty <= 0 {
			continue
		}
		m[sku] = it.Qty
	}
	for _, it := range client {
		sku := strings.TrimSpace(it.SKUID)
		if sku == "" || it.Qty <= 0 {
			continue
		}
		if cur, ok := m[sku]; !ok || it.Qty > cur {
			m[sku] = it.Qty
		}
	}
	out := make([]CartItemInput, 0, len(m))
	for sku, qty := range m {
		out = append(out, CartItemInput{SKUID: sku, Qty: qty})
	}
	return out
}

func decodeItems(raw datatypes.JSON) ([]CartItemInput, error) {
	if len(raw) == 0 {
		return []CartItemInput{}, nil
	}
	var items []CartItemInput
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	normalized, err := normalizeItems(items)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}
