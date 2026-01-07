import { useRuntimeConfig } from '#imports'
import { defineStore } from 'pinia'
import { useSkuApi } from '~/composables/api/useSku'
import type { SkuListParams } from '~/composables/api/useSku'
import type {
	ProductSku,
	SkuBarcodeBatchResult,
	SkuBarcodeGenerateRequest,
	SkuBulkTask,
	SkuBulkTaskStatus,
	SkuChannelMapping,
	SkuChannelMappingPayload,
	SkuChannelPublishPayload,
	SkuGeneratorCandidate,
	SkuGeneratorDefaults,
	SkuGeneratorRequest,
	SkuInventorySnapshot,
	SkuInventoryWarehouse,
	SkuSerialQuery,
	SkuSerialRecord,
	SkuSerialRecordInput,
	SkuSpecValue,
	SkuUpsertRequest,
	SkuUpsertResult,
} from '~/types/product/sku'

interface ProductSkuState {
	items: ProductSku[]
	total: number
	loading: boolean
	generatorLoading: boolean
	upsertLoading: boolean
	bulkTasks: Record<string, SkuBulkTask>
	bulkTaskStatus: Record<string, SkuBulkTaskStatus>
	generatorCandidates: SkuGeneratorCandidate[]
	generatorDefaults: SkuGeneratorDefaults
	apiBase: string
	channelMappings: Record<string, SkuChannelMapping[]>
	inventorySnapshots: Record<string, SkuInventorySnapshot>
	barcodeResults: Record<string, SkuBarcodeBatchResult | null>
	serialRecords: Record<string, SkuSerialRecord[]>
	lastQuery: SkuListParams | null
	itemsMap: Record<string, ProductSku>
}

const defaultConfig = () => {
	const config = useRuntimeConfig()
	const { productSkuApiBase } = config.public
	return productSkuApiBase || '/_p/com.powerx.plugin.ecommerce/api/v1/products/skus'
}

const ACTIVE_BULK_STATUSES: SkuBulkTaskStatus[] = ['pending', 'approved', 'running']

const normalizeListParams = (params?: SkuListParams): SkuListParams => {
	const safe = params ?? {}
	const normalized: SkuListParams = {
		page: typeof safe.page === 'number' && safe.page > 0 ? safe.page : 1,
		pageSize: typeof safe.pageSize === 'number' && safe.pageSize > 0 ? safe.pageSize : 20,
	}
	const spuId = String(safe.spuId ?? '').trim()
	if (spuId) normalized.spuId = spuId
	const status = String(safe.status ?? '').trim()
	if (status) normalized.status = status
	const channel = String(safe.channel ?? '').trim()
	if (channel) normalized.channel = channel
	return normalized
}

