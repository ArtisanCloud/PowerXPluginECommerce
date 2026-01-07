package product_category

import (
	"context"
	"errors"
	"strings"

	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	repository "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository"
	"gorm.io/gorm"
)

// TemplateRepository persists category templates and versions.
type TemplateRepository struct {
	*repository.BaseRepository[productcategory.CategoryTemplate]
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{
		BaseRepository: repository.NewBaseRepository[productcategory.CategoryTemplate](db),
	}
}

type TemplateListFilters struct {
	Keyword  string
	Status   string
	Page     int
	PageSize int
}

func (r *TemplateRepository) GetTemplate(ctx context.Context, tenantUUID string, id string) (*productcategory.CategoryTemplate, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("template repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	id = strings.TrimSpace(id)
	if tenantUUID == "" || id == "" {
		return nil, errors.New("tenant_uuid and id are required")
	}
	var record productcategory.CategoryTemplate
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplate{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, id).
		First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *TemplateRepository) ListTemplates(ctx context.Context, tenantUUID string, filters TemplateListFilters) ([]productcategory.CategoryTemplate, int64, error) {
	if r == nil || r.DB == nil {
		return nil, 0, errors.New("template repository not initialized")
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
	query := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplate{}).
		Where("tenant_uuid = ?", tenantUUID)
	if strings.TrimSpace(filters.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(filters.Status))
	}
	if strings.TrimSpace(filters.Keyword) != "" {
		like := "%" + strings.TrimSpace(filters.Keyword) + "%"
		query = query.Where("name ILIKE ?", like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []productcategory.CategoryTemplate
	if err := query.Order("updated_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *TemplateRepository) ListFields(ctx context.Context, tenantUUID string, templateID string) ([]productcategory.CategoryTemplateField, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("template repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	templateID = strings.TrimSpace(templateID)
	if tenantUUID == "" || templateID == "" {
		return nil, errors.New("tenant_uuid and template_id are required")
	}
	var records []productcategory.CategoryTemplateField
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplateField{}).
		Where("tenant_uuid = ? AND template_id = ?", tenantUUID, templateID).
		Order("sort_order ASC, field_key ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *TemplateRepository) ReplaceFields(ctx context.Context, tx *gorm.DB, tenantUUID string, templateID string, fields []productcategory.CategoryTemplateField) error {
	if r == nil {
		return errors.New("template repository not initialized")
	}
	if tx == nil {
		tx = r.DB
	}
	if tx == nil {
		return errors.New("template repository database is not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	templateID = strings.TrimSpace(templateID)
	if tenantUUID == "" || templateID == "" {
		return errors.New("tenant_uuid and template_id are required")
	}
	if err := tx.WithContext(ctx).
		Where("tenant_uuid = ? AND template_id = ?", tenantUUID, templateID).
		Delete(&productcategory.CategoryTemplateField{}).Error; err != nil {
		return err
	}
	if len(fields) == 0 {
		return nil
	}
	for idx := range fields {
		fields[idx].TenantUUID = tenantUUID
		fields[idx].TemplateID = templateID
	}
	return tx.WithContext(ctx).Create(&fields).Error
}

func (r *TemplateRepository) LatestPublishedVersion(ctx context.Context, tenantUUID string, templateID string) (*productcategory.CategoryTemplateVersion, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("template repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	templateID = strings.TrimSpace(templateID)
	if tenantUUID == "" || templateID == "" {
		return nil, errors.New("tenant_uuid and template_id are required")
	}
	var version productcategory.CategoryTemplateVersion
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplateVersion{}).
		Where("tenant_uuid = ? AND template_id = ? AND published_at IS NOT NULL", tenantUUID, templateID).
		Order("version_number DESC").
		First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *TemplateRepository) GetVersion(ctx context.Context, tenantUUID string, versionID string) (*productcategory.CategoryTemplateVersion, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("template repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	versionID = strings.TrimSpace(versionID)
	if tenantUUID == "" || versionID == "" {
		return nil, errors.New("tenant_uuid and version_id are required")
	}
	var version productcategory.CategoryTemplateVersion
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplateVersion{}).
		Where("tenant_uuid = ? AND id = ?", tenantUUID, versionID).
		First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *TemplateRepository) ListVersions(ctx context.Context, tenantUUID string, templateID string, limit int) ([]productcategory.CategoryTemplateVersion, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("template repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	templateID = strings.TrimSpace(templateID)
	if tenantUUID == "" || templateID == "" {
		return nil, errors.New("tenant_uuid and template_id are required")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var records []productcategory.CategoryTemplateVersion
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplateVersion{}).
		Where("tenant_uuid = ? AND template_id = ?", tenantUUID, templateID).
		Order("version_number DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *TemplateRepository) MaxVersionNumber(ctx context.Context, tenantUUID string, templateID string) (int, error) {
	if r == nil || r.DB == nil {
		return 0, errors.New("template repository not initialized")
	}
	tenantUUID = strings.TrimSpace(tenantUUID)
	templateID = strings.TrimSpace(templateID)
	if tenantUUID == "" || templateID == "" {
		return 0, errors.New("tenant_uuid and template_id are required")
	}
	var max int
	if err := r.DB.WithContext(ctx).
		Model(&productcategory.CategoryTemplateVersion{}).
		Where("tenant_uuid = ? AND template_id = ?", tenantUUID, templateID).
		Select("COALESCE(MAX(version_number), 0)").
		Scan(&max).Error; err != nil {
		return 0, err
	}
	return max, nil
}
