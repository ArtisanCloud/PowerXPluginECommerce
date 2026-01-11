import { apiDel, apiGet, apiPost, apiPatch, apiPut, useApiClient } from './_client'
import type { ApiResponse } from './_base'

export interface SpuListParams {
  keyword?: string
  status?: string
  type?: string
  channel?: string
  categoryId?: string
  categoryPathPrefix?: string
  page?: number
  pageSize?: number
}

export interface SpuSummary {
  id: string
  name: string
  code: string
  status: string
  type: string
  skuCount?: number
  updatedAt?: string
}

export interface SpuLocaleContent {
  locale: string
  title: string
  description?: string
}

export interface SpuDetail extends SpuSummary {
  categoryId: string
  categoryPath: string
  brandId?: string
  defaultLocale: string
  responsibleUser?: string
  tags?: string[]
  locales?: SpuLocaleContent[]
  currentVersionId?: string
  createdAt: string
  updatedAt: string
}

export interface SpuChannelEntry {
  id: string
  channel: string
  availability: string
  publishAt?: string
  withdrawAt?: string
  auditState?: string
  contentOverride?: Record<string, any>
  lastFeedback?: Record<string, any>
  updatedAt?: string
}

export interface SpuChannelPayload {
  channel: string
  availability: string
  publishAt?: string | null
  withdrawAt?: string | null
  title?: string
  description?: string
}

export interface SpuSubscriptionPlan {
  id: string
  planCode: string
  name: string
  billingCycle: string
  billingValue?: number
  price: number
  currency: string
  trialDays?: number
  autoRenew: boolean
  cancelPolicy: string
  effectScope: string
  status: string
  updatedAt: string
  createdAt: string
}

export interface SubscriptionPlanPayload {
  planCode?: string
  name: string
  billingCycle: string
  billingValue?: number
  price: number
  currency?: string
  trialDays?: number
  autoRenew?: boolean
  cancelPolicy: string
  effectScope?: string
  status?: string
  metadata?: Record<string, any>
}

export interface SpuVersionSummary {
  id: string
  versionNumber: number
  status: string
  submittedBy?: string
  submittedAt?: string
  approvedAt?: string
  createdAt: string
}

export interface SpuVersionDetail extends SpuVersionSummary {
  spuId: string
  diff: VersionDiffEntry[]
  approvals: VersionApproval[]
  payload: Record<string, any>
  rollbackSourceId?: string
}

export interface VersionDiffEntry {
  field: string
  before?: any
  after?: any
}

export interface VersionApproval {
  id: string
  role: string
  status: string
  comment?: string
  actedBy?: string
  actedAt?: string
  slaDueAt?: string
  chainOrder: number
}

export interface SubmitSpuPayload {
  versionId?: string
  comment?: string
}

export interface PublishSpuPayload {
  versionId: string
  channels: string[]
  publishMode?: string
}

export interface WithdrawSpuPayload {
  channels?: string[]
  withdrawAt?: string
  reason?: string
}

export interface DeleteSpuPayload {
  reason: string
}

export interface ReviseSpuPayload {
  reason: string
}

export interface VersionListResponse {
  items: SpuVersionSummary[]
  total: number
  page: number
  pageSize: number
}

export interface VersionListParams {
  status?: string
  page?: number
  pageSize?: number
}

export interface ApprovalActionPayload {
  comment?: string
}

export interface RollbackPayload {
  targetVersionId: string
  reason?: string
}

export interface SpuSkuLink {
  id?: string
  code: string
  name: string
  attributes?: Record<string, any>
  inventoryRef?: string
  pricing: { price: number; currency: string }
  cloneFrom?: string
}

interface SpuListResponse {
  items: SpuSummary[]
  meta: { total: number; page?: number; pageSize?: number }
}

interface SpuSkuListResponse {
  items: SpuSkuLink[]
}

