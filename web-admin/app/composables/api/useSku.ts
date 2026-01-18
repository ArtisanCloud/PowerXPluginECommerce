import { apiGet, apiPost } from './_client'
import type { ApiResponse } from './_base'
import type {
	ProductSku,
	SkuBarcodeBatchResult,
	SkuBarcodeGenerateRequest,
	SkuBulkTask,
	SkuBulkTaskRequest,
	SkuChannelMapping,
	SkuChannelMappingPayload,
	SkuChannelPublishPayload,
	SkuExportPayload,
	SkuGeneratorRequest,
	SkuGeneratorResponse,
	SkuImportMode,
	SkuInventoryAdjustRequest,
	SkuInventorySnapshot,
	SkuSerialQuery,
	SkuSerialRecord,
	SkuSerialRecordInput,
	SkuUpsertRequest,
	SkuUpsertResult,
} from '~/types/product/sku'

export interface SkuListParams {
	spuId?: string
	status?: string
	channel?: string
	locale?: string
	page?: number
	pageSize?: number
}

export interface SkuListResponse {
	items: ProductSku[]
	pagination?: {
		page: number
		size: number
		total: number
	}
}

type AnySku = Record<string, any>

const normalizeSpec = (spec: AnySku) => ({
	specId: String(spec?.specId ?? spec?.spec_id ?? '').trim(),
	specName: spec?.specName ?? spec?.spec_name,
	valueId: String(spec?.valueId ?? spec?.value_id ?? '').trim(),
	valueName: spec?.valueName ?? spec?.value_name,
})

const normalizeSku = (sku: AnySku): ProductSku => ({
	id: String(sku?.id ?? '').trim(),
	tenantUuid: String(sku?.tenantUuid ?? sku?.tenant_uuid ?? '').trim(),
	spuId: String(sku?.spuId ?? sku?.spu_id ?? '').trim(),
	spuName: sku?.spuName ?? sku?.spu_name,
	skuCode: String(sku?.skuCode ?? sku?.sku_code ?? '').trim(),
	specs: Array.isArray(sku?.specs) ? sku.specs.map(normalizeSpec) : [],
	specDisplay: sku?.specDisplay ?? sku?.spec_display,
	salePrice: typeof sku?.salePrice === 'number' ? sku.salePrice : (typeof sku?.sale_price === 'number' ? sku.sale_price : undefined),
	currency: sku?.currency ?? sku?.currency_code,
	barcode: sku?.barcode,
	status: sku?.status,
	lifecyclePhase: sku?.lifecyclePhase ?? sku?.lifecycle_phase,
	minOrderQty: sku?.minOrderQty ?? sku?.min_order_qty,
	priceRefs: sku?.priceRefs ?? sku?.price_refs,
	logistics: sku?.logistics,
	tags: sku?.tags,
	media: sku?.media,
	inventory: sku?.inventory,
	channels: sku?.channels,
	createdAt: sku?.createdAt ?? sku?.created_at,
	updatedAt: sku?.updatedAt ?? sku?.updated_at,
})

const normalizeSkuList = (resp: SkuListResponse): SkuListResponse => ({
	...resp,
	items: Array.isArray(resp?.items) ? resp.items.map((item) => normalizeSku(item as AnySku)) : [],
})

