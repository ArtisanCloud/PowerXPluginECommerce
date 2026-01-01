package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// SerialRepository manages optional serial/batch records for SKUs.
type SerialRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUSerialRecord]
}

func NewSerialRepository(db *gorm.DB) *SerialRepository {
	return &SerialRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUSerialRecord](db)}
}
