package product_sku

import (
	"context"
	"errors"

	productskumodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_sku"
)

// ListBySPU returns all SKU records for a tenant + SPU combination.
func (r *SKURepository) ListBySPU(ctx context.Context, tenantID, spuID string) ([]productskumodel.ProductSKU, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("sku repository is not initialized")
	}
	var rows []productskumodel.ProductSKU
	err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND spu_id = ?", tenantID, spuID).
		Find(&rows).Error
	return rows, err
}
