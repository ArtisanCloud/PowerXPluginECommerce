import { apiDel, apiGet, apiPatch, apiPost, apiPut } from "./_client";
import type { ApiResponse } from "./_base";

export type PricebookStatus = "active" | "archived";
export type PricebookType = "sales" | "purchase";
export type PricebookVersionState = "draft" | "active" | "archived" | "expired";

export type PricebookScopes = {
  channel_ids?: string[];
  customer_group_ids?: string[];
  supplier_ids?: string[];
};

export type PricebookVersion = {
  id: string;
  pricebook_id: string;
  version: number;
  state: PricebookVersionState | string;
  effective_at: string;
  expires_at?: string | null;
  published_at?: string | null;
  published_by?: string | null;
  note?: string | null;
};

export type PricebookItem = {
  id: string;
  pricebook_id: string;
  version_id: string;
  sku_id: string;
  sku_code?: string;
  spu_id?: string;
  spu_name?: string;
  spec_display?: string;
  base_amount_minor?: number | null;
  sale_amount_minor?: number | null;
  msrp_amount_minor?: number | null;
  cost_amount_minor?: number | null;
  min_amount_minor?: number | null;
  max_amount_minor?: number | null;
  tax_included: boolean;
  meta?: Record<string, any> | null;
  created_at?: string;
  updated_at?: string;
};

export type Pricebook = {
  id: string;
  code: string;
  name: string;
  type: PricebookType | string;
  currency: string;
  status: PricebookStatus | string;
  description?: string;
  current_version_id?: string | null;
  current_version?: number | null;
  current_version_state?: string | null;
  scopes?: PricebookScopes | null;
  created_at?: string;
  updated_at?: string;
};

export type PageMeta = { page: number; page_size: number; total: number };

export type PricebookListResponse = {
  items: Pricebook[];
  meta: PageMeta;
};

export type PricebookVersionListResponse = {
  items: PricebookVersion[];
};

export type PricebookItemsListResponse = {
  items: PricebookItem[];
  meta: PageMeta;
};

export type PricebookListParams = {
  keyword?: string;
  type?: string;
  currency?: string;
  status?: string;
  page?: number;
  page_size?: number;
};

export type CreatePricebookPayload = {
  code: string;
  name: string;
  type?: PricebookType;
  currency: string;
  description?: string;
  scopes?: PricebookScopes;
};

export type UpdatePricebookPayload = {
  name?: string;
  description?: string;
  status?: PricebookStatus;
  scopes?: PricebookScopes;
};

export type CreateVersionPayload = {
  copy_from_version_id?: string;
};

export type PublishVersionPayload = {
  effective_at?: string;
  expires_at?: string;
  note?: string;
};

export type ArchiveVersionPayload = {
  note?: string;
};

export type UpsertItemsPayload = {
  items: Array<{
    sku_id: string;
    base_amount_minor?: number | null;
    sale_amount_minor?: number | null;
    msrp_amount_minor?: number | null;
    cost_amount_minor?: number | null;
    min_amount_minor?: number | null;
    max_amount_minor?: number | null;
    tax_included?: boolean | null;
    meta?: Record<string, any> | null;
  }>;
};

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

export function usePricebookApi() {
  const basePath = "admin/pricing/pricebooks";

  return {
    list: (params?: PricebookListParams, init?: any) =>
      unwrap(apiGet<ApiEnvelope<PricebookListResponse>>(basePath, params, init)),

    create: (payload: CreatePricebookPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<Pricebook>>(basePath, payload, init)),

    update: (id: string, payload: UpdatePricebookPayload, init?: any) =>
      unwrap(apiPatch<ApiEnvelope<Pricebook>>(`${basePath}/${id}`, payload, init)),

    remove: (id: string, init?: any) =>
      unwrap(apiDel<ApiEnvelope<{ deleted: boolean }>>(`${basePath}/${id}`, undefined, init)),

    listVersions: (pricebookId: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<PricebookVersionListResponse>>(`${basePath}/${pricebookId}/versions`, undefined, init)),

    createVersion: (pricebookId: string, payload?: CreateVersionPayload, init?: any) =>
      unwrap(apiPost<ApiEnvelope<PricebookVersion>>(`${basePath}/${pricebookId}/versions`, payload || {}, init)),

    publishVersion: (pricebookId: string, versionId: string, payload?: PublishVersionPayload, init?: any) =>
      unwrap(
        apiPost<ApiEnvelope<PricebookVersion>>(
          `${basePath}/${pricebookId}/versions/${versionId}/publish`,
          payload || {},
          init,
        ),
      ),

    archiveVersion: (pricebookId: string, versionId: string, payload?: ArchiveVersionPayload, init?: any) =>
      unwrap(
        apiPost<ApiEnvelope<PricebookVersion>>(
          `${basePath}/${pricebookId}/versions/${versionId}/archive`,
          payload || {},
          init,
        ),
      ),

    listItems: (pricebookId: string, versionId: string, params?: { page?: number; page_size?: number; sku_id?: string }, init?: any) =>
      unwrap(
        apiGet<ApiEnvelope<PricebookItemsListResponse>>(
          `${basePath}/${pricebookId}/versions/${versionId}/items`,
          params,
          init,
        ),
      ),

    upsertItems: (pricebookId: string, versionId: string, payload: UpsertItemsPayload, init?: any) =>
      unwrap(
        apiPut<ApiEnvelope<{ upserted: number; skipped?: string[] }>>(
          `${basePath}/${pricebookId}/versions/${versionId}/items`,
          payload,
          init,
        ),
      ),
  };
}
