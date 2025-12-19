package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// AttributeRepository manages SKU attribute associations.
type AttributeRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUAttribute]
}

func NewAttributeRepository(db *gorm.DB) *AttributeRepository {
	return &AttributeRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUAttribute](db)}
}