export function useSpuApi() {
  const { baseURL } = useApiClient()
  const basePath = 'admin/product/spus'

  const unwrap = async <T>(promise: Promise<ApiResponse<T>>) => {
    const resp = await promise
    return resp.data
  }

  const listSpus = (params?: SpuListParams, init?: any) =>
    unwrap(apiGet<ApiResponse<SpuListResponse>>(basePath, params, init))

  const getSpu = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<SpuDetail>>(`${basePath}/${id}`, undefined, init))

  const createSpu = (payload: Record<string, any>, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuDetail>>(basePath, payload, init))

  const updateSpu = (id: string, payload: Record<string, any>, init?: any) =>
    unwrap(apiPatch<ApiResponse<SpuDetail>>(`${basePath}/${id}`, payload, init))

  const submitSpu = (id: string, payload?: SubmitSpuPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuDetail>>(`${basePath}/${id}/submit`, payload ?? {}, init))

  const publishSpu = (id: string, payload: PublishSpuPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuDetail>>(`${basePath}/${id}/publish`, payload, init))
  const withdrawSpu = (id: string, payload: WithdrawSpuPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuDetail>>(`${basePath}/${id}/withdraw`, payload, init))
  const deleteSpu = (id: string, payload: DeleteSpuPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuDetail>>(`${basePath}/${id}/delete`, payload, init))
  const reviseSpu = (id: string, payload: ReviseSpuPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuDetail>>(`${basePath}/${id}/revise`, payload, init))

  const listSpuSkus = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<SpuSkuListResponse>>(`${basePath}/${id}/skus`, undefined, init))

  const replaceSpuSkus = (id: string, items: SpuSkuLink[], init?: any) =>
    unwrap(apiPut<ApiResponse<SpuSkuListResponse>>(`${basePath}/${id}/skus`, { items }, init))
  const listSpuVersions = (id: string, params?: VersionListParams, init?: any) =>
    unwrap(apiGet<ApiResponse<VersionListResponse>>(`${basePath}/${id}/versions`, params, init))
  const getSpuVersion = (id: string, versionId: string, init?: any) =>
    unwrap(apiGet<ApiResponse<SpuVersionDetail>>(`${basePath}/${id}/versions/${versionId}`, undefined, init))
  const approveSpuVersion = (id: string, versionId: string, payload?: ApprovalActionPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuVersionDetail>>(`${basePath}/${id}/versions/${versionId}/approve`, payload ?? {}, init))
  const rejectSpuVersion = (id: string, versionId: string, payload?: ApprovalActionPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuVersionDetail>>(`${basePath}/${id}/versions/${versionId}/reject`, payload ?? {}, init))
  const rollbackSpuVersion = (id: string, versionId: string, payload: RollbackPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuVersionDetail>>(`${basePath}/${id}/versions/${versionId}/rollback`, payload, init))
  const listSpuChannels = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<{ items: SpuChannelEntry[] }>>(`${basePath}/${id}/channels`, undefined, init))
  const upsertSpuChannel = (id: string, payload: SpuChannelPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuChannelEntry>>(`${basePath}/${id}/channels`, payload, init))
  const deleteSpuChannel = (id: string, channel: string, init?: any) =>
    unwrap(apiDel<ApiResponse<unknown>>(`${basePath}/${id}/channels/${channel}`, init))
  const listSubscriptionPlans = (id: string, init?: any) =>
    unwrap(apiGet<ApiResponse<{ items: SpuSubscriptionPlan[] }>>(`${basePath}/${id}/subscription-plans`, undefined, init))
  const createSubscriptionPlan = (id: string, payload: SubscriptionPlanPayload, init?: any) =>
    unwrap(apiPost<ApiResponse<SpuSubscriptionPlan>>(`${basePath}/${id}/subscription-plans`, payload, init))
  const updateSubscriptionPlan = (id: string, planId: string, payload: SubscriptionPlanPayload, init?: any) =>
    unwrap(apiPatch<ApiResponse<SpuSubscriptionPlan>>(`${basePath}/${id}/subscription-plans/${planId}`, payload, init))
  const deleteSubscriptionPlan = (id: string, planId: string, init?: any) =>
    unwrap(apiDel<ApiResponse<unknown>>(`${basePath}/${id}/subscription-plans/${planId}`, init))
  const importSpus = (formData: FormData, init?: any) =>
    unwrap(apiPost<ApiResponse<{ taskId: string }>>(`${basePath}/import`, formData, init))
  const exportSpus = (payload: Record<string, any>, init?: any) =>
    unwrap(apiPost<ApiResponse<{ taskId: string }>>(`${basePath}/export`, payload, init))

  return {
    baseURL,
    listSpus,
    createSpu,
    updateSpu,
    getSpu,
    submitSpu,
    publishSpu,
    withdrawSpu,
    deleteSpu,
    reviseSpu,
    listSpuSkus,
    replaceSpuSkus,
    listSpuVersions,
    getSpuVersion,
    approveSpuVersion,
    rejectSpuVersion,
    rollbackSpuVersion,
    listSpuChannels,
    upsertSpuChannel,
    deleteSpuChannel,
    listSubscriptionPlans,
    createSubscriptionPlan,
    updateSubscriptionPlan,
    deleteSubscriptionPlan,
    importSpus,
    exportSpus,
  }
}
