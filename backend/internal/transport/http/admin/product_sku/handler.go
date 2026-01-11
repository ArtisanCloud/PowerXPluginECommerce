package product_sku

import (
	"errors"
	"net/http"
	"strings"

	productskuservice "github.com/ArtisanCloud/PowerXPlugin/plugins/com-powerx-plugin-ecommerce/backend/internal/services/admin/product_sku"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func respondStub(c *gin.Context, message string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": message,
	})
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		err = errors.New("unknown error")
	}
	c.JSON(classifyError(err), gin.H{
		"message": err.Error(),
	})
}

func classifyError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "forbidden"),
		strings.Contains(msg, "permission denied"),
		strings.Contains(msg, "permission"):
		return http.StatusForbidden
	case strings.Contains(msg, "conflict"):
		return http.StatusConflict
	case strings.Contains(msg, "required"),
		strings.Contains(msg, "invalid"),
		strings.Contains(msg, "negative"),
		strings.Contains(msg, "too many"),
		strings.Contains(msg, "not found"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// Handler wires basic CRUD endpoints for SKU resources.
type Handler struct {
	service *productskuservice.Service
}

func NewHandler(service *productskuservice.Service) *Handler {
	return &Handler{service: service}
}

type skuListQuery struct {
	SPUID    string `form:"spuId"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Q        string `form:"q"`
	Locale   string `form:"locale"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

func (h *Handler) Get(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	locale := strings.TrimSpace(c.Query("locale"))
	item, err := h.service.GetSku(c.Request.Context(), c.Param("id"), locale)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) List(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	var req skuListQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		respondError(c, err)
		return
	}
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		keyword = strings.TrimSpace(req.Q)
	}
	result, err := h.service.ListSkus(c.Request.Context(), productskuservice.SkuListQuery{
		SPUID:    req.SPUID,
		Status:   req.Status,
		Keyword:  keyword,
		Locale:   req.Locale,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) Upsert(c *gin.Context) {
	if h.service == nil {
		respondError(c, errors.New("product SKU service unavailable"))
		return
	}
	var req productskuservice.SkuUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, err)
		return
	}
	result, err := h.service.UpsertSkus(c.Request.Context(), req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) Update(c *gin.Context) {
	respondStub(c, "SKU update not implemented")
}

func (h *Handler) Delete(c *gin.Context) {
	respondStub(c, "SKU delete not implemented")
}
