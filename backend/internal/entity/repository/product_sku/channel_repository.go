package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// ChannelRepository persists channel mapping metadata for SKU variants.
type ChannelRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUChannel]
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUChannel](db)}
}
