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

export function useProductSpecApi() {
  const basePath = "admin/product/spus";

  return {
    list: (spuId: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<{ groups: ProductSpecGroup[] }>>(`${basePath}/${spuId}/spec-groups`, undefined, init)),

    replace: (spuId: string, payload: ReplaceProductSpecRequest, init?: any) =>
      unwrap(apiPut<ApiEnvelope<{ groups: ProductSpecGroup[] }>>(`${basePath}/${spuId}/spec-groups`, payload, init)),
  };
}

