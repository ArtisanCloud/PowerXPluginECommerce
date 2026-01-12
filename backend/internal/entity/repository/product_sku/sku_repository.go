package product_sku

import (
	"context"
	"errors"
	"strings"

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

// SkuExportFilters are used when exporting SKU snapshots.
type SkuExportFilters struct {
	SPUID   string
	Status  string
	Keyword string
	Limit   int
}

// SkuListFilters control pagination + filtering for list endpoints.
type SkuListFilters struct {
	SPUID    string
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

// ListForExport lists SKUs for CSV export given tenant scoped filters.
func (r *SKURepository) ListForExport(ctx context.Context, tenantID string, filters SkuExportFilters) ([]productskumodel.ProductSKU, error) {
	query := r.DB.WithContext(ctx).Where("tenant_uuid = ?", tenantID)
	if filters.SPUID != "" {
		query = query.Where("spu_id = ?", filters.SPUID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Keyword != "" {
		keyword := strings.ToLower(filters.Keyword)
		pattern := "%" + keyword + "%"
		query = query.Where("(LOWER(sku_code) LIKE ? OR LOWER(barcode) LIKE ?)", pattern, pattern)
	}
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	var skus []productskumodel.ProductSKU
	if err := query.Order("created_at DESC").Find(&skus).Error; err != nil {
		return nil, err
	}
	return skus, nil
}

// List returns paginated SKU rows along with total count.
func (r *SKURepository) List(ctx context.Context, tenantID string, filters SkuListFilters) ([]productskumodel.ProductSKU, int64, error) {
	if r == nil || r.DB == nil {
		return nil, 0, errors.New("sku repository is not initialized")
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	} else if filters.PageSize > 200 {
		filters.PageSize = 200
	}

	query := r.DB.WithContext(ctx).Model(&productskumodel.ProductSKU{}).
		Where("tenant_uuid = ?", tenantID)
	if strings.TrimSpace(filters.SPUID) != "" {
		query = query.Where("spu_id = ?", strings.TrimSpace(filters.SPUID))
	}
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := strings.ToLower(filters.Keyword)
		pattern := "%" + keyword + "%"
		query = query.Where("(LOWER(sku_code) LIKE ? OR LOWER(barcode) LIKE ?)", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []productskumodel.ProductSKU{}, 0, nil
	}

	var skus []productskumodel.ProductSKU
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Order("created_at DESC").
		Limit(filters.PageSize).
		Offset(offset).
		Find(&skus).Error; err != nil {
		return nil, 0, err
	}
	return skus, total, nil
}

// FindByBarcode locates a SKU by barcode for uniqueness validation.
func (r *SKURepository) FindByBarcode(ctx context.Context, tenantID, barcode string) (*productskumodel.ProductSKU, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("sku repository is not initialized")
	}
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return nil, errors.New("barcode is required")
	}
	var sku productskumodel.ProductSKU
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND barcode = ?", tenantID, barcode).
		First(&sku).Error; err != nil {
		return nil, err
	}
	return &sku, nil
}

// FindByID locates a SKU by id scoped to the tenant.
func (r *SKURepository) FindByID(ctx context.Context, tenantID, skuID string) (*productskumodel.ProductSKU, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("sku repository is not initialized")
	}
	skuID = strings.TrimSpace(skuID)
	if skuID == "" {
		return nil, errors.New("sku id is required")
	}
	var sku productskumodel.ProductSKU
	if err := r.DB.WithContext(ctx).
		Where("tenant_uuid = ? AND id = ? AND deleted_at IS NULL", tenantID, skuID).
		First(&sku).Error; err != nil {
		return nil, err
	}
	return &sku, nil
}
