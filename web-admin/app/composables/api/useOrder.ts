import { apiGet, apiPost } from "./_client";
import type { ApiResponse } from "./_base";
import type {
  CancelOrderRequest,
  OrderDetail,
  OrderListResponse,
  OrderSummary,
} from "~/types/order";

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

export type OrderListQuery = {
  page?: number;
  pageSize?: number;
  status?: string;
  customerId?: string;
};

export function useOrderApi() {
  const basePath = "admin/orders";

  return {
    listOrders: (query?: OrderListQuery, init?: any) =>
      unwrap(apiGet<ApiEnvelope<OrderListResponse>>(basePath, query, init)),

    getOrder: (id: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<OrderDetail>>(`${basePath}/${id}`, undefined, init)),

    cancelOrder: (id: string, payload?: CancelOrderRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<OrderSummary>>(`${basePath}/${id}/cancel`, payload || {}, init)),
  };
}