export function useSkuApi() {
	const basePath = 'admin/product/skus'
	const unwrap = async <T>(promise: Promise<T | ApiResponse<T>>) => {
		const resp = await promise
		if (resp && typeof resp === 'object' && resp !== null && Object.prototype.hasOwnProperty.call(resp as any, 'data')) {
			return (resp as ApiResponse<T>).data
		}
		return resp as T
	}

	return {
		list: (params?: SkuListParams, init?: any) =>
			unwrap(apiGet<SkuListResponse>(basePath, params, init)).then(normalizeSkuList),
		get: (skuId: string, params?: { locale?: string }, init?: any) =>
			unwrap(apiGet<any>(`${basePath}/${skuId}`, params, init)).then((item) => normalizeSku(item as AnySku)),
		listBySpu: (spuId: string, params?: Omit<SkuListParams, 'spuId'>, init?: any) =>
			unwrap(apiGet<SkuListResponse>(basePath, { ...params, spuId }, init)).then(normalizeSkuList),
		generate: (spuId: string, payload: SkuGeneratorRequest, init?: any) =>
			unwrap(
				apiPost<ApiResponse<SkuGeneratorResponse>>(
					`admin/product/spus/${spuId}/skus/generate`,
					payload,
					init,
				),
			),
		create: (payload: SkuUpsertRequest, init?: any) => {
			const skus = Array.isArray(payload?.skus)
				? payload.skus.map((sku) => ({
						...sku,
						specs: Array.isArray((sku as any)?.specs) ? (sku as any).specs : [],
				  }))
				: [];
			return unwrap(apiPost<SkuUpsertResult>(basePath, { ...payload, skus }, init));
		},
		submitBulkTask: (payload: SkuBulkTaskRequest, init?: any) =>
			unwrap(apiPost<SkuBulkTask>(`${basePath}/bulk-tasks`, payload, init)),
		importByFile: (file: File, mode: SkuImportMode = 'upsert', init?: any) => {
			const formData = new FormData()
			formData.append('file', file)
			formData.append('mode', mode)
			return unwrap(apiPost<SkuBulkTask>(`${basePath}/import`, formData, init))
		},
		exportSkus: (payload: SkuExportPayload, init?: any) =>
			unwrap(apiPost<SkuBulkTask>(`${basePath}/export`, payload, init)),
		getBulkTask: (taskId: string, init?: any) =>
			unwrap(apiGet<SkuBulkTask>(`${basePath}/bulk-tasks/${taskId}`, undefined, init)),
		decideBulkTask: (taskId: string, payload: { decision: string; note?: string }, init?: any) =>
			unwrap(apiPost<SkuBulkTask>(`${basePath}/bulk-tasks/${taskId}/approval`, payload, init)),
		retryBulkTask: (taskId: string, init?: any) =>
			unwrap(apiPost<SkuBulkTask>(`${basePath}/bulk-tasks/${taskId}/retry`, {}, init)),
		listChannels: (skuId: string, init?: any) =>
			unwrap(apiGet<SkuChannelMapping[]>(`${basePath}/${skuId}/channels`, undefined, init)),
		saveChannelMapping: (skuId: string, payload: SkuChannelMappingPayload, init?: any) =>
			unwrap(apiPost<SkuChannelMapping>(`${basePath}/${skuId}/channels`, payload, init)),
		publishChannelMapping: (skuId: string, payload: SkuChannelPublishPayload, init?: any) =>
			unwrap(apiPost<SkuBulkTask>(`${basePath}/${skuId}/channels/publish`, payload, init)),
		getInventorySnapshot: (skuId: string, init?: any) =>
			unwrap(apiGet<SkuInventorySnapshot>(`${basePath}/${skuId}/inventory`, undefined, init)),
		adjustInventory: (skuId: string, payload: SkuInventoryAdjustRequest, init?: any) =>
			unwrap(apiPost<SkuInventorySnapshot>(`${basePath}/${skuId}/inventory/adjust`, payload, init)),
		generateBarcodes: (skuId: string, payload: SkuBarcodeGenerateRequest, init?: any) =>
			unwrap(apiPost<SkuBarcodeBatchResult>(`${basePath}/${skuId}/barcodes`, payload, init)),
		listSerials: (skuId: string, params?: SkuSerialQuery, init?: any) =>
			unwrap(apiGet<SkuSerialRecord[]>(`${basePath}/${skuId}/serials`, params, init)),
		createSerial: (skuId: string, payload: SkuSerialRecordInput, init?: any) =>
			unwrap(apiPost<SkuSerialRecord>(`${basePath}/${skuId}/serials`, payload, init)),
	}
}
