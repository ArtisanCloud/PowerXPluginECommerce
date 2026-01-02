package product_category

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	productmodel "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product"
	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	catrepo "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/repository/product_category"
	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/pkg/utils"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string { return "validation failed" }

func (e ValidationErrors) add(field, message string) ValidationErrors {
	return append(e, ValidationError{Field: field, Message: message})
}

func (e ValidationErrors) empty() bool { return len(e) == 0 }

type ConflictError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

func (e *ConflictError) Error() string {
	if e == nil {
		return "conflict"
	}
	return e.Message
}

type CategoryCreateRequest struct {
	ParentID       *string `json:"parentId"`
	Code           string  `json:"code"`
	DisplayName    string  `json:"displayName"`
	AliasSlug      string  `json:"aliasSlug"`
	SortOrder      int     `json:"sortOrder"`
	Status         string  `json:"status"`
	TemplateID     *string `json:"templateId"`
	SEOTitle       string  `json:"seoTitle"`
	SEODescription string  `json:"seoDescription"`
	SEOKeywords    string  `json:"seoKeywords"`
	ImageURL       string  `json:"imageUrl"`
	IsFeatured     bool    `json:"isFeatured"`
}

type CategoryUpdateRequest struct {
	Code           *string `json:"code"`
	DisplayName    *string `json:"displayName"`
	AliasSlug      *string `json:"aliasSlug"`
	SortOrder      *int    `json:"sortOrder"`
	TemplateID     *string `json:"templateId"`
	SEOTitle       *string `json:"seoTitle"`
	SEODescription *string `json:"seoDescription"`
	SEOKeywords    *string `json:"seoKeywords"`
	ImageURL       *string `json:"imageUrl"`
	IsFeatured     *bool   `json:"isFeatured"`
}

type CategoryMoveRequest struct {
	ParentID  *string `json:"parentId"`
	SortOrder *int    `json:"sortOrder"`
}

type CategoryStatusRequest struct {
	Status string `json:"status"`
}

type CategoryListFilters struct {
	Keyword  string
	Status   string
	ParentID *string
	Page     int
	PageSize int
}

type CategoryListResult struct {
	Items    []productcategory.ProductCategory `json:"items"`
	Total    int64                             `json:"total"`
	Page     int                               `json:"page"`
	PageSize int                               `json:"pageSize"`
}

type CategoryNode struct {
	ID          string         `json:"id"`
	ParentID    *string        `json:"parentId,omitempty"`
	Code        string         `json:"code"`
	DisplayName string         `json:"displayName"`
	AliasSlug   string         `json:"aliasSlug"`
	Path        string         `json:"path"`
	Level       int            `json:"level"`
	SortOrder   int            `json:"sortOrder"`
	Status      string         `json:"status"`
	ImageURL    string         `json:"imageUrl,omitempty"`
	IsFeatured  bool           `json:"isFeatured"`
	Children    []CategoryNode `json:"children,omitempty"`
}

