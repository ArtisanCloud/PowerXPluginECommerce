import { apiGet, apiPatch, apiPost } from './_client'
import type { ApiResponse } from './_base'

export type TemplateStatus = 'enabled' | 'disabled'

export interface CategoryTemplateField {
  id?: string
  fieldKey: string
  group?: string
  fieldType: string
  required: boolean
  sortOrder: number
  validationRules?: Record<string, any>
  defaultValue?: any
  i18nLabel?: Record<string, any>
  i18nHelp?: Record<string, any>
  inheritable?: boolean
}

export interface CategoryTemplate {
  id: string
  name: string
  status: string
  applicableLevel?: number | null
  notes?: string
  currentPublishedVersionId?: string | null
  currentPublishedVersionNo?: number
  lastPublishedAt?: string | null
  fields?: CategoryTemplateField[]
  updatedAt?: string
  createdAt?: string
}

export interface TemplateListResponse {
  items: CategoryTemplate[]
  meta: { total: number; page: number; pageSize: number }
}

export interface TemplateCreatePayload {
  name: string
  status?: TemplateStatus
  applicableLevel?: number | null
  notes?: string
  fields?: CategoryTemplateField[]
}

export interface TemplateUpdatePayload {
  name?: string
  status?: TemplateStatus
  applicableLevel?: number | null
  notes?: string
  fields?: CategoryTemplateField[]
}

export interface TemplatePublishPayload {
  publishedBy?: string
}

export interface TemplateRollbackPayload {
  targetVersionId: string
  rollbackBy?: string
}

export interface TemplateVersionSummary {
  id: string
  templateId: string
  versionNumber: number
  publishedAt?: string
  publishedBy?: string
  rollbackFrom?: string
  createdAt?: string
}

export interface EffectiveTemplateResponse {
  templateId: string
  versionId: string
  versionNumber: number
  publishedAt?: string
  template: CategoryTemplate
  fields: CategoryTemplateField[]
}

export interface TemplateSimulatePayload {
  categoryId?: string
  templateId?: string
  attributes?: Record<string, any>
}

export interface TemplateImpactSummary {
  templateId: string
  boundCategories: number
  boundCategoryIds?: string[]
  affectedProducts: number
}

export interface TemplateRecheckPayload {
  reason?: string
  dryRun?: boolean
}

export interface TemplateRecheckResult {
  jobRunId: string
  status: string
}

export function useCategoryTemplateApi() {
  const basePath = 'admin/product/category-templates'

  const unwrap = async <T>(promise: Promise<ApiResponse<T>>) => {
    const resp = await promise
    return resp.data
  }

  const list = (params?: { keyword?: string; status?: string; page?: number; pageSize?: number }, init?: any) =>
    unwrap(apiGet<ApiResponse<TemplateListResponse>>(basePath, params, init))

  const get = (id: string, init?: any) => unwrap(apiGet<ApiResponse<CategoryTemplate>>(`${basePath}/${id}`, undefined, init))

  const create = (payload: TemplateCreatePayload, init?: any) =>
    unwrap(apiPost<ApiResponse<CategoryTemplate>>(basePath, payload, init))

  const update = (id: string, payload: TemplateUpdatePayload, init?: any) =>
    unwrap(apiPatch<ApiResponse<CategoryTemplate>>(`${basePath}/${id}`, payload, init))

  const publish = (id: string, payload?: TemplatePublishPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<TemplateVersionSummary>>(`${basePath}/${id}/publish`, payload ?? {}, init))

  const rollback = (id: string, payload: TemplateRollbackPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<TemplateVersionSummary>>(`${basePath}/${id}/rollback`, payload, init))

  const versions = (id: string, params?: { limit?: number }, init?: any) =>
    unwrap(apiGet<ApiResponse<{ items: TemplateVersionSummary[] }>>(`${basePath}/${id}/versions`, params, init))

  const effective = (categoryId: string, init?: any) =>
    unwrap(apiGet<ApiResponse<EffectiveTemplateResponse>>(`${basePath}/effective`, { categoryId }, init))

  const preview = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<CategoryTemplate>>(`${basePath}/${id}/preview`, undefined, init))

  const simulate = (id: string, payload: TemplateSimulatePayload, init?: any) =>
    unwrap(apiPost<ApiResponse<{ ok: boolean }>>(`${basePath}/${id}/simulate`, payload, init))

  const impact = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<TemplateImpactSummary>>(`${basePath}/${id}/impact`, undefined, init))

  const triggerRecheck = (id: string, payload?: TemplateRecheckPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<TemplateRecheckResult>>(`${basePath}/${id}/impact/recheck`, payload ?? {}, init))

  return {
    list,
    get,
    create,
    update,
    publish,
    rollback,
    versions,
    effective,
    preview,
    simulate,
    impact,
    triggerRecheck,
  }
}

