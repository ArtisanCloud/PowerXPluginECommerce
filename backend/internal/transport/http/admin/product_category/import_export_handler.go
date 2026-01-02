package product_category

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	categoryservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_category"
	"github.com/gin-gonic/gin"
)

type exportRequest struct {
	Kind       string `json:"kind"`
	CategoryID string `json:"categoryId"`
}

func (h *Handler) CategoriesImport(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	kind := strings.TrimSpace(strings.ToLower(c.Query("kind")))
	if kind == "" {
		kind = strings.TrimSpace(strings.ToLower(c.PostForm("kind")))
	}
	if kind == "" {
		kind = "mappings"
	}
	if kind != "mappings" {
		contracts.ResponseBadRequest(c, "unsupported import kind")
		return
	}
	categoryID := strings.TrimSpace(c.Query("categoryId"))
	if categoryID == "" {
		categoryID = strings.TrimSpace(c.PostForm("categoryId"))
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		contracts.ResponseBadRequest(c, "file is required")
		return
	}
	defer file.Close()
	result, err := h.service.ImportMappingsCSV(c.Request.Context(), file, categoryservice.MappingImportOptions{CategoryID: categoryID})
	if err != nil {
		respondMappingError(c, err)
		return
	}
	if result.Failed > 0 {
		contracts.ResponseErrorWithDetails(c, http.StatusUnprocessableEntity, contracts.ErrCodeValidationFailed, "导入存在失败行", result)
		return
	}
	contracts.ResponseSuccessWithMessage(c, result, "import completed")
}

func (h *Handler) CategoriesExport(c *gin.Context) {
	if h == nil || h.service == nil || !h.service.Ready() {
		contracts.ResponseServiceUnavailable(c, "category service unavailable", nil)
		return
	}
	var req exportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	kind := strings.TrimSpace(strings.ToLower(req.Kind))
	if kind == "" {
		kind = "mappings"
	}
	if kind != "mappings" {
		contracts.ResponseBadRequest(c, "unsupported export kind")
		return
	}
	body, filename, err := h.service.ExportMappingsCSV(c.Request.Context(), categoryservice.MappingExportOptions{
		CategoryID: req.CategoryID,
	})
	if err != nil {
		respondMappingError(c, err)
		return
	}
	if filename == "" {
		filename = "category_mappings.csv"
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	_, _ = c.Writer.Write(bytes.TrimPrefix(body, []byte("\xef\xbb\xbf")))
}
