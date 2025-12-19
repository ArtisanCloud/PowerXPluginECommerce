package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// SKURepository exposes helpers for CRUD on ProductSKU entities.
type SKURepository struct {
	*repo.BaseRepository[productskumodel.ProductSKU]
}

// NewSKURepository constructs the repository backed by the shared BaseRepository.
func NewSKURepository(db *gorm.DB) *SKURepository {
	return &SKURepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKU](db)}
}
