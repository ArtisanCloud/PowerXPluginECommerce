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

export type PromotionType = "amount_off" | "percent_off";
export type PromotionStatus = "draft" | "active" | "paused" | "expired";

export type PromotionCampaign = {
  id: string;
  code: string;
  name: string;
  description?: string;
  promotion_type: PromotionType;
  condition_rule: Record<string, any>;
  scope_rule: Record<string, any>;
  action_rule: Record<string, any>;
  stacking_rule: Record<string, any>;
  valid_from: string;
  valid_to: string;
  status: PromotionStatus;
  created_at?: string;
  updated_at?: string;
};

export type PromotionListParams = {
  keyword?: string;
  promotion_type?: PromotionType | "";
  status?: PromotionStatus | "";
  channel?: string;
  page?: number;
  page_size?: number;
};

export type PromotionListResponse = {
  items: PromotionCampaign[];
  total: number;
  page: number;
  page_size: number;
};

export type PromotionPayload = {
  code?: string;
  name: string;
  description?: string;
  promotion_type?: PromotionType;
  condition_rule: Record<string, any>;
  scope_rule: Record<string, any>;
  action_rule: Record<string, any>;
  stacking_rule: Record<string, any>;
  valid_from: string;
  valid_to: string;
  save_action?: "draft" | "activate";
};

export function usePromotionsApi() {
  const basePath = "/admin/promotions";
  return {
    list: (params?: PromotionListParams, init?: any) =>
      unwrap(apiGet<ApiEnvelope<PromotionListResponse>>(basePath, params, init)),

    create: (payload: PromotionPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<PromotionCampaign>>(basePath, payload, init)),

    update: (id: string, payload: Partial<PromotionPayload>, init?: any) =>
      unwrap(apiPatch<ApiEnvelope<PromotionCampaign>>(`${basePath}/${id}`, payload, init)),

    activate: (id: string, init?: any) =>
      unwrap(apiPost<ApiEnvelope<PromotionCampaign>>(`${basePath}/${id}/activate`, {}, init)),

    pause: (id: string, init?: any) =>
      unwrap(apiPost<ApiEnvelope<PromotionCampaign>>(`${basePath}/${id}/pause`, {}, init)),

    clone: (id: string, init?: any) =>
      unwrap(apiPost<ApiEnvelope<PromotionCampaign>>(`${basePath}/${id}/clone`, {}, init)),

    auditLogs: (id: string, params?: { page?: number; page_size?: number }, init?: any) =>
      unwrap(apiGet<ApiEnvelope<{ items: unknown[]; total: number }>>(`${basePath}/${id}/audit-logs`, params, init)),
  };
}