func (s *Service) CreateCategory(ctx context.Context, req CategoryCreateRequest) (*productcategory.ProductCategory, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Code = strings.TrimSpace(req.Code)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.AliasSlug = strings.TrimSpace(req.AliasSlug)
	req.Status = strings.TrimSpace(strings.ToLower(req.Status))
	if req.Status == "" {
		req.Status = productcategory.CategoryStatusEnabled
	}
	var vErrs ValidationErrors
	if req.Code == "" {
		vErrs = vErrs.add("code", "code 为必填")
	}
	if req.DisplayName == "" {
		vErrs = vErrs.add("displayName", "displayName 为必填")
	}
	if req.AliasSlug == "" {
		vErrs = vErrs.add("aliasSlug", "aliasSlug 为必填")
	}
	if req.Status != productcategory.CategoryStatusEnabled && req.Status != productcategory.CategoryStatusDisabled {
		vErrs = vErrs.add("status", "status 必须为 enabled/disabled")
	}
	if !vErrs.empty() {
		return nil, vErrs
	}

	now := time.Now().UTC()
	result := &productcategory.ProductCategory{}
	err = s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewCategoryRepository(tx)
		var parent *productcategory.ProductCategory
		if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
			p, err := repoTx.GetByID(ctx, tenantID, strings.TrimSpace(*req.ParentID), true)
			if err != nil {
				return err
			}
			parent = p
		}
		id := utils.NewUUID()
		path := buildCategoryPath(parent, id)
		level := 0
		if parent != nil {
			level = parent.Level + 1
		}
		category := &productcategory.ProductCategory{
			ID:             id,
			TenantUUID:     tenantID,
			ParentID:       normalizeOptionalUUID(req.ParentID),
			Code:           req.Code,
			DisplayName:    req.DisplayName,
			AliasSlug:      req.AliasSlug,
			Path:           path,
			Level:          level,
			SortOrder:      req.SortOrder,
			Status:         req.Status,
			TemplateID:     normalizeOptionalUUID(req.TemplateID),
			SEOTitle:       strings.TrimSpace(req.SEOTitle),
			SEODescription: strings.TrimSpace(req.SEODescription),
			SEOKeywords:    strings.TrimSpace(req.SEOKeywords),
			ImageURL:       strings.TrimSpace(req.ImageURL),
			IsFeatured:     req.IsFeatured,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tx.Create(category).Error; err != nil {
			if conflict := toCategoryConflictError(err); conflict != nil {
				return conflict
			}
			return err
		}
		*result = *category
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) UpdateCategory(ctx context.Context, id string, req CategoryUpdateRequest) (*productcategory.ProductCategory, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}

	updates := map[string]any{}
	var vErrs ValidationErrors
	if req.Code != nil {
		code := strings.TrimSpace(*req.Code)
		if code == "" {
			vErrs = vErrs.add("code", "code 不能为空")
		} else {
			updates["code"] = code
		}
	}
	if req.DisplayName != nil {
		name := strings.TrimSpace(*req.DisplayName)
		if name == "" {
			vErrs = vErrs.add("displayName", "displayName 不能为空")
		} else {
			updates["display_name"] = name
		}
	}
	if req.AliasSlug != nil {
		slug := strings.TrimSpace(*req.AliasSlug)
		if slug == "" {
			vErrs = vErrs.add("aliasSlug", "aliasSlug 不能为空")
		} else {
			updates["alias_slug"] = slug
		}
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.TemplateID != nil {
		updates["template_id"] = normalizeOptionalUUID(req.TemplateID)
	}
	if req.SEOTitle != nil {
		updates["seo_title"] = strings.TrimSpace(*req.SEOTitle)
	}
	if req.SEODescription != nil {
		updates["seo_description"] = strings.TrimSpace(*req.SEODescription)
	}
	if req.SEOKeywords != nil {
		updates["seo_keywords"] = strings.TrimSpace(*req.SEOKeywords)
	}
	if req.ImageURL != nil {
		updates["image_url"] = strings.TrimSpace(*req.ImageURL)
	}
	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}
	if !vErrs.empty() {
		return nil, vErrs
	}
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}
	updates["updated_at"] = time.Now().UTC()

	var updated *productcategory.ProductCategory
	err = s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewCategoryRepository(tx)
		if _, err := repoTx.GetByID(ctx, tenantID, id, true); err != nil {
			return err
		}
		if err := tx.Model(&productcategory.ProductCategory{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Updates(updates).Error; err != nil {
			if conflict := toCategoryConflictError(err); conflict != nil {
				return conflict
			}
			return err
		}
		record, err := repoTx.GetByID(ctx, tenantID, id, false)
		if err != nil {
			return err
		}
		updated = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) ListCategories(ctx context.Context, filters CategoryListFilters) (*CategoryListResult, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	items, total, err := s.CategoryRepo.List(ctx, tenantID, catrepo.CategoryListFilters{
		Keyword:  filters.Keyword,
		Status:   filters.Status,
		ParentID: filters.ParentID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	return &CategoryListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) AdminTree(ctx context.Context) ([]CategoryNode, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	records, err := s.CategoryRepo.ListAllForTree(ctx, tenantID, true)
	if err != nil {
		return nil, err
	}
	return buildTree(records, false), nil
}

func (s *Service) MiniAppTree(ctx context.Context) ([]CategoryNode, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	records, err := s.CategoryRepo.ListAllForTree(ctx, tenantID, true)
	if err != nil {
		return nil, err
	}
	return buildTree(records, true), nil
}

func (s *Service) MoveCategory(ctx context.Context, id string, req CategoryMoveRequest) (*productcategory.ProductCategory, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("id is required")
	}
	var updated *productcategory.ProductCategory
	err = s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewCategoryRepository(tx)
		node, err := repoTx.GetByID(ctx, tenantID, id, true)
		if err != nil {
			return err
		}
		var parent *productcategory.ProductCategory
		nextParentID := normalizeOptionalUUID(req.ParentID)
		if nextParentID != nil && strings.TrimSpace(*nextParentID) != "" {
			if strings.TrimSpace(*nextParentID) == node.ID {
				return ValidationErrors{}.add("parentId", "parentId 不能为自身")
			}
			p, err := repoTx.GetByID(ctx, tenantID, strings.TrimSpace(*nextParentID), true)
			if err != nil {
				return err
			}
			if strings.HasPrefix(p.Path, node.Path) {
				return ValidationErrors{}.add("parentId", "禁止迁移到子孙节点（会形成环）")
			}
			parent = p
		}
		newPath := buildCategoryPath(parent, node.ID)
		newLevel := 0
		if parent != nil {
			newLevel = parent.Level + 1
		}
		levelDelta := newLevel - node.Level
		if strings.TrimSpace(newPath) != strings.TrimSpace(node.Path) || levelDelta != 0 {
			if err := repoTx.UpdateSubtreePathAndLevel(ctx, tx, tenantID, node.Path, newPath, levelDelta); err != nil {
				return err
			}
		}

		updates := map[string]any{
			"parent_id":  nextParentID,
			"path":       newPath,
			"level":      newLevel,
			"updated_at": time.Now().UTC(),
		}
		if req.SortOrder != nil {
			updates["sort_order"] = *req.SortOrder
		}
		if err := tx.Model(&productcategory.ProductCategory{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Updates(updates).Error; err != nil {
			return err
		}
		record, err := repoTx.GetByID(ctx, tenantID, id, false)
		if err != nil {
			return err
		}
		updated = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) SetCategoryStatus(ctx context.Context, id string, req CategoryStatusRequest) (*productcategory.ProductCategory, error) {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return nil, errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return nil, err
	}
	id = strings.TrimSpace(id)
	req.Status = strings.TrimSpace(strings.ToLower(req.Status))
	if id == "" {
		return nil, errors.New("id is required")
	}
	if req.Status != productcategory.CategoryStatusEnabled && req.Status != productcategory.CategoryStatusDisabled {
		return nil, ValidationErrors{}.add("status", "status 必须为 enabled/disabled")
	}

	var updated *productcategory.ProductCategory
	err = s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewCategoryRepository(tx)
		if _, err := repoTx.GetByID(ctx, tenantID, id, true); err != nil {
			return err
		}
		if err := tx.Model(&productcategory.ProductCategory{}).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Updates(map[string]any{"status": req.Status, "updated_at": time.Now().UTC()}).Error; err != nil {
			return err
		}
		record, err := repoTx.GetByID(ctx, tenantID, id, false)
		if err != nil {
			return err
		}
		updated = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) DeleteGuard(ctx context.Context, id string) error {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	var guardErr error
	err = s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		repoTx := catrepo.NewCategoryRepository(tx)
		node, err := repoTx.GetByID(ctx, tenantID, id, true)
		if err != nil {
			return err
		}
		children, err := repoTx.CountChildren(ctx, tenantID, node.ID)
		if err != nil {
			return err
		}
		if children > 0 {
			guardErr = fmt.Errorf("类目存在子类目（%d），禁止删除；请先迁移或停用", children)
			return nil
		}
		var productCount int64
		if err := tx.WithContext(ctx).Model(&productmodel.SPU{}).
			Where("tenant_uuid = ? AND (category_id = ? OR category_path LIKE ?)", tenantID, node.ID, node.Path+"%").
			Count(&productCount).Error; err != nil {
			return err
		}
		if productCount > 0 {
			guardErr = fmt.Errorf("类目已被商品引用（%d），禁止删除；请先迁移商品或停用类目", productCount)
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}
	return guardErr
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	if s == nil || !s.Ready() || s.CategoryRepo == nil {
		return errors.New("category service not initialized")
	}
	tenantID, err := s.tenantFromContext(ctx)
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required")
	}
	if err := s.DeleteGuard(ctx, id); err != nil {
		return err
	}
	return s.CategoryRepo.WithTenantTx(ctx, tenantID, func(tx *gorm.DB) error {
		// 确认存在
		repoTx := catrepo.NewCategoryRepository(tx)
		if _, err := repoTx.GetByID(ctx, tenantID, id, false); err != nil {
			return err
		}
		// 软删除关联数据，避免遗留脏数据（若未来加 FK 也能兼容）
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND category_id = ?", tenantID, id).
			Delete(&productcategory.CategoryMapping{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND category_id = ?", tenantID, id).
			Delete(&productcategory.CategoryLocale{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("tenant_uuid = ? AND category_id = ?", tenantID, id).
			Delete(&productcategory.CategoryPermission{}).Error; err != nil {
			return err
		}
		return tx.WithContext(ctx).
			Where("tenant_uuid = ? AND id = ?", tenantID, id).
			Delete(&productcategory.ProductCategory{}).Error
	})
}

func normalizeOptionalUUID(input *string) *string {
	if input == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func buildCategoryPath(parent *productcategory.ProductCategory, id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if parent == nil || strings.TrimSpace(parent.Path) == "" {
		return fmt.Sprintf("/%s/", id)
	}
	return parent.Path + id + "/"
}

func toCategoryConflictError(err error) *ConflictError {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch strings.TrimSpace(pgErr.ConstraintName) {
		case "uk_product_categories_tenant_code":
			return &ConflictError{Field: "code", Message: "类目编码已存在"}
		case "uk_product_categories_tenant_alias_slug":
			return &ConflictError{Field: "aliasSlug", Message: "alias/slug 已存在"}
		case "uk_product_categories_tenant_parent_display_name":
			return &ConflictError{Field: "displayName", Message: "同一父类目下 displayName 已存在"}
		default:
			return &ConflictError{Message: "数据冲突（可能违反唯一性约束）"}
		}
	}
	return nil
}

func buildTree(records []productcategory.ProductCategory, hideDisabledSubtree bool) []CategoryNode {
	type rawNode struct {
		self     CategoryNode
		enabled  bool
		parentID *string
	}
	nodes := make(map[string]*rawNode, len(records))
	children := make(map[string][]*rawNode)
	for _, record := range records {
		rec := record
		rec.Normalize()
		n := &rawNode{
			self: CategoryNode{
				ID:          rec.ID,
				ParentID:    rec.ParentID,
				Code:        rec.Code,
				DisplayName: rec.DisplayName,
				AliasSlug:   rec.AliasSlug,
				Path:        rec.Path,
				Level:       rec.Level,
				SortOrder:   rec.SortOrder,
				Status:      rec.Status,
				ImageURL:    rec.ImageURL,
				IsFeatured:  rec.IsFeatured,
			},
			enabled:  rec.IsEnabled(),
			parentID: rec.ParentID,
		}
		nodes[rec.ID] = n
		parentKey := ""
		if rec.ParentID != nil {
			parentKey = strings.TrimSpace(*rec.ParentID)
		}
		children[parentKey] = append(children[parentKey], n)
	}

	// 稳定排序：同级按 sort_order，再按 display_name
	for _, siblings := range children {
		sort.SliceStable(siblings, func(i, j int) bool {
			if siblings[i].self.SortOrder != siblings[j].self.SortOrder {
				return siblings[i].self.SortOrder < siblings[j].self.SortOrder
			}
			return siblings[i].self.DisplayName < siblings[j].self.DisplayName
		})
	}

	var walk func(parentKey string, parentVisible bool) []CategoryNode
	walk = func(parentKey string, parentVisible bool) []CategoryNode {
		siblings := children[parentKey]
		if len(siblings) == 0 {
			return nil
		}
		out := make([]CategoryNode, 0, len(siblings))
		for _, child := range siblings {
			visible := parentVisible
			if hideDisabledSubtree && (!child.enabled || !parentVisible) {
				visible = false
			}
			if !visible {
				continue
			}
			next := child.self
			next.Children = walk(child.self.ID, true)
			out = append(out, next)
		}
		return out
	}

	// roots：parent_id 为空或父节点缺失的都作为 root（避免脏数据导致全部不可见）
	roots := make([]*rawNode, 0)
	for _, n := range nodes {
		if n.parentID == nil {
			roots = append(roots, n)
			continue
		}
		if _, ok := nodes[strings.TrimSpace(*n.parentID)]; !ok {
			roots = append(roots, n)
		}
	}
	sort.SliceStable(roots, func(i, j int) bool {
		if roots[i].self.SortOrder != roots[j].self.SortOrder {
			return roots[i].self.SortOrder < roots[j].self.SortOrder
		}
		return roots[i].self.DisplayName < roots[j].self.DisplayName
	})
	out := make([]CategoryNode, 0, len(roots))
	for _, root := range roots {
		if hideDisabledSubtree && !root.enabled {
			continue
		}
		n := root.self
		n.Children = walk(n.ID, true)
		out = append(out, n)
	}
	return out
}
