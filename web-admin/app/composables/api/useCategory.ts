import { apiDel, apiGet, apiPatch, apiPost, apiPut } from './_client'
import type { ApiResponse } from './_base'

export type CategoryStatus = 'enabled' | 'disabled'

export interface CategoryNode {
  id: string
  parentId?: string | null
  code: string
  displayName: string
  aliasSlug: string
  path: string
  level: number
  sortOrder: number
  status: CategoryStatus | string
  imageUrl?: string
  isFeatured?: boolean
  children?: CategoryNode[]
}

export interface CategoryTreeResponse {
  items: CategoryNode[]
}

export interface CategoryListParams {
  keyword?: string
  status?: string
  parentId?: string
  page?: number
  pageSize?: number
}

export interface CategoryListResponse {
  items: CategoryNode[]
  meta: { total: number; page: number; pageSize: number }
}

export interface CategoryCreatePayload {
  parentId?: string | null
  code: string
  displayName: string
  aliasSlug: string
  sortOrder?: number
  status?: CategoryStatus
  templateId?: string | null
  seoTitle?: string
  seoDescription?: string
  seoKeywords?: string
  imageUrl?: string
  isFeatured?: boolean
}

export interface CategoryUpdatePayload {
  code?: string
  displayName?: string
  aliasSlug?: string
  sortOrder?: number
  templateId?: string | null
  seoTitle?: string
  seoDescription?: string
  seoKeywords?: string
  imageUrl?: string
  isFeatured?: boolean
}

export interface CategoryMovePayload {
  parentId?: string | null
  sortOrder?: number
}

export interface CategoryStatusPayload {
  status: CategoryStatus
}

export type CategorySaleSpecStatus = 'active' | 'disabled' | string

export interface CategorySaleSpecOption {
  id?: string
  groupId?: string
  code: string
  name: string
  sortOrder: number
  meta?: Record<string, any> | null
  status: CategorySaleSpecStatus
}

export interface CategorySaleSpecGroup {
  id?: string
  categoryId?: string
  code: string
  name: string
  sortOrder: number
  required: boolean
  allowCustom: boolean
  status: CategorySaleSpecStatus
  options: CategorySaleSpecOption[]
}

export interface ReplaceCategorySaleSpecsPayload {
  groups: CategorySaleSpecGroup[]
}

export interface CategorySaleSpecsResponse {
  groups: CategorySaleSpecGroup[]
}

export function useCategoryApi() {
  const basePath = 'admin/product/categories'

  const unwrap = async <T>(promise: Promise<ApiResponse<T>>) => {
    const resp = await promise
    return resp.data
  }

  const tree = (init?: any) =>
    unwrap(apiGet<ApiResponse<CategoryTreeResponse>>(`${basePath}/tree`, undefined, init))

  const list = (params?: CategoryListParams, init?: any) =>
    unwrap(apiGet<ApiResponse<CategoryListResponse>>(basePath, params, init))

  const create = (payload: CategoryCreatePayload, init?: any) =>
    unwrap(apiPost<ApiResponse<CategoryNode>>(basePath, payload, init))

  const update = (id: string, payload: CategoryUpdatePayload, init?: any) =>
    unwrap(apiPatch<ApiResponse<CategoryNode>>(`${basePath}/${id}`, payload, init))

  const move = (id: string, payload: CategoryMovePayload, init?: any) =>
    unwrap(apiPost<ApiResponse<CategoryNode>>(`${basePath}/${id}/move`, payload, init))

  const setStatus = (id: string, payload: CategoryStatusPayload, init?: any) =>
    unwrap(apiPatch<ApiResponse<CategoryNode>>(`${basePath}/${id}/status`, payload, init))

  const remove = (id: string, init?: any) =>
    unwrap(apiDel<ApiResponse<{ ok: boolean }>>(`${basePath}/${id}`, undefined, init))

  const saleSpecs = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<CategorySaleSpecsResponse>>(`${basePath}/${id}/sale-specs`, undefined, init))

  const replaceSaleSpecs = (id: string, payload: ReplaceCategorySaleSpecsPayload, init?: any) =>
    unwrap(apiPut<ApiResponse<CategorySaleSpecsResponse>>(`${basePath}/${id}/sale-specs`, payload, init))

  return {
    tree,
    list,
    create,
    update,
    move,
    setStatus,
    remove,
    saleSpecs,
    replaceSaleSpecs,
  }
}
