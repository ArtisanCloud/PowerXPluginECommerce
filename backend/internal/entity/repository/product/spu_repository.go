package product

import (
	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// SPURepository wraps BaseRepository to expose tenant-aware helpers for SPU entities.
type SPURepository struct {
	*repo.BaseRepository[productmodel.SPU]
}

// NewSPURepository constructs a repository backed by the shared BaseRepository helpers.
func NewSPURepository(db *gorm.DB) *SPURepository {
	return &SPURepository{BaseRepository: repo.NewBaseRepository[productmodel.SPU](db)}
}
