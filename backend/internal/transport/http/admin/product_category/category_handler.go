package product_category

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	productcategory "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/entity/models/product_category"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) Tree(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	nodes, err := h.service.AdminTree(c.Request.Context())
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"items": nodes})
}

func (h *Handler) List(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	parentID := strings.TrimSpace(c.Query("parentId"))
	var parent *string
	if c.Query("parentId") != "" {
		parent = &parentID
	}
	result, err := h.service.ListCategories(c.Request.Context(), categoryservice.CategoryListFilters{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Status:   strings.TrimSpace(c.Query("status")),
		ParentID: parent,
		Page:     toInt(c.Query("page"), 1),
		PageSize: toInt(c.Query("pageSize"), 50),
	})
	if err != nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, err.Error())
		return
	}
	items := make([]categoryservice.CategoryNode, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toCategoryNode(item))
	}
	contracts.ResponseSuccess(c, gin.H{
		"items": items,
		"meta":  gin.H{"total": result.Total, "page": result.Page, "pageSize": result.PageSize},
	})
}

func (h *Handler) Create(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	created, err := h.service.CreateCategory(c.Request.Context(), req)
	if err != nil {
		respondCategoryError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, toCategoryNode(*created), "category created")
}

func (h *Handler) Update(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	updated, err := h.service.UpdateCategory(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondCategoryError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, toCategoryNode(*updated), "category updated")
}

func (h *Handler) Move(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.CategoryMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	updated, err := h.service.MoveCategory(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondCategoryError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, toCategoryNode(*updated), "category moved")
}

func (h *Handler) SetStatus(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req categoryservice.CategoryStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	updated, err := h.service.SetCategoryStatus(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		respondCategoryError(c, err)
		return
	}
	contracts.ResponseSuccessWithMessage(c, toCategoryNode(*updated), "category status updated")
}

func (h *Handler) Delete(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "id is required")
		return
	}
	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		if err == gorm.ErrRecordNotFound {
			contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "category not found")
			return
		}
		// 删除保护（存在子类目/商品引用）视为冲突
		contracts.ResponseError(c, http.StatusConflict, contracts.ErrCodeConflict, err.Error())
		return
	}
	contracts.ResponseSuccess(c, gin.H{"ok": true})
}

func respondCategoryError(c *gin.Context, err error) {
	if err == nil {
		contracts.ResponseError(c, http.StatusInternalServerError, contracts.ErrCodeInternalError, "unknown error")
		return
	}
	if vErrs, ok := err.(categoryservice.ValidationErrors); ok {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "参数校验失败", vErrs)
		return
	}
	if cErr, ok := err.(*categoryservice.ConflictError); ok {
		contracts.ResponseErrorWithDetails(c, http.StatusConflict, contracts.ErrCodeConflict, cErr.Message, cErr)
		return
	}
	if err == gorm.ErrRecordNotFound {
		contracts.ResponseError(c, http.StatusNotFound, contracts.ErrCodeNotFound, "category not found")
		return
	}
	contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
}

func toInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

func toCategoryNode(category productcategory.ProductCategory) categoryservice.CategoryNode {
	parentID := category.ParentID
	return categoryservice.CategoryNode{
		ID:          category.ID,
		ParentID:    parentID,
		Code:        category.Code,
		DisplayName: category.DisplayName,
		AliasSlug:   category.AliasSlug,
		Path:        category.Path,
		Level:       category.Level,
		SortOrder:   category.SortOrder,
		Status:      category.Status,
		ImageURL:    category.ImageURL,
		IsFeatured:  category.IsFeatured,
	}
}
