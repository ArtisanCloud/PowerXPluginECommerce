package product_sku

import (
	"encoding/json"
	"time"
)

// SkuSpec describes a single spec/value pairing for SKU identification.
type SkuSpec struct {
	SpecID    string `json:"spec_id"`
	SpecName  string `json:"spec_name,omitempty"`
	ValueID   string `json:"value_id"`
	ValueName string `json:"value_name,omitempty"`
}

// SpecValueSelection captures metadata for a selectable spec value.
type SpecValueSelection struct {
	ValueID   string `json:"value_id"`
	ValueName string `json:"value_name,omitempty"`
	ValueCode string `json:"value_code,omitempty"`
}

// SpecSelection represents a specification and the values selected by the client.
type SpecSelection struct {
	SpecID   string               `json:"spec_id"`
	SpecName string               `json:"spec_name,omitempty"`
	ValueIDs []string             `json:"value_ids"`
	Values   []SpecValueSelection `json:"values,omitempty"`
}

// SkuGeneratorDefaults carries the default fields applied to generated SKUs.
type SkuGeneratorDefaults struct {
	BarcodePrefix string  `json:"barcode_prefix,omitempty"`
	MinOrderQty   int     `json:"min_order_qty,omitempty"`
	CostPrice     float64 `json:"cost_price,omitempty"`
	Weight        float64 `json:"weight,omitempty"`
	Dimensions    string  `json:"dimensions,omitempty"`
}

// SkuGeneratorRequest is the payload for `/spus/{id}/skus/generate`.
type SkuGeneratorRequest struct {
	SpecSelections []SpecSelection      `json:"specSelections"`
	Defaults       SkuGeneratorDefaults `json:"defaults"`
}

// SkuGeneratorCandidate represents a single cartesian combination preview.
type SkuGeneratorCandidate struct {
	Specs           []SkuSpec            `json:"specs"`
	PreviewSKUCode  string               `json:"preview_sku_code"`
	DefaultValues   SkuGeneratorDefaults `json:"default_values"`
	Exists          bool                 `json:"exists"`
	Selected        bool                 `json:"selected"`
	ConflictReasons []string             `json:"conflict_reasons,omitempty"`
}

// SkuGeneratorResponse returns the preview list.
type SkuGeneratorResponse struct {
	Candidates []SkuGeneratorCandidate `json:"candidates"`
}

