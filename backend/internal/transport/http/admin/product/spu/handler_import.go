package spu

import (
	"io"
	"net/http"

	"github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/contracts"
	spuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product/spu"
	"github.com/gin-gonic/gin"
)

// ImportExportHandler exposes bulk import/export endpoints.
type ImportExportHandler struct {
	service *spuservice.ImportService
}

// NewImportExportHandler constructs an instance guarded for nil services.
func NewImportExportHandler(service *spuservice.ImportService) *ImportExportHandler {
	return &ImportExportHandler{service: service}
}

// Import consumes multipart payloads, kicking off asynchronous import jobs.
func (h *ImportExportHandler) Import(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "import service unavailable")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, "缺少 file 字段")
		return
	}
	src, err := file.Open()
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	defer src.Close()
	content, err := io.ReadAll(src)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	templateID := c.PostForm("templateId")
	req := spuservice.ImportRequest{
		TemplateID: templateID,
		Filename:   file.Filename,
		Payload:    content,
	}
	taskID, err := h.service.StartImport(c.Request.Context(), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"taskId": taskID}, "导入已排队")
}

// Export schedules an export job based on provided filters/fields.
func (h *ImportExportHandler) Export(c *gin.Context) {
	if h.service == nil {
		contracts.ResponseError(c, http.StatusServiceUnavailable, contracts.ErrCodeInternalError, "import service unavailable")
		return
	}
	var req spuservice.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	taskID, err := h.service.StartExport(c.Request.Context(), req)
	if err != nil {
		contracts.ResponseError(c, http.StatusBadRequest, contracts.ErrCodeInvalidRequest, err.Error())
		return
	}
	contracts.ResponseSuccessWithMessage(c, gin.H{"taskId": taskID}, "导出已排队")
}
