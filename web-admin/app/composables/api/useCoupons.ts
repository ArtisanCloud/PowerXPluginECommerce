import { apiGet, apiPatch, apiPost } from "./_client";
import type { ApiResponse } from "./_base";

type ApiEnvelope<T> = ApiResponse<T> & {
  request_id?: string;
  timestamp?: string;
  [key: string]: any;
};

const unwrap = async <T>(promise: Promise<ApiEnvelope<T> | T>): Promise<T> => {
  const resp = await promise;
  if (resp && typeof resp === "object" && "data" in (resp as Record<string, any>)) {
    return (resp as ApiEnvelope<T>).data;
  }
  return resp as T;
};

export type CouponTemplate = {
  id: string;
  tenant_uuid: string;
  code: string;
  name: string;
  coupon_type: string;
  threshold_rule?: Record<string, any>;
  scope_rule?: Record<string, any>;
  stacking_rule?: Record<string, any>;
  refund_rule?: Record<string, any>;
  valid_from: string;
  valid_to: string;
  status: string;
  created_at?: string;
  updated_at?: string;
};

export type CouponTemplateListResponse = {
  items: CouponTemplate[];
  total: number;
  page: number;
  page_size: number;
};

export type CouponAsset = {
  id: string;
  tenant_uuid: string;
  template_id: string;
  user_id: string;
  coupon_code: string;
  status: string;
  reserved_order_id?: string | null;
  reserved_at?: string | null;
  redeemed_at?: string | null;
  valid_from?: string | null;
  valid_to?: string | null;
  created_at?: string;
  updated_at?: string;
};

export type CouponAssetListResponse = {
  items: CouponAsset[];
  total: number;
  page: number;
  page_size: number;
};

export type CouponUsageLog = {
  id: string;
  tenant_uuid: string;
  asset_id: string;
  order_id?: string | null;
  coupon_code?: string;
  user_id?: string;
  action: string;
  action_reason: string;
  idempotency_key: string;
  request_id?: string;
  created_by?: string;
  created_at: string;
};

export type CouponUsageLogListResponse = {
  items: CouponUsageLog[];
  total: number;
  page: number;
  page_size: number;
};

export type CreateCouponTemplatePayload = {
  code: string;
  name: string;
  coupon_type: string;
  threshold_rule?: Record<string, any>;
  scope_rule?: Record<string, any>;
  stacking_rule?: Record<string, any>;
  refund_rule?: Record<string, any>;
  valid_from: string;
  valid_to: string;
  status?: string;
};

export type UpdateCouponTemplatePayload = Partial<CreateCouponTemplatePayload>;

export type IssueCouponsPayload = {
  template_id: string;
  user_ids: string[];
  quantity_per_user?: number;
  valid_from?: string;
  valid_to?: string;
  meta?: Record<string, any>;
};

export type IssueCouponsResponse = {
  issued: number;
  asset_ids: string[];
};

export function useCouponsApi() {
  const basePath = "/admin/coupons";
  return {
    listTemplates: (params?: { keyword?: string; status?: string; page?: number; pageSize?: number }, init?: any) =>
      unwrap(apiGet<ApiEnvelope<CouponTemplateListResponse>>(`${basePath}/templates`, params, init)),

    createTemplate: (payload: CreateCouponTemplatePayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<CouponTemplate>>(`${basePath}/templates`, payload, init)),

    updateTemplate: (id: string, payload: UpdateCouponTemplatePayload, init?: any) =>
      unwrap(apiPatch<ApiEnvelope<CouponTemplate>>(`${basePath}/templates/${id}`, payload, init)),

    issueCoupons: (payload: IssueCouponsPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<IssueCouponsResponse>>(`${basePath}/issues`, payload, init)),

    listAssets: (
      params?: {
        templateId?: string;
        userId?: string;
        status?: string;
        orderId?: string;
        couponCode?: string;
        page?: number;
        pageSize?: number;
      },
      init?: any,
    ) => unwrap(apiGet<ApiEnvelope<CouponAssetListResponse>>(`${basePath}/assets`, params, init)),

    listUsageLogs: (
      params?: {
        assetId?: string;
        orderId?: string;
        action?: string;
        couponCode?: string;
        userId?: string;
        page?: number;
        pageSize?: number;
      },
      init?: any,
    ) => unwrap(apiGet<ApiEnvelope<CouponUsageLogListResponse>>(`${basePath}/usage-logs`, params, init)),
  };
}
