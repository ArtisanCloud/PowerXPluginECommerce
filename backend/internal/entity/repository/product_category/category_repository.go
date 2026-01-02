package product_category

import (
	"context"
	"errors"
	"strings"

	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CategoryRepository provides persistence helpers for product categories.
type CategoryRepository struct {
	*repository.BaseRepository[productcategory.ProductCategory]
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: repository.NewBaseRepository[productcategory.ProductCategory](db),
	}
}

type CategoryListFilters struct {
	Keyword  string
	Status   string
	ParentID *string
	Page     int
	PageSize int
}

func (r *CategoryRepository) GetByID(ctx context.Context, tenantUUID string, id string, lock bool) (*productcategory.ProductCategory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("category repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" || id == "" {
		return nil, errors.New("tenant_uuid and id are required")
	}
	query := r.DB.WithContext(ctx).Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var record productcategory.ProductCategory
	if err := query.First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *CategoryRepository) List(ctx context.Context, tenantUUID string, filters CategoryListFilters) ([]productcategory.ProductCategory, int64, error) {
	if r == nil || r.DB == nil {
		return nil, 0, errors.New("category repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, 0, errors.New("tenant_uuid is required")
	}
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	query := r.DB.WithContext(ctx).Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ?", tenantUUID)
	if filters.ParentID != nil {
		if strings.TrimSpace(*filters.ParentID) == "" {
			query = query.Where("parent_id IS NULL")
		} else {
			query = query.Where("parent_id = ?", strings.TrimSpace(*filters.ParentID))
		}
	}
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.Keyword) != "" {
		like := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where("code ILIKE ? OR display_name ILIKE ? OR alias_slug ILIKE ?", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []productcategory.ProductCategory
	if err := query.Order("level ASC, sort_order ASC, display_name ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *CategoryRepository) ListAllForTree(ctx context.Context, tenantUUID string, includeDisabled bool) ([]productcategory.ProductCategory, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("category repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	if tenantUUID == "" {
		return nil, errors.New("tenant_uuid is required")
	}
	query := r.DB.WithContext(ctx).Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ?", tenantUUID)
	if !includeDisabled {
		query = query.Where("status = ?", productcategory.CategoryStatusEnabled)
	}
	var records []productcategory.ProductCategory
	if err := query.Order("level ASC, sort_order ASC, display_name ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *CategoryRepository) CountChildren(ctx context.Context, tenantUUID string, id string) (int64, error) {
	if r == nil || r.DB == nil {
		return 0, errors.New("category repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" || id == "" {
		return 0, errors.New("tenant_uuid and id are required")
	}
	var total int64
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ? AND parent_id = ?", tenantUUID, id).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *CategoryRepository) CountByPathPrefix(ctx context.Context, tenantUUID string, pathPrefix string) (int64, error) {
	if r == nil || r.DB == nil {
		return 0, errors.New("category repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	pathPrefix = strings.TrimSpace(pathPrefix)
	if tenantUUID == "" || pathPrefix == "" {
		return 0, errors.New("tenant_uuid and pathPrefix are required")
	}
	var total int64
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ? AND path LIKE ?", tenantUUID, pathPrefix+"%").
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *CategoryRepository) UpdateSubtreePathAndLevel(
	ctx context.Context,
	tx *gorm.DB,
	tenantUUID string,
	oldPrefix string,
	newPrefix string,
	levelDelta int,
) error {
	if r == nil {
		return errors.New("category repository not initialized")
	}
	if tx == nil {
		tx = r.DB
	}
	if tx == nil {
		return errors.New("category repository database is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	oldPrefix = strings.TrimSpace(oldPrefix)
	newPrefix = strings.TrimSpace(newPrefix)
	if tenantUUID == "" || oldPrefix == "" || newPrefix == "" {
		return errors.New("tenant_uuid, oldPrefix and newPrefix are required")
	}
	start := len(oldPrefix) + 1
	pathExpr := "? || substring(path from ?)"
	if tx.Dialector != nil && strings.EqualFold(tx.Dialector.Name(), "sqlite") {
		pathExpr = "? || substr(path, ?)"
	}
	updates := map[string]any{
		"path":  gorm.Expr(pathExpr, newPrefix, start),
		"level": gorm.Expr("level + ?", levelDelta),
	}
	return tx.WithContext(ctx).
		Model(&productcategory.ProductCategory{}).
		Where("tenant_uuid = ? AND path LIKE ?", tenantUUID, oldPrefix+"%").
		Updates(updates).Error
}
