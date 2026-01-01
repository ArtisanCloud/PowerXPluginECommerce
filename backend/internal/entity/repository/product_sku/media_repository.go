package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// MediaRepository persists SKU level images/videos.
type MediaRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUMedia]
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUMedia](db)}
}
