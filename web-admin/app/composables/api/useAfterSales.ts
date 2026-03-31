import { apiGet, apiPost } from './_client'
import type { ApiResponse } from './_base'

export type AfterSaleCaseType = 'refund_only' | 'return_refund' | 'exchange'

export interface AfterSaleSummary {
  id: string
  caseNo: string
  orderId: string
  orderItemId: string
  customerId?: string
  caseType: AfterSaleCaseType | string
  status: string
  reasonCode?: string
  reasonDetail?: string
  requestedAmountMinor?: number
  currency?: string
  createdAt?: string
}

export interface AfterSaleTimelineEvent {
  action: string
  fromStatus?: string
  toStatus?: string
  operatorType?: string
  operatorId?: string
  note?: string
  createdAt?: string
}

export interface AfterSaleDetail {
  case: AfterSaleSummary
  timeline: AfterSaleTimelineEvent[]
}

export interface AfterSaleListResponse {
  items: AfterSaleSummary[]
  meta?: {
    total?: number
    page?: number
    pageSize?: number
  }
}

export interface MiniAppAfterSaleCreateRequest {
  orderId: string
  orderItemId: string
  caseType: AfterSaleCaseType
  reasonCode: string
  reasonDetail?: string
  requestedQty?: number
  requestedAmountMinor?: number
  currency?: string
  evidences?: Array<{
    evidenceType: 'image' | 'video' | 'text' | 'other'
    contentRef: string
    description?: string
  }>
}

export interface CaseDecisionRequest {
  reasonCode?: string
  note?: string
}

export interface ReverseLogisticsLinkRequest {
  reverseWaybillId?: string
  reverseWaybillNo: string
  carrierCode?: string
  note?: string
}

export interface AfterSaleDashboardSnapshot {
  pendingCount: number
  processingCount: number
  completedCount: number
  rejectedCount: number
}

const toInt = (value: unknown): number => {
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

const normalizeSummary = (item: Record<string, any>): AfterSaleSummary => ({
  id: String(item?.id ?? '').trim(),
  caseNo: String(item?.caseNo ?? item?.case_no ?? '').trim(),
  orderId: String(item?.orderId ?? item?.order_id ?? '').trim(),
  orderItemId: String(item?.orderItemId ?? item?.order_item_id ?? '').trim(),
  customerId: String(item?.customerId ?? item?.customer_id ?? '').trim(),
  caseType: String(item?.caseType ?? item?.case_type ?? '').trim(),
  status: String(item?.status ?? '').trim(),
  reasonCode: String(item?.reasonCode ?? item?.reason_code ?? '').trim(),
  reasonDetail: String(item?.reasonDetail ?? item?.reason_detail ?? '').trim(),
  requestedAmountMinor: item?.requestedAmountMinor ?? item?.requested_amount_minor,
  currency: item?.currency,
  createdAt: item?.createdAt ?? item?.created_at,
})

const normalizeList = (data: any): AfterSaleListResponse => ({
  items: Array.isArray(data?.items) ? data.items.map((item: Record<string, any>) => normalizeSummary(item)) : [],
  meta: {
    total: toInt(data?.meta?.total),
    page: toInt(data?.meta?.page),
    pageSize: toInt(data?.meta?.pageSize ?? data?.meta?.page_size),
  },
})

const normalizeDetail = (data: any): AfterSaleDetail => ({
  case: normalizeSummary((data?.case ?? {}) as Record<string, any>),
  timeline: Array.isArray(data?.timeline) ? data.timeline.map((event: Record<string, any>) => ({
    action: String(event?.action ?? '').trim(),
    fromStatus: event?.fromStatus ?? event?.from_status,
    toStatus: event?.toStatus ?? event?.to_status,
    operatorType: event?.operatorType ?? event?.operator_type,
    operatorId: event?.operatorId ?? event?.operator_id,
    note: event?.note,
    createdAt: event?.createdAt ?? event?.created_at,
  })) : [],
})

const unwrap = async <T>(promise: Promise<T | ApiResponse<T>>) => {
  const resp = await promise
  if (resp && typeof resp === 'object' && Object.prototype.hasOwnProperty.call(resp as any, 'data')) {
    return (resp as ApiResponse<T>).data
  }
  return resp as T
}

export const useAfterSalesApi = () => {
  const miniBase = 'mini-app/after-sales'
  const adminBase = 'admin/after-sales/cases'

  return {
    listMiniAppCases: (params?: { status?: string; page?: number; pageSize?: number }) =>
      unwrap(apiGet<AfterSaleListResponse>(miniBase, params)).then(normalizeList),

    createMiniAppCase: (payload: MiniAppAfterSaleCreateRequest) =>
      unwrap(apiPost<AfterSaleDetail>(miniBase, payload)).then(normalizeDetail),

    getMiniAppCase: (id: string) =>
      unwrap(apiGet<AfterSaleDetail>(`${miniBase}/${id}`)).then(normalizeDetail),

    listAdminCases: (params?: { status?: string; caseType?: string; keyword?: string; page?: number; pageSize?: number }) =>
      unwrap(apiGet<AfterSaleListResponse>(adminBase, params)).then(normalizeList),

    getAdminCase: (id: string) =>
      unwrap(apiGet<AfterSaleDetail>(`${adminBase}/${id}`)).then(normalizeDetail),

    acceptCase: (id: string) =>
      unwrap(apiPost(`${adminBase}/${id}/accept`, {})),

    reviewCase: (id: string, payload?: CaseDecisionRequest) =>
      unwrap(apiPost(`${adminBase}/${id}/review`, payload ?? {})),

    approveCase: (id: string, payload?: CaseDecisionRequest) =>
      unwrap(apiPost(`${adminBase}/${id}/approve`, payload ?? {})),

    rejectCase: (id: string, payload: CaseDecisionRequest) =>
      unwrap(apiPost(`${adminBase}/${id}/reject`, payload)),

    completeCase: (id: string) =>
      unwrap(apiPost(`${adminBase}/${id}/complete`, {})),

    closeCase: (id: string) =>
      unwrap(apiPost(`${adminBase}/${id}/close`, {})),

    linkReverseLogistics: (id: string, payload: ReverseLogisticsLinkRequest) =>
      unwrap(apiPost(`${adminBase}/${id}/reverse-logistics`, payload)),

    getDashboard: () =>
      unwrap(apiGet<AfterSaleDashboardSnapshot>('admin/after-sales/dashboard')),
  }
}
