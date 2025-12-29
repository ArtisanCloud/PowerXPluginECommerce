package product_sku

import (
	"net/http"
	"strconv"
	"time"

	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
)

// SerialHandler controls serial / batch level CRUD endpoints.
type SerialHandler struct {
	service *productskuservice.Service
}

func NewSerialHandler(service *productskuservice.Service) *SerialHandler {
	return &SerialHandler{service: service}
}

func (h *SerialHandler) List(c *gin.Context) {
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	skuID := c.Param("id")
	filters := productskuservice.SerialRecordFilters{
		Status: c.Query("status"),
		Batch:  c.Query("batch"),
	}
	if limit := c.Query("limit"); limit != "" {
		if parsed, err := parseInt(limit); err == nil {
			filters.Limit = parsed
		}
	}
	records, err := h.service.ListSerialRecords(c.Request.Context(), skuID, filters)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, records)
}

func (h *SerialHandler) Create(c *gin.Context) {
	if h.service == nil {
		respondError(c, ErrServiceUnavailable)
		return
	}
	skuID := c.Param("id")
	var payload struct {
		SerialNo  string     `json:"serial_no"`
		BatchNo   string     `json:"batch_no"`
		Status    string     `json:"status"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, err)
		return
	}
	input := productskuservice.SerialRecordInput{
		SerialNo:  payload.SerialNo,
		BatchNo:   payload.BatchNo,
		Status:    payload.Status,
		ExpiresAt: payload.ExpiresAt,
	}
	record, err := h.service.UpsertSerialRecord(c.Request.Context(), skuID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, record)
}

func parseInt(value string) (int, error) {
	return strconv.Atoi(value)
}
