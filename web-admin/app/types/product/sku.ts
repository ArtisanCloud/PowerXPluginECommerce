export type SkuStatus = 'draft' | 'ready' | 'online' | 'offline'

export interface SkuSpecValue {
  specId: string
  specName: string
  valueId: string
  valueName: string
}

export interface SkuPriceTier {
  minQty: number
  price: number
}

export interface SkuPriceReference {
  priceListId: string
  currency?: string
  tiers?: SkuPriceTier[]
}

export interface SkuInventoryWarehouse {
  warehouseId: string
  availableQty: number
  lockedQty: number
  inTransitQty: number
  safetyStock: number
  alertLevel?: 'alert' | 'warning'
  lastSyncedAt?: string
}

export interface SkuInventorySnapshot {
  warehouses: SkuInventoryWarehouse[]
  summary: SkuInventoryWarehouse
  lastSyncedAt?: string
  isStale?: boolean
}

export interface SkuInventoryAdjustRequest {
  delta: number
}

export type SkuBarcodeMode = 'auto' | 'manual'

export interface SkuBarcodeGenerateRequest {
  mode?: SkuBarcodeMode
  prefix?: string
  count?: number
  length?: number
  codes?: string[]
  labelTemplate?: string
}

export interface SkuBarcodeCheckResult {
  barcode: string
  unique: boolean
  conflictSkuId?: string
}

export interface SkuBarcodeLabel {
  barcode: string
  svg: string
}

export interface SkuBarcodeBatchResult {
  items: SkuBarcodeCheckResult[]
  labels?: SkuBarcodeLabel[]
}

export type SkuChannelStatus = 'pending' | 'published' | 'failed' | 'offline'
export type SkuSyncMode = 'push' | 'pull' | 'hybrid'

export interface SkuMediaAsset {
  id?: string
  type: 'image' | 'video'
  url: string
  isPrimary?: boolean
  channelOverride?: string
  sortOrder?: number
}

export interface SkuChannelMapping {
  id?: string
  channelCode: string
  channelSkuId: string
  status: SkuChannelStatus
  publishTime?: string
  syncMode?: SkuSyncMode
  lastError?: string
  priceOverride?: Record<string, unknown>
  mediaOverride?: SkuMediaAsset[]
  publishTaskId?: string
  metadata?: Record<string, unknown>
}

export interface SkuChannelMappingPayload {
  channelCode: string
  channelSkuId: string
  status?: SkuChannelStatus
  syncMode?: SkuSyncMode
  publishTime?: string
  priceOverride?: Record<string, unknown>
  mediaOverride?: SkuMediaAsset[]
  metadata?: Record<string, unknown>
}

export interface SkuChannelPublishPayload {
  channelCode: string
  force?: boolean
}

export interface SkuDefaultLogistics {
  hsCode?: string
  packageType?: string
  weight?: number
  dimensions?: string
}

export interface ProductSku {
  id: string
  tenantUuid: string
  spuId: string
  spuName?: string
  skuCode: string
  specs: SkuSpecValue[]
  specDisplay?: string
  salePrice?: number
  currency?: string
  barcode?: string
  status: SkuStatus
  lifecyclePhase?: string
  minOrderQty?: number
  priceRefs?: SkuPriceReference[]
  logistics?: SkuDefaultLogistics
  tags?: string[]
  media?: SkuMediaAsset[]
  inventory?: SkuInventoryWarehouse[]
  channels?: SkuChannelMapping[]
  createdAt?: string
  updatedAt?: string
}

export interface SkuGeneratorDefaults {
  barcodePrefix?: string
  costPrice?: number
  minOrderQty?: number
  weight?: number
  dimensions?: string
}

export interface SkuGeneratorCandidate {
  specs: SkuSpecValue[]
  previewSkuCode: string
  defaultValues: SkuGeneratorDefaults
  selected?: boolean
  exists?: boolean
  conflictReasons?: string[]
}

export interface SkuSpecSelection {
  specId: string
  specName?: string
  valueIds: string[]
  values?: Array<{ valueId: string; valueName?: string; valueCode?: string }>
}

export interface SkuGeneratorRequest {
  specSelections: SkuSpecSelection[]
  defaults: SkuGeneratorDefaults
}

export interface SkuGeneratorResponse {
  candidates: SkuGeneratorCandidate[]
}

export type BulkTaskType =
  | 'price_adjustment'
  | 'inventory_adjustment'
  | 'channel_publish'
  | 'import'
  | 'export'

export type SkuBulkTaskStatus =
  | 'pending'
  | 'approved'
  | 'running'
  | 'succeeded'
  | 'failed'
  | 'cancelled'

export interface SkuBulkTaskStats {
  succeeded: number
  failed: number
}

export interface SkuBulkTask {
  taskId: string
  taskType: BulkTaskType
  scope?: Record<string, unknown>
  operation?: Record<string, unknown>
  status: SkuBulkTaskStatus
  approvalRequired: boolean
  approvalState?: 'pending' | 'approved' | 'rejected'
  approvalReason?: string
  approvalThreshold?: number
  affectedCount?: number
  submittedBy?: string
  approvedBy?: string
  approvedAt?: string
  errorReportUrl?: string
  stats?: SkuBulkTaskStats
  createdAt?: string
  updatedAt?: string
  result?: Record<string, unknown>
}

export interface SkuSerialRecord {
  id?: string
  skuId: string
  serialNo: string
  batchNo?: string
  expiresAt?: string
  status?: 'available' | 'allocated' | 'consumed'
  auditLogId?: string
  createdAt?: string
}

export interface SkuSerialRecordInput {
  serialNo: string
  batchNo?: string
  status?: SkuSerialRecord['status']
  expiresAt?: string
}

export interface SkuSerialQuery {
  status?: string
  batch?: string
  limit?: number
}

export interface SkuListResponse {
  items?: ProductSku[]
  list?: ProductSku[]
  page?: number
  page_size?: number
  pageSize?: number
  total?: number
  pagination?: {
    page: number
    size: number
    total: number
  }
}

export interface SkuUpsertRequest {
  skus: Array<{
    spuId: string
    skuCode: string
    specs: SkuSpecValue[]
    barcode?: string
    status?: SkuStatus
    minOrderQty?: number
    defaultValues?: SkuGeneratorDefaults
  }>
}

export interface SkuUpsertResult {
  created: number
  skipped: string[]
  summaries: Array<{
    id: string
    spuId: string
    skuCode: string
    status: SkuStatus
    specs: SkuSpecValue[]
  }>
}

export type BulkOperationType =
  | 'price_fixed'
  | 'price_percent'
  | 'inventory_fixed'
  | 'inventory_replace'

export type SkuImportMode = 'upsert' | 'inventory-only'

export interface SkuExportPayload {
  format?: string
  limit?: number
  filters?: {
    status?: string
    keyword?: string
    spu_id?: string
  }
}

export interface BulkAdjustmentScope {
  skuIds: string[]
  filters?: Record<string, any>
}

export interface BulkAdjustmentOperation {
  type: BulkOperationType
  value: number | Record<string, unknown>
}

export interface BulkApprovalContext {
  thresholdAmount?: number
  reason?: string
}

export interface SkuBulkTaskRequest {
  scope: BulkAdjustmentScope
  operation: BulkAdjustmentOperation
  approvalContext?: BulkApprovalContext
}
