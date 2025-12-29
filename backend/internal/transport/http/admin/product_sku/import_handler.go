package product_sku

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

const maxImportUpload = 2 << 20 // 2MB

// ImportExportHandler wires SKU import/export endpoints.
type ImportExportHandler struct {
	service *productskuservice.Service
}

func NewImportExportHandler(service *productskuservice.Service) *ImportExportHandler {
	return &ImportExportHandler{service: service}
}

func (h *ImportExportHandler) Import(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		respondError(c, err)
		return
	}
	src, err := file.Open()
	if err != nil {
		respondError(c, err)
		return
	}
	defer src.Close()
	buffer := bytes.NewBuffer(nil)
	if _, err := io.CopyN(buffer, src, maxImportUpload+1); err != nil && err != io.EOF {
		respondError(c, err)
		return
	}
	if buffer.Len() > maxImportUpload {
		respondError(c, fmt.Errorf("import payload exceeds %d bytes", maxImportUpload))
		return
	}
	mode := c.DefaultPostForm("mode", "upsert")
	resp, err := h.service.ImportSkus(c.Request.Context(), productskuservice.SkuImportRequest{
		FileName: filepath.Base(file.Filename),
		Mode:     mode,
		Content:  buffer.Bytes(),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, resp)
}

func (h *ImportExportHandler) Export(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	var payload struct {
		Format  string            `json:"format"`
		Filters map[string]string `json:"filters"`
		Limit   int               `json:"limit"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil && err != io.EOF {
		respondError(c, err)
		return
	}
	if payload.Filters == nil {
		payload.Filters = make(map[string]string)
	}
	resp, err := h.service.ExportSkus(c.Request.Context(), productskuservice.SkuExportRequest{
		Format:  payload.Format,
		Filters: payload.Filters,
		Limit:   payload.Limit,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, resp)
}
