import { apiGet, apiPut } from "./_client";
import type { ApiResponse } from "./_base";

export type ProductSpecStatus = "active" | "disabled" | "deprecated" | string;

export type ProductSpecOption = {
  id: string;
  group_id: string;
  code: string;
  name: string;
  sort_order: number;
  meta?: Record<string, any> | null;
  status: ProductSpecStatus;
};

export type ProductSpecGroup = {
  id: string;
  code: string;
  name: string;
  sort_order: number;
  required: boolean;
  status: ProductSpecStatus;
  options: ProductSpecOption[];
};

export type ReplaceProductSpecRequest = {
  groups: Array<{
    id?: string;
    code: string;
    name: string;
    sort_order?: number;
    required?: boolean;
    status?: ProductSpecStatus;
    options?: Array<{
      id?: string;
      code: string;
      name: string;
      sort_order?: number;
      meta?: Record<string, any> | null;
      status?: ProductSpecStatus;
    }>;
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

const normalizeSpecOption = (opt: Record<string, any>): ProductSpecOption => ({
  id: String(opt?.id ?? '').trim(),
  group_id: String(opt?.group_id ?? opt?.groupId ?? '').trim(),
  code: String(opt?.code ?? '').trim(),
  name: String(opt?.name ?? '').trim(),
  sort_order: opt?.sort_order ?? opt?.sortOrder ?? 0,
  meta: opt?.meta ?? null,
  status: opt?.status ?? 'active',
})

const normalizeSpecGroup = (group: Record<string, any>): ProductSpecGroup => ({
  id: String(group?.id ?? '').trim(),
  code: String(group?.code ?? '').trim(),
  name: String(group?.name ?? '').trim(),
  sort_order: group?.sort_order ?? group?.sortOrder ?? 0,
  required: Boolean(group?.required),
  status: group?.status ?? 'active',
  options: Array.isArray(group?.options) ? group.options.map((opt: Record<string, any>) => normalizeSpecOption(opt)) : [],
})

export function useProductSpecApi() {
  const basePath = "admin/product/spus";

  return {
    list: (spuId: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<{ groups: ProductSpecGroup[] }>>(`${basePath}/${spuId}/spec-groups`, undefined, init)).then((resp) => ({
        groups: Array.isArray(resp?.groups) ? resp.groups.map((group) => normalizeSpecGroup(group as Record<string, any>)) : [],
      })),

    replace: (spuId: string, payload: ReplaceProductSpecRequest, init?: any) =>
      unwrap(apiPut<ApiEnvelope<{ groups: ProductSpecGroup[] }>>(`${basePath}/${spuId}/spec-groups`, payload, init)).then((resp) => ({
        groups: Array.isArray(resp?.groups) ? resp.groups.map((group) => normalizeSpecGroup(group as Record<string, any>)) : [],
      })),
  };
}
