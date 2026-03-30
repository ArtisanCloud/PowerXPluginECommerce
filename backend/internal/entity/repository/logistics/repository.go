package logistics

import (
	"context"

	BaseRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// Repository wraps BaseRepository and enforces tenant transaction scope.
type Repository[T any] struct {
	*BaseRepo.BaseRepository[T]
}

// NewRepository creates a tenant-aware repository wrapper.
func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{BaseRepository: BaseRepo.NewBaseRepository[T](db)}
}

// BeginTenantTx binds transaction scope to tenant_uuid from request context.
func (r *Repository[T]) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := AuthX.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

// WithTenantTx executes callback inside tenant-scoped transaction.
func (r *Repository[T]) WithTenantTx(ctx context.Context, fn func(*gorm.DB) error) error {
	tenantUUID, err := AuthX.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	return r.BaseRepository.WithTenantTx(ctx, tenantUUID, fn)
}
