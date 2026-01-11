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
			unwrap(apiGet<SkuListResponse>(basePath, params, init)),
		listBySpu: (spuId: string, params?: Omit<SkuListParams, 'spuId'>, init?: any) =>
			unwrap(apiGet<SkuListResponse>(basePath, { ...params, spuId }, init)),
		generate: (spuId: string, payload: SkuGeneratorRequest, init?: any) =>
			unwrap(
				apiPost<ApiResponse<SkuGeneratorResponse>>(
					`admin/product/spus/${spuId}/skus/generate`,
					payload,
					init,
				),
			),
		create: (payload: SkuUpsertRequest, init?: any) =>
			unwrap(apiPost<SkuUpsertResult>(basePath, payload, init)),
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