// SkuListQuery describes filters for GET /admin/product/skus.
type SkuListQuery struct {
	SPUID    string `json:"spu_id,omitempty"`
	Status   string `json:"status,omitempty"`
	Keyword  string `json:"keyword,omitempty"`
	Locale   string `json:"locale,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// SkuListItem summarizes SKU details for list API.
type SkuListItem struct {
	ID          string     `json:"id"`
	SPUID       string     `json:"spu_id"`
	SPUName     string     `json:"spu_name,omitempty"`
	SKUCode     string     `json:"sku_code"`
	Status      string     `json:"status"`
	Barcode     string     `json:"barcode,omitempty"`
	Specs       []SkuSpec  `json:"specs,omitempty"`
	SpecDisplay string     `json:"spec_display,omitempty"`
	SalePrice   *float64   `json:"sale_price,omitempty"`
	Currency    string     `json:"currency,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// SkuListResult is the paginated response payload.
type SkuListResult struct {
	Items    []SkuListItem `json:"items"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Total    int64         `json:"total"`
}

// SkuUpsertPayload defines a SKU to create or clone.
type SkuUpsertPayload struct {
	SPUID         string               `json:"spu_id"`
	SKUCode       string               `json:"sku_code"`
	Barcode       string               `json:"barcode,omitempty"`
	Status        string               `json:"status,omitempty"`
	MinOrderQty   int                  `json:"min_order_qty,omitempty"`
	Specs         []SkuSpec            `json:"specs"`
	DefaultValues SkuGeneratorDefaults `json:"default_values,omitempty"`
}

// SkuUpsertRequest batches multiple SKU creations.
type SkuUpsertRequest struct {
	SKUs []SkuUpsertPayload `json:"skus"`
}

// SkuUpsertResult summarizes the outcome of a bulk create.
type SkuUpsertResult struct {
	Created   int          `json:"created"`
	Skipped   []string     `json:"skipped"`
	Summaries []SkuSummary `json:"summaries"`
}

// SkuSummary exposes minimal SKU info back to the caller.
type SkuSummary struct {
	ID      string    `json:"id"`
	SPUID   string    `json:"spu_id"`
	SKUCode string    `json:"sku_code"`
	Status  string    `json:"status"`
	Specs   []SkuSpec `json:"specs"`
}

const (
	BulkTaskStatusPending   = "pending"
	BulkTaskStatusApproved  = "approved"
	BulkTaskStatusRejected  = "rejected"
	BulkTaskStatusRunning   = "running"
	BulkTaskStatusSucceeded = "succeeded"
	BulkTaskStatusFailed    = "failed"
	BulkTaskStatusCancelled = "cancelled"
)

const (
	BulkOpPriceFixed       = "price_fixed"
	BulkOpPricePercent     = "price_percent"
	BulkOpInventoryFixed   = "inventory_fixed"
	BulkOpInventoryReplace = "inventory_replace"
)

// BulkAdjustmentScope describes the affected SKUs or filter for a bulk task.
type BulkAdjustmentScope struct {
	SKUIDs  []string               `json:"sku_ids"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// BulkAdjustmentOperation captures the operation type and raw JSON value.
type BulkAdjustmentOperation struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

// BulkApprovalContext keeps approval rationale for the task.
type BulkApprovalContext struct {
	ThresholdAmount float64 `json:"threshold_amount,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// BulkAdjustmentRequest is the request payload for submitting a bulk task.
type BulkAdjustmentRequest struct {
	Scope            BulkAdjustmentScope     `json:"scope"`
	Operation        BulkAdjustmentOperation `json:"operation"`
	ApprovalContext  *BulkApprovalContext    `json:"approval_context,omitempty"`
	TenantUUID       string                  `json:"-"`
	RequestedBy      string                  `json:"-"`
	RequestedByRoles []string                `json:"-"`
}

// BulkTaskStats summarizes per-task execution counters.
type BulkTaskStats struct {
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

// BulkTaskResponse mirrors the contract for submit/get bulk task endpoints.
type BulkTaskResponse struct {
	TaskID            string         `json:"task_id"`
	TaskType          string         `json:"task_type,omitempty"`
	Scope             map[string]any `json:"scope,omitempty"`
	Operation         map[string]any `json:"operation,omitempty"`
	ApprovalRequired  bool           `json:"approval_required"`
	ApprovalState     string         `json:"approval_state,omitempty"`
	ApprovalReason    string         `json:"approval_reason,omitempty"`
	ApprovalThreshold float64        `json:"approval_threshold,omitempty"`
	Status            string         `json:"status"`
	AffectedCount     int            `json:"affected_count,omitempty"`
	SubmittedBy       string         `json:"submitted_by,omitempty"`
	ApprovedBy        string         `json:"approved_by,omitempty"`
	ErrorReportURL    string         `json:"error_report,omitempty"`
	Stats             *BulkTaskStats `json:"stats,omitempty"`
	Result            map[string]any `json:"result,omitempty"`
	CreatedAt         *time.Time     `json:"created_at,omitempty"`
	UpdatedAt         *time.Time     `json:"updated_at,omitempty"`
	ApprovedAt        *time.Time     `json:"approved_at,omitempty"`
}

// BulkTaskApprovalDecision captures approval actions from reviewers.
type BulkTaskApprovalDecision struct {
	Decision string `json:"decision"`
	Note     string `json:"note,omitempty"`
}

// SkuImportRow contains normalized import row data.
type SkuImportRow struct {
	SKUID     string   `json:"sku_id,omitempty"`
	SKUCode   string   `json:"sku_code"`
	Price     *float64 `json:"price,omitempty"`
	Inventory *int64   `json:"inventory,omitempty"`
}

// SkuImportRequest represents an uploaded file ready for parsing.
type SkuImportRequest struct {
	FileName string
	Mode     string
	Content  []byte
}

// SkuExportRequest configures export format and filters.
type SkuExportRequest struct {
	Format  string
	Filters map[string]string
	Limit   int
}

const (
	BarcodeModeAuto   = "auto"
	BarcodeModeManual = "manual"
)

// BarcodeGenerateRequest defines payload for barcode generation / validation.
type BarcodeGenerateRequest struct {
	Mode          string   `json:"mode"`
	Prefix        string   `json:"prefix,omitempty"`
	Count         int      `json:"count,omitempty"`
	Length        int      `json:"length,omitempty"`
	Codes         []string `json:"codes,omitempty"`
	LabelTemplate string   `json:"label_template,omitempty"`
}

// BarcodeCheckResult contains uniqueness check outcome for a barcode.
type BarcodeCheckResult struct {
	Barcode       string `json:"barcode"`
	Unique        bool   `json:"unique"`
	ConflictSKUID string `json:"conflict_sku_id,omitempty"`
}

// BarcodeLabel exposes lightweight label payload to render/print.
type BarcodeLabel struct {
	Barcode string `json:"barcode"`
	SVG     string `json:"svg"`
}

// BarcodeBatchResult wraps validation results and printable labels.
type BarcodeBatchResult struct {
	Items  []BarcodeCheckResult `json:"items"`
	Labels []BarcodeLabel       `json:"labels,omitempty"`
}

// SerialRecordInput captures required fields to create/update SKU serials.
type SerialRecordInput struct {
	SerialNo  string     `json:"serial_no"`
	BatchNo   string     `json:"batch_no,omitempty"`
	Status    string     `json:"status,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// SerialRecordFilters allows optional filtering for serial listing.
type SerialRecordFilters struct {
	Status string `json:"status,omitempty"`
	Batch  string `json:"batch_no,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// SerialRecordDTO represents API response payload for serial entries.
type SerialRecordDTO struct {
	ID        string     `json:"id"`
	SKUID     string     `json:"sku_id"`
	SerialNo  string     `json:"serial_no"`
	BatchNo   string     `json:"batch_no,omitempty"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	AuditLog  string     `json:"audit_log_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ChannelMappingRequest carries payload to upsert channel metadata.
type ChannelMappingRequest struct {
	ChannelCode   string           `json:"channel_code"`
	ChannelSKUId  string           `json:"channel_sku_id"`
	Status        string           `json:"status,omitempty"`
	PublishTime   *time.Time       `json:"publish_time,omitempty"`
	SyncMode      string           `json:"sync_mode,omitempty"`
	PriceOverride map[string]any   `json:"price_override,omitempty"`
	MediaOverride []map[string]any `json:"media_override,omitempty"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
}

// ChannelMapping represents persisted channel state returned to clients.
type ChannelMapping struct {
	ID              string           `json:"id"`
	ChannelCode     string           `json:"channel_code"`
	ChannelSKUId    string           `json:"channel_sku_id"`
	Status          string           `json:"status"`
	PublishTime     *time.Time       `json:"publish_time,omitempty"`
	SyncMode        string           `json:"sync_mode,omitempty"`
	PriceOverride   map[string]any   `json:"price_override,omitempty"`
	MediaOverride   []map[string]any `json:"media_override,omitempty"`
	LastError       string           `json:"last_error,omitempty"`
	PublishTaskID   *string          `json:"publish_task_id,omitempty"`
	LastPublishedAt *time.Time       `json:"last_published_at,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
}

// ChannelPublishRequest triggers publish with optional force flag.
type ChannelPublishRequest struct {
	ChannelCode string `json:"channel_code"`
	Force       bool   `json:"force,omitempty"`
}

// InventoryWarehouseSnapshot captures per-warehouse inventory counters.
type InventoryWarehouseSnapshot struct {
	WarehouseID  string     `json:"warehouse_id"`
	AvailableQty int64      `json:"available_qty"`
	LockedQty    int64      `json:"locked_qty"`
	InTransitQty int64      `json:"in_transit_qty"`
	SafetyStock  int64      `json:"safety_stock"`
	AlertLevel   string     `json:"alert_level,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

// InventorySnapshot summarizes SKU inventory health.
type InventorySnapshot struct {
	Warehouses   []InventoryWarehouseSnapshot `json:"warehouses"`
	Total        InventoryWarehouseSnapshot   `json:"summary"`
	LastSyncedAt *time.Time                   `json:"last_synced_at,omitempty"`
	IsStale      bool                         `json:"is_stale"`
}

// InventoryAdjustRequest represents an incremental adjustment payload for a SKU inventory.
// Delta is an integer unit count and MUST NOT result in negative available inventory.
type InventoryAdjustRequest struct {
	Delta int64 `json:"delta"`
}
