package product_sku

import (
	"context"
	"errors"
	"fmt"
	"strings"

	basemodels "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models"
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
	Locale   string
	Page     int
	PageSize int
}

// ListForExport lists SKUs for CSV export given tenant scoped filters.
func (r *SKURepository) ListForExport(ctx context.Context, tenantID string, filters SkuExportFilters) ([]productskumodel.ProductSKU, error) {
	skuTable := basemodels.S(basemodels.TableProductSkus)
	query := r.DB.WithContext(ctx).Table(skuTable).
		Where(fmt.Sprintf("%s.tenant_uuid = ?", skuTable), tenantID)
	if filters.SPUID != "" {
		query = query.Where(fmt.Sprintf("%s.spu_id = ?", skuTable), filters.SPUID)
	}
	if filters.Status != "" {
		query = query.Where(fmt.Sprintf("%s.status = ?", skuTable), filters.Status)
	}
	if filters.Keyword != "" {
		keyword := strings.ToLower(filters.Keyword)
		pattern := "%" + keyword + "%"
		spuTable := basemodels.S(basemodels.TableProductSpus)
		query = query.Joins(
			fmt.Sprintf(
				"LEFT JOIN %s spu ON spu.id = %s.spu_id AND spu.tenant_uuid = %s.tenant_uuid AND spu.deleted_at IS NULL",
				spuTable,
				skuTable,
				skuTable,
			),
		)
		query = query.Where(
			fmt.Sprintf(
				"(LOWER(%s.sku_code) LIKE ? OR LOWER(%s.barcode) LIKE ? OR LOWER(spu.name) LIKE ? OR LOWER(spu.code) LIKE ?)",
				skuTable,
				skuTable,
			),
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	var skus []productskumodel.ProductSKU
	if err := query.Order(fmt.Sprintf("%s.created_at DESC", skuTable)).Find(&skus).Error; err != nil {
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

	skuTable := basemodels.S(basemodels.TableProductSkus)
	query := r.DB.WithContext(ctx).Table(skuTable).
		Where(fmt.Sprintf("%s.tenant_uuid = ?", skuTable), tenantID)
	if strings.TrimSpace(filters.SPUID) != "" {
		query = query.Where(fmt.Sprintf("%s.spu_id = ?", skuTable), strings.TrimSpace(filters.SPUID))
	}
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where(fmt.Sprintf("%s.status = ?", skuTable), strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.Keyword) != "" {
		keyword := strings.ToLower(filters.Keyword)
		pattern := "%" + keyword + "%"
		spuTable := basemodels.S(basemodels.TableProductSpus)
		locale := strings.TrimSpace(filters.Locale)
		query = query.Joins(
			fmt.Sprintf(
				"LEFT JOIN %s spu ON spu.id = %s.spu_id AND spu.tenant_uuid = %s.tenant_uuid AND spu.deleted_at IS NULL",
				spuTable,
				skuTable,
				skuTable,
			),
		)
		if locale != "" {
			spuLocaleTable := basemodels.S(basemodels.TableProductSpuLocales)
			query = query.Joins(
				fmt.Sprintf(
					"LEFT JOIN %s spuloc ON spuloc.spu_id = %s.spu_id AND spuloc.tenant_uuid = %s.tenant_uuid AND spuloc.locale = ?",
					spuLocaleTable,
					skuTable,
					skuTable,
				),
				locale,
			)
			query = query.Where(
				fmt.Sprintf(
					"(LOWER(%s.sku_code) LIKE ? OR LOWER(%s.barcode) LIKE ? OR LOWER(spu.name) LIKE ? OR LOWER(spu.code) LIKE ? OR LOWER(COALESCE(spuloc.title, '')) LIKE ?)",
					skuTable,
					skuTable,
				),
				pattern,
				pattern,
				pattern,
				pattern,
				pattern,
			)
		} else {
			query = query.Where(
				fmt.Sprintf(
					"(LOWER(%s.sku_code) LIKE ? OR LOWER(%s.barcode) LIKE ? OR LOWER(spu.name) LIKE ? OR LOWER(spu.code) LIKE ?)",
					skuTable,
					skuTable,
				),
				pattern,
				pattern,
				pattern,
				pattern,
			)
		}
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
	if err := query.Order(fmt.Sprintf("%s.created_at DESC", skuTable)).
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
