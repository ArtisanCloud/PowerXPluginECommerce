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

export interface SkuInventorySnapshot {
  warehouseId: string
  availableQty: number
  lockedQty: number
  inTransitQty: number
  safetyStock: number
  lastSyncedAt?: string
  alertThreshold?: number
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
  taskId?: string
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
  skuCode: string
  specs: SkuSpecValue[]
  barcode?: string
  status: SkuStatus
  lifecyclePhase?: string
  minOrderQty?: number
  priceRefs?: SkuPriceReference[]
  logistics?: SkuDefaultLogistics
  tags?: string[]
  media?: SkuMediaAsset[]
  inventory?: SkuInventorySnapshot[]
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
  status: SkuBulkTaskStatus
  approvalRequired: boolean
  approvalState?: 'pending' | 'approved' | 'rejected'
  affectedCount?: number
  submittedBy?: string
  approvedBy?: string
  errorReportUrl?: string
  stats?: SkuBulkTaskStats
  createdAt?: string
  updatedAt?: string
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

export interface SkuListResponse {
  items: ProductSku[]
  pagination: {
    page: number
    size: number
    total: number
  }
}
