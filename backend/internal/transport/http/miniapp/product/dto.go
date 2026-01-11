package product

import "time"

type miniAppProductListQuery struct {
	Keyword            string `form:"keyword"`
	Q                  string `form:"q"`
	Tags               string `form:"tags"`
	Tag                string `form:"tag"`
	Status             string `form:"status"`
	Type               string `form:"type"`
	CategoryID         string `form:"categoryId"`
	CategoryPathPrefix string `form:"categoryPathPrefix"`
	// Sort: comprehensive|updatedAt|sales|price; Order: asc|desc.
	Sort     string   `form:"sort"`
	Order    string   `form:"order"`
	MinPrice *float64 `form:"minPrice"`
	MaxPrice *float64 `form:"maxPrice"`
	InStock  *bool    `form:"inStock"`
	HasPlans *bool    `form:"hasPlans"`
	Page     int      `form:"page"`
	PageSize int      `form:"pageSize"`
}

type miniAppProductTagListQuery struct {
	CategoryID         string `form:"categoryId"`
	CategoryPathPrefix string `form:"categoryPathPrefix"`
	Limit              int    `form:"limit"`
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

type productTagListResponse struct {
	Items []productTagItem `json:"items"`
}

type productTagItem struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

type miniAppProductSummary struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Type       string     `json:"type,omitempty"`
	Status     string     `json:"status,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
	CoverURL   string     `json:"coverUrl,omitempty"`
	MinPrice   *float64   `json:"minPrice,omitempty"`
	MaxPrice   *float64   `json:"maxPrice,omitempty"`
	Currency   string     `json:"currency,omitempty"`
	PriceLabel string     `json:"priceLabel,omitempty"`
	SKUCount   int        `json:"skuCount,omitempty"`
}

type miniAppProductDetail struct {
	ID           string   `json:"id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Type         string   `json:"type,omitempty"`
	Status       string   `json:"status,omitempty"`
	CategoryID   string   `json:"categoryId,omitempty"`
	CategoryPath string   `json:"categoryPath,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CoverURL     string   `json:"coverUrl,omitempty"`
	Subtitle     string   `json:"subtitle,omitempty"`
	Description  string   `json:"description,omitempty"`
	MinPrice     *float64 `json:"minPrice,omitempty"`
	MaxPrice     *float64 `json:"maxPrice,omitempty"`
	Currency     string   `json:"currency,omitempty"`
	PriceLabel   string   `json:"priceLabel,omitempty"`
	SKUCount     int      `json:"skuCount,omitempty"`
}

type miniAppProductDetailBundle struct {
	SPU  miniAppProductDetail     `json:"spu"`
	Spec miniAppProductSpecBundle `json:"spec"`
	SKUs []miniAppSkuDetail       `json:"skus"`
}

type miniAppProductSpecBundle struct {
	Groups []miniAppSpecGroup `json:"groups"`
}

type miniAppSpecGroup struct {
	ID        string              `json:"id"`
	Code      string              `json:"code"`
	Name      string              `json:"name"`
	Required  bool                `json:"required"`
	SortOrder int                 `json:"sortOrder"`
	Options   []miniAppSpecOption `json:"options"`
}

type miniAppSpecOption struct {
	ID        string         `json:"id"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	SortOrder int            `json:"sortOrder"`
	Meta      map[string]any `json:"meta,omitempty"`
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
	ImageURL  string     `json:"imageUrl,omitempty"`
	Price     *float64   `json:"price,omitempty"`
	Currency  string     `json:"currency,omitempty"`
	StockQty  *int       `json:"stockQty,omitempty"`
}

type miniAppSkuDetail struct {
	ID            string            `json:"id"`
	Code          string            `json:"code"`
	Price         *float64          `json:"price,omitempty"`
	Currency      string            `json:"currency,omitempty"`
	ImageURL      string            `json:"imageUrl,omitempty"`
	StockQty      *int              `json:"stockQty,omitempty"`
	Spec          map[string]string `json:"spec,omitempty"`
	SpecSignature string            `json:"specSignature,omitempty"`
}

type subscriptionPlanListResponse struct {
	Items []miniAppSubscriptionPlan `json:"items"`
}

type miniAppSubscriptionPlan struct {
	ID           string  `json:"id"`
	PlanCode     string  `json:"planCode"`
	Name         string  `json:"name"`
	BillingCycle string  `json:"billingCycle"`
	BillingValue int     `json:"billingValue,omitempty"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	TrialDays    int     `json:"trialDays,omitempty"`
	AutoRenew    bool    `json:"autoRenew,omitempty"`
	CancelPolicy string  `json:"cancelPolicy,omitempty"`
	Status       string  `json:"status,omitempty"`
}
