package product

import "time"

type miniAppProductListQuery struct {
	Keyword  string `form:"keyword"`
	Q        string `form:"q"`
	Status   string `form:"status"`
	Type     string `form:"type"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type miniAppSkuListQuery struct {
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type productListResponse struct {
	Items    []miniAppProductSummary `json:"items"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
	Total    int64                   `json:"total"`
}

type miniAppProductSummary struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Type      string     `json:"type,omitempty"`
	Status    string     `json:"status,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type skuListResponse struct {
	Items    []miniAppSkuSummary `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"pageSize"`
	Total    int64               `json:"total"`
}

type miniAppSkuSummary struct {
	ID        string     `json:"id"`
	SPUID     string     `json:"spuId"`
	Code      string     `json:"code"`
	Status    string     `json:"status"`
	Barcode   string     `json:"barcode,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}
