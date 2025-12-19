package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// InventoryRepository manages per-warehouse inventory snapshots.
type InventoryRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUInventory]
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUInventory](db)}
}
