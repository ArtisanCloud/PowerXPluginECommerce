package product_sku

import (
	"context"
	"errors"

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

// ListBySKU returns all warehouse snapshots for the sku.
func (r *InventoryRepository) ListBySKU(ctx context.Context, tenantID, skuID string) ([]productskumodel.ProductSKUInventory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("inventory repository is not initialized")
	}
	var rows []productskumodel.ProductSKUInventory
	err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND sku_id = ?", tenantID, skuID).
		Order("warehouse_id ASC").
		Find(&rows).Error
	return rows, err
}
