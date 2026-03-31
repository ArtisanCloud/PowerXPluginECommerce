package after_sales

import (
	"context"

	BaseRepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	AuthX "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/middleware"
	"gorm.io/gorm"
)

// Repository wraps BaseRepository with tenant transaction helpers.
type Repository[T any] struct {
	*BaseRepo.BaseRepository[T]
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{BaseRepository: BaseRepo.NewBaseRepository[T](db)}
}

func (r *Repository[T]) BeginTenantTx(ctx context.Context) (*gorm.DB, error) {
	tenantUUID, err := AuthX.RequireTenantUUID(ctx)
	if err != nil {
		return nil, err
	}
	return r.BaseRepository.BeginTenantTx(ctx, tenantUUID)
}

func (r *Repository[T]) WithTenantTx(ctx context.Context, fn func(*gorm.DB) error) error {
	tenantUUID, err := AuthX.RequireTenantUUID(ctx)
	if err != nil {
		return err
	}
	return r.BaseRepository.WithTenantTx(ctx, tenantUUID, fn)
}
