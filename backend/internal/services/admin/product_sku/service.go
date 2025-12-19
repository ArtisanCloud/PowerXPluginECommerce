package product_sku

import (
	"context"
	"errors"

	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/shared/app"
)

// Service groups the SKU level orchestration entrypoints. Concrete logic will be
// implemented in later phases but we expose strongly typed repositories now so
// downstream handlers can start wiring requests.
type Service struct {
	deps             *app.Deps
	SKURepo          *repo.SKURepository
	AttributeRepo    *repo.AttributeRepository
	ChannelRepo      *repo.ChannelRepository
	InventoryRepo    *repo.InventoryRepository
	MediaRepo        *repo.MediaRepository
	BulkTaskRepo     *repo.BulkTaskRepository
	BulkTaskItemRepo *repo.BulkTaskItemRepository
	SerialRepo       *repo.SerialRepository
}

// NewService builds the SKU service scaffold using shared dependencies.
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	return &Service{
		deps:             deps,
		SKURepo:          repo.NewSKURepository(deps.DB),
		AttributeRepo:    repo.NewAttributeRepository(deps.DB),
		ChannelRepo:      repo.NewChannelRepository(deps.DB),
		InventoryRepo:    repo.NewInventoryRepository(deps.DB),
		MediaRepo:        repo.NewMediaRepository(deps.DB),
		BulkTaskRepo:     repo.NewBulkTaskRepository(deps.DB),
		BulkTaskItemRepo: repo.NewBulkTaskItemRepository(deps.DB),
		SerialRepo:       repo.NewSerialRepository(deps.DB),
	}
}

// Ready reports whether the service can start serving requests.
func (s *Service) Ready() bool {
	return s != nil && s.deps != nil && s.deps.DB != nil
}

// NotImplemented is a helper for unfinished orchestration paths.
func (s *Service) NotImplemented() error {
	return errors.New("product SKU service not implemented yet")
}

// HealthProbe is invoked by router wiring to keep future health checks simple.
func (s *Service) HealthProbe(ctx context.Context) error {
	if !s.Ready() {
		return errors.New("product SKU service not ready")
	}
	return ctx.Err()
}
