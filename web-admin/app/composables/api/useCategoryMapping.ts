import { apiGet, useApiClient } from './_client'
import type { ApiResponse } from './_base'

export interface CategoryMapping {
  id: string
  categoryId: string
  channel: string
  platformCategoryId: string
  strategy: string
  syncStatus: string
  metadata?: Record<string, any>
  updatedAt?: string
  createdAt?: string
}

export interface MappingListResponse {
  items: CategoryMapping[]
}

export interface MappingUpsertPayload {
  operation?: 'upsert' | 'delete'
  channel: string
  platformCategoryId?: string
  strategy?: string
  syncStatus?: string
  metadata?: Record<string, any>
}

export interface MappingImportResult {
  total: number
  succeeded: number
  failed: number
  errors?: Array<{ row: number; field: string; message: string }>
}

export interface AuditRecord {
  id: string
  action: string
  permissionCode: string
  actorId: string
  summary?: string
  diff?: any
  occurredAt: string
}

export function useCategoryMappingApi() {
  const { client } = useApiClient()

  const unwrap = async <T>(promise: Promise<ApiResponse<T>>) => {
    const resp = await promise
    return resp.data
  }

  const list = (categoryId: string, init?: any) =>
    unwrap(apiGet<ApiResponse<MappingListResponse>>(`admin/product/categories/${categoryId}/mappings`, undefined, init))

  const upsert = (categoryId: string, payload: MappingUpsertPayload, init?: any) =>
    unwrap(client<ApiResponse<CategoryMapping>>(`admin/product/categories/${categoryId}/mappings`, { method: 'POST', body: JSON.stringify(payload), ...init }))

  const remove = (categoryId: string, channel: string, init?: any) =>
    unwrap(
      client<ApiResponse<{ ok: boolean }>>(`admin/product/categories/${categoryId}/mappings`, {
        method: 'POST',
        body: JSON.stringify({ operation: 'delete', channel }),
        ...init,
      }),
    )

  const importMappings = async (file: File, categoryId: string, init?: any) => {
    const form = new FormData()
    form.append('file', file)
    form.append('kind', 'mappings')
    form.append('categoryId', categoryId)
    const resp = await client<ApiResponse<MappingImportResult>>('admin/product/categories/import', {
      method: 'POST',
      query: { kind: 'mappings', categoryId },
      body: form,
      ...init,
    })
    return resp.data
  }

  const exportMappings = async (categoryId: string, init?: any) => {
    const resp = await client.raw('admin/product/categories/export', {
      method: 'POST',
      body: JSON.stringify({ kind: 'mappings', categoryId }),
      responseType: 'blob' as const,
      ...init,
    })
    return resp
  }

  const audit = (categoryId: string, params?: { limit?: number }, init?: any) =>
    unwrap(apiGet<ApiResponse<{ items: AuditRecord[] }>>(`admin/product/categories/${categoryId}/audit`, params, init))

  return {
    list,
    upsert,
    remove,
    importMappings,
    exportMappings,
    audit,
  }
}

