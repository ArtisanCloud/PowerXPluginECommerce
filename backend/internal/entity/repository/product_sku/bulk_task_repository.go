package product_sku

import (
	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
	repo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// BulkTaskRepository exposes helpers for task level operations.
type BulkTaskRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUBulkTask]
}

func NewBulkTaskRepository(db *gorm.DB) *BulkTaskRepository {
	return &BulkTaskRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUBulkTask](db)}
}

// BulkTaskItemRepository persists per SKU execution results.
type BulkTaskItemRepository struct {
	*repo.BaseRepository[productskumodel.ProductSKUBulkTaskItem]
}

func NewBulkTaskItemRepository(db *gorm.DB) *BulkTaskItemRepository {
	return &BulkTaskItemRepository{BaseRepository: repo.NewBaseRepository[productskumodel.ProductSKUBulkTaskItem](db)}
}
