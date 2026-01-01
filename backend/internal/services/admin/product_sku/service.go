package product_sku

import (
	"context"
	"errors"
	"strings"

	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_sku"
	authx "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	productskulogger "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/observability/product_sku"
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
	logger           *productskulogger.Logger
}

// NewService builds the SKU service scaffold using shared dependencies.
func NewService(deps *app.Deps) *Service {
	if deps == nil || deps.DB == nil {
		return &Service{deps: deps}
	}
	var skuLogger *productskulogger.Logger
	if entry := deps.RuntimeLogger(deps.Ctx, "product_sku", nil); entry != nil {
		skuLogger = productskulogger.NewLogger(entry)
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
		logger:           skuLogger,
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

func (s *Service) tenantFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("request context missing")
	}
	if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
		return tid, nil
	}
	return "", errors.New("tenant context missing")
}

func (s *Service) emitEvent(ctx context.Context, action string, payload map[string]any) {
	if s == nil || s.logger == nil {
		return
	}
	var metadata map[string]any
	if len(payload) > 0 {
		metadata = payload
	}
	event := productskulogger.Event{
		Action:   action,
		Metadata: metadata,
	}
	if ctx != nil {
		if tid, ok := authx.TenantUUIDFromContext(ctx); ok && strings.TrimSpace(tid) != "" {
			event.TenantID = tid
			if payload == nil {
				payload = map[string]any{}
			}
			if _, exists := payload["tenant_id"]; !exists {
				payload["tenant_id"] = tid
			}
			event.Metadata = payload
		}
		if reqID, ok := ctx.Value("request_id").(string); ok && strings.TrimSpace(reqID) != "" {
			event.RequestID = strings.TrimSpace(reqID)
		}
	}
	s.logger.EmitEvent(event)
}

// ListSkus returns paginated SKU summaries for the tenant in context.
func (s *Service) ListSkus(ctx context.Context, query SkuListQuery) (*SkuListResult, error) {
	if s == nil || !s.Ready() {
		return nil, errors.New("product SKU service not ready")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	filters := repo.SkuListFilters{
		SPUID:    strings.TrimSpace(query.SPUID),
		Status:   strings.TrimSpace(query.Status),
		Keyword:  strings.TrimSpace(query.Keyword),
		Page:     query.Page,
		PageSize: query.PageSize,
	}
	rows, total, err := s.SKURepo.List(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}
	items := make([]SkuListItem, len(rows))
	for i, row := range rows {
		createdAt := row.CreatedAt
		updatedAt := row.UpdatedAt
		items[i] = SkuListItem{
			ID:        row.ID,
			SPUID:     row.SPUID,
			SKUCode:   row.SKUCode,
			Status:    row.Status,
			Barcode:   row.Barcode,
			CreatedAt: &createdAt,
			UpdatedAt: &updatedAt,
		}
	}
	return &SkuListResult{
		Items:    items,
		Page:     filters.Page,
		PageSize: filters.PageSize,
		Total:    total,
	}, nil
}