export const useProductSkuStore = defineStore('product-sku', {
	state: (): ProductSkuState => ({
		items: [],
		total: 0,
		loading: false,
		generatorLoading: false,
		upsertLoading: false,
		bulkTasks: {},
		bulkTaskStatus: {},
		generatorCandidates: [],
		generatorDefaults: {},
		apiBase: defaultConfig(),
		channelMappings: {},
		inventorySnapshots: {},
		barcodeResults: {},
		serialRecords: {},
		lastQuery: null,
		itemsMap: {},
	}),
	actions: {
		async fetchList(params?: SkuListParams) {
			const api = useSkuApi()
			const query = normalizeListParams(params)
			this.loading = true
			this.lastQuery = query
			try {
				const response = await api.list(query)
				const items = response?.items ?? (response as any)?.list ?? []
				const totalFromServer = response?.pagination?.total ?? (response as any)?.total ?? (response as any)?.meta?.total
				this.setItems(items as ProductSku[], typeof totalFromServer === 'number' ? totalFromServer : items.length)
				return this.items
			} catch (error) {
				throw error
			} finally {
				this.loading = false
			}
		},
		async refreshList() {
			if (this.lastQuery) {
				return this.fetchList(this.lastQuery)
			}
			return this.fetchList()
		},
		setApiBase(base: string) {
			if (!base) return
			this.apiBase = base.replace(/\/+$/, '')
		},
		async fetchChannelMappings(skuId: string) {
			if (!skuId) return []
			const api = useSkuApi()
			const response = await api.listChannels(skuId).catch(() => [])
			const normalized = (response ?? []).map(normalizeChannelMapping)
			this.channelMappings[skuId] = normalized
			return normalized
		},
		async saveChannelMapping(skuId: string, payload: SkuChannelMappingPayload) {
			if (!skuId) {
				throw new Error('skuId is required')
			}
			const api = useSkuApi()
			const response = await api.saveChannelMapping(skuId, payload)
			const mapping = normalizeChannelMapping(response)
			const current = this.channelMappings[skuId] ?? []
			const filtered = current.filter((item) => item.channelCode !== mapping.channelCode)
			this.channelMappings[skuId] = [...filtered, mapping].sort((a, b) => a.channelCode.localeCompare(b.channelCode))
			return mapping
		},
		async publishChannelMapping(skuId: string, payload: SkuChannelPublishPayload) {
			if (!skuId) {
				throw new Error('skuId is required')
			}
			const api = useSkuApi()
			const response = await api.publishChannelMapping(skuId, payload)
			await this.fetchChannelMappings(skuId)
			return response
		},
		async fetchInventorySnapshot(skuId: string) {
			if (!skuId) return null
			const api = useSkuApi()
			const response = await api.getInventorySnapshot(skuId).catch(() => null)
			if (!response) {
				return null
			}
			const snapshot = normalizeInventorySnapshot(response)
			this.inventorySnapshots[skuId] = snapshot
			return snapshot
		},
		async generateBarcodes(skuId: string, payload: SkuBarcodeGenerateRequest) {
			if (!skuId) {
				throw new Error('skuId is required')
			}
			const api = useSkuApi()
			const result = await api.generateBarcodes(skuId, payload)
			this.barcodeResults[skuId] = result
			return result
		},
		async fetchSerialRecords(skuId: string, filters?: SkuSerialQuery) {
			if (!skuId) return []
			const api = useSkuApi()
			const response = await api.listSerials(skuId, filters).catch(() => [])
			const normalized = response ?? []
			this.serialRecords[skuId] = normalized
			return normalized
		},
		async createSerialRecord(skuId: string, payload: SkuSerialRecordInput) {
			if (!skuId) {
				throw new Error('skuId is required')
			}
			const api = useSkuApi()
			const record = await api.createSerial(skuId, payload)
			const existing = this.serialRecords[skuId] ?? []
			this.serialRecords[skuId] = [record, ...existing]
			return record
		},
		setItems(items: ProductSku[], total?: number) {
			this.items = items
			this.itemsMap = {}
			for (const item of items) {
				if (item?.id) {
					this.itemsMap[item.id] = item
				}
			}
			if (typeof total === 'number') {
				this.total = total
			} else {
				this.total = items.length
			}
		},
		setGeneratorDefaults(defaults: SkuGeneratorDefaults) {
			this.generatorDefaults = { ...defaults }
		},
		trackBulkTask(task: SkuBulkTask) {
			this.bulkTasks[task.taskId] = task
			if (task.status) {
				this.bulkTaskStatus[task.taskId] = task.status
			}
		},
		updateBulkTaskStatus(taskId: string, status: SkuBulkTaskStatus) {
			if (!taskId) return
			this.bulkTaskStatus[taskId] = status
		},
		async fetchGeneratorCandidates(spuId: string, request: SkuGeneratorRequest) {
			const api = useSkuApi()
			this.generatorLoading = true
			try {
				const resp = await api.generate(spuId, request)
				this.generatorCandidates = (resp?.candidates ?? []).map(normalizeCandidate)
				return this.generatorCandidates
			} finally {
				this.generatorLoading = false
			}
		},
		async createSkus(payload: SkuUpsertRequest): Promise<SkuUpsertResult> {
			const api = useSkuApi()
			this.upsertLoading = true
			try {
				const serverPayload = {
					skus: payload.skus.map((item) => ({
						spu_id: item.spuId,
						sku_code: item.skuCode,
						barcode: item.barcode,
						status: item.status,
						min_order_qty: item.minOrderQty,
						default_values: item.defaultValues,
						specs: item.specs.map((spec) => ({
							spec_id: spec.specId,
							spec_name: spec.specName,
							value_id: spec.valueId,
							value_name: spec.valueName,
						})),
					})),
				}
				const result = normalizeUpsertResult(await api.create(serverPayload as any))
				if (result?.created) {
					await this.refreshList()
				}
				return result
			} finally {
				this.upsertLoading = false
			}
		},
		reset() {
			this.items = []
			this.total = 0
			this.bulkTasks = {}
			this.bulkTaskStatus = {}
			this.generatorCandidates = []
			this.generatorDefaults = {}
			this.channelMappings = {}
			this.inventorySnapshots = {}
			this.barcodeResults = {}
			this.serialRecords = {}
			this.itemsMap = {}
			this.lastQuery = null
		},
		async refreshBulkTask(taskId: string) {
			if (!taskId) return null
			const api = useSkuApi()
			const response = await api.getBulkTask(taskId).catch(() => null)
			if (!response) {
				return null
			}
			const normalized = normalizeBulkTaskResponse(response)
			if (normalized?.taskId) {
				this.trackBulkTask(normalized)
			}
			return normalized
		},
		async refreshActiveBulkTasks() {
			const pendingTasks = Object.values(this.bulkTasks).filter((task) => isActiveBulkTask(task))
			if (!pendingTasks.length) {
				return []
			}
			return Promise.all(
				pendingTasks.map((task) => this.refreshBulkTask(task.taskId)),
			)
		},
	},
})

function normalizeCandidate(candidate: any): SkuGeneratorCandidate {
	const specs = (candidate?.specs ?? []).map(normalizeSpec)
	return {
		specs,
		previewSkuCode: candidate?.preview_sku_code ?? candidate?.previewSkuCode ?? '',
		defaultValues: candidate?.default_values ?? candidate?.defaultValues ?? {},
		selected: candidate?.selected ?? !candidate?.exists,
		exists: Boolean(candidate?.exists),
		conflictReasons: candidate?.conflict_reasons ?? candidate?.conflictReasons ?? [],
	}
}

