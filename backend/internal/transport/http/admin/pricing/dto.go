package pricing

import "time"

type PricebookScopes struct {
	ChannelIDs       []string `json:"channel_ids,omitempty"`
	CustomerGroupIDs []string `json:"customer_group_ids,omitempty"`
	SupplierIDs      []string `json:"supplier_ids,omitempty"`
}

type PricebookDTO struct {
	ID               string           `json:"id"`
	Code             string           `json:"code"`
	Name             string           `json:"name"`
	Type             string           `json:"type"`
	Currency         string           `json:"currency"`
	Status           string           `json:"status"`
	Description      string           `json:"description,omitempty"`
	CurrentVersionID *string          `json:"current_version_id,omitempty"`
	Scopes           *PricebookScopes `json:"scopes,omitempty"`
	CreatedAt        *time.Time       `json:"created_at,omitempty"`
	UpdatedAt        *time.Time       `json:"updated_at,omitempty"`
}

type PageMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type PricebookListResponse struct {
	Items []PricebookDTO `json:"items"`
	Meta  PageMeta       `json:"meta"`
}

type PricebookCreateRequest struct {
	Code        string           `json:"code" binding:"required"`
	Name        string           `json:"name" binding:"required"`
	Type        string           `json:"type,omitempty"`
	Currency    string           `json:"currency" binding:"required"`
	Description string           `json:"description,omitempty"`
	Scopes      *PricebookScopes `json:"scopes,omitempty"`
}

type PricebookUpdateRequest struct {
	Name        *string          `json:"name,omitempty"`
	Description *string          `json:"description,omitempty"`
	Status      *string          `json:"status,omitempty"`
	Scopes      *PricebookScopes `json:"scopes,omitempty"`
}

type VersionDTO struct {
	ID          string     `json:"id"`
	PricebookID string     `json:"pricebook_id"`
	Version     int        `json:"version"`
	State       string     `json:"state"`
	EffectiveAt time.Time  `json:"effective_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	PublishedBy *string    `json:"published_by,omitempty"`
	Note        *string    `json:"note,omitempty"`
}

type VersionCreateRequest struct {
	CopyFromVersionID *string `json:"copy_from_version_id,omitempty"`
}

type VersionPublishRequest struct {
	EffectiveAt *time.Time `json:"effective_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Note        *string    `json:"note,omitempty"`
}

type VersionArchiveRequest struct {
	Note *string `json:"note,omitempty"`
}

type PricebookItemInput struct {
	SKUID           string         `json:"sku_id" binding:"required"`
	BaseAmountMinor *int64         `json:"base_amount_minor,omitempty"`
	SaleAmountMinor *int64         `json:"sale_amount_minor,omitempty"`
	MsrpAmountMinor *int64         `json:"msrp_amount_minor,omitempty"`
	CostAmountMinor *int64         `json:"cost_amount_minor,omitempty"`
	MinAmountMinor  *int64         `json:"min_amount_minor,omitempty"`
	MaxAmountMinor  *int64         `json:"max_amount_minor,omitempty"`
	TaxIncluded     *bool          `json:"tax_included,omitempty"`
	Meta            map[string]any `json:"meta,omitempty"`
}

type ItemsUpsertRequest struct {
	Items []PricebookItemInput `json:"items" binding:"required"`
}

type ItemsUpsertResponse struct {
	Upserted int      `json:"upserted"`
	Skipped  []string `json:"skipped,omitempty"`
}