function normalizeUpsertResult(result: any): SkuUpsertResult {
	return {
		created: result?.created ?? 0,
		skipped: result?.skipped ?? [],
		summaries: (result?.summaries ?? []).map((item: any) => ({
			id: item?.id,
			spuId: item?.spu_id ?? item?.spuId,
			skuCode: item?.sku_code ?? item?.skuCode,
			status: item?.status,
			specs: (item?.specs ?? []).map(normalizeSpec),
		})),
	}
}

function normalizeSpec(spec: any): SkuSpecValue {
	return {
		specId: spec?.spec_id ?? spec?.specId,
		specName: spec?.spec_name ?? spec?.specName,
		valueId: spec?.value_id ?? spec?.valueId,
		valueName: spec?.value_name ?? spec?.valueName,
	}
}

function isActiveBulkTask(task?: SkuBulkTask) {
	if (!task) return false
	if (ACTIVE_BULK_STATUSES.includes(task.status)) {
		return true
	}
	if (task.approvalRequired && (task.approvalState === 'pending' || !task.approvalState)) {
		return true
	}
	return false
}

function normalizeBulkTaskResponse(task: any): SkuBulkTask {
  if (!task) {
    return {
      taskId: '',
      taskType: 'price_adjustment',
      status: 'pending',
      approvalRequired: Boolean(task?.approval_required ?? task?.approvalRequired),
    }
  }
  return {
    taskId: task?.task_id ?? task?.taskId ?? '',
    taskType: task?.task_type ?? task?.taskType ?? 'price_adjustment',
    scope: task?.scope,
    operation: task?.operation,
    status: (task?.status ?? 'pending') as SkuBulkTaskStatus,
    approvalRequired: Boolean(task?.approval_required ?? task?.approvalRequired),
    approvalState: task?.approval_state ?? task?.approvalState,
    approvalReason: task?.approval_reason ?? task?.approvalReason,
    approvalThreshold: task?.approval_threshold ?? task?.approvalThreshold,
    affectedCount: task?.affected_count ?? task?.affectedCount,
    submittedBy: task?.submitted_by ?? task?.submittedBy,
    approvedBy: task?.approved_by ?? task?.approvedBy,
    approvedAt: task?.approved_at ?? task?.approvedAt,
    errorReportUrl: task?.error_report ?? task?.errorReportUrl,
    stats: task?.stats,
    createdAt: task?.created_at ?? task?.createdAt,
    updatedAt: task?.updated_at ?? task?.updatedAt,
    result: task?.result,
  }
}

function normalizeChannelMapping(payload: any): SkuChannelMapping {
	const price = payload?.price_override ?? payload?.priceOverride
	const media = payload?.media_override ?? payload?.mediaOverride ?? []
	return {
		id: payload?.id,
		channelCode: payload?.channel_code ?? payload?.channelCode ?? '',
		channelSkuId: payload?.channel_sku_id ?? payload?.channelSkuId ?? '',
		status: payload?.status ?? 'pending',
		publishTime: payload?.publish_time ?? payload?.publishTime,
		syncMode: payload?.sync_mode ?? payload?.syncMode,
		lastError: payload?.last_error ?? payload?.lastError,
		priceOverride: price,
		mediaOverride: media,
		publishTaskId: payload?.publish_task_id ?? payload?.publishTaskId,
		metadata: payload?.metadata,
	}
}

function normalizeInventorySnapshot(raw: any): SkuInventorySnapshot {
	const warehouses = (raw?.warehouses ?? []).map((item: any) => normalizeInventoryWarehouse(item))
	const summary = normalizeInventoryWarehouse(raw?.summary, 'total')
	return {
		warehouses,
		summary,
		lastSyncedAt: raw?.last_synced_at ?? raw?.lastSyncedAt ?? summary.lastSyncedAt,
		isStale: Boolean(raw?.is_stale ?? raw?.isStale ?? false),
	}
}

function normalizeInventoryWarehouse(raw: any, fallbackId = ''): SkuInventoryWarehouse {
	if (!raw) {
		return {
			warehouseId: fallbackId,
			availableQty: 0,
			lockedQty: 0,
			inTransitQty: 0,
			safetyStock: 0,
		}
	}
	return {
		warehouseId: raw?.warehouse_id ?? raw?.warehouseId ?? fallbackId,
		availableQty: Number(raw?.available_qty ?? raw?.availableQty ?? 0),
		lockedQty: Number(raw?.locked_qty ?? raw?.lockedQty ?? 0),
		inTransitQty: Number(raw?.in_transit_qty ?? raw?.inTransitQty ?? 0),
		safetyStock: Number(raw?.safety_stock ?? raw?.safetyStock ?? 0),
		alertLevel: raw?.alert_level ?? raw?.alertLevel,
		lastSyncedAt: raw?.last_synced_at ?? raw?.lastSyncedAt,
	}
}
