import { apiGet, apiPatch, apiPost } from "./_client";
import type { ApiResponse } from "./_base";
import type {
  CancelOrderRequest,
  CreateOrderRequest,
  OrderDetail,
  OrderBenefitReview,
  OrderListResponse,
  OrderSummary,
  ShippingAddress,
} from "~/types/order";

type ApiEnvelope<T> = ApiResponse<T> & {
  request_id?: string;
  timestamp?: string;
  [key: string]: any;
};

type RawRecord = Record<string, any>;

const pick = (raw: RawRecord, camel: string, snake: string) =>
  raw[camel] ?? raw[snake];

const normalizeBenefitReview = (raw: RawRecord): OrderBenefitReview => ({
  id: Number(pick(raw, "id", "id") || 0),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  orderNo: String(pick(raw, "orderNo", "order_no") || ""),
  benefitType: String(pick(raw, "benefitType", "benefit_type") || ""),
  benefitCode: String(pick(raw, "benefitCode", "benefit_code") || ""),
  valueType: String(pick(raw, "valueType", "value_type") || ""),
  value: Number(pick(raw, "value", "value") || 0),
  amountMinor: Number(pick(raw, "amountMinor", "amount_minor") || 0),
  currency: String(pick(raw, "currency", "currency") || ""),
  stackingAllowed: Boolean(pick(raw, "stackingAllowed", "stacking_allowed")),
  status: String(pick(raw, "status", "status") || ""),
  submittedBy: String(pick(raw, "submittedBy", "submitted_by") || ""),
  submittedAt: String(pick(raw, "submittedAt", "submitted_at") || ""),
  reviewedBy: String(pick(raw, "reviewedBy", "reviewed_by") || ""),
  reviewedAt: pick(raw, "reviewedAt", "reviewed_at") || null,
  reviewReason: String(pick(raw, "reviewReason", "review_reason") || ""),
  note: String(pick(raw, "note", "note") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

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
  orderNo?: string;
  customerId?: string;
};

export type BenefitReviewCreateRequest = {
  benefitType: "coupon" | "giftcard";
  benefitCode: string;
  valueType: "amount" | "percent" | "balance";
  value: number;
  currency?: string;
  stackingAllowed: boolean;
  note?: string;
};

export type BenefitReviewDecisionRequest = {
  reviewIds: Array<number | string>;
  reason?: string;
};

export function useOrderApi() {
  const basePath = "/admin/orders";

  return {
    listOrders: (query?: OrderListQuery, init?: any) =>
      unwrap(apiGet<ApiEnvelope<OrderListResponse>>(basePath, query, init)),

    getOrder: (id: string, init?: any) =>
      unwrap(apiGet<ApiEnvelope<OrderDetail>>(`${basePath}/${id}`, undefined, init)),

    cancelOrder: (id: string, payload?: CancelOrderRequest, init?: any) =>
      unwrap(apiPost<ApiEnvelope<OrderSummary>>(`${basePath}/${id}/cancel`, payload || {}, init)),

    updateShippingAddress: (id: string, shippingAddress: ShippingAddress, init?: any) =>
      unwrap(
        apiPatch<ApiEnvelope<OrderSummary>>(
          `${basePath}/${id}/shipping-address`,
          { shippingAddress },
          init,
        ),
      ),

    listBenefitReviews: async (orderId: string, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/${orderId}/benefit-reviews`, undefined, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeBenefitReview) : [];
    },

    createBenefitReview: async (orderId: string, payload: BenefitReviewCreateRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/${orderId}/benefit-reviews`, payload, init),
      );
      return normalizeBenefitReview(raw || {});
    },

    approveBenefitReviews: async (payload: BenefitReviewDecisionRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord[]>>(`${basePath}/benefit-reviews/approve`, payload, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeBenefitReview) : [];
    },

    rejectBenefitReviews: async (payload: BenefitReviewDecisionRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord[]>>(`${basePath}/benefit-reviews/reject`, payload, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeBenefitReview) : [];
    },

    searchBenefitCodes: async (benefitType: "coupon" | "giftcard", keyword: string, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<Array<{ label: string; value: string }>>>(
          `${basePath}/benefit-codes`,
          { type: benefitType, q: keyword },
          init,
        ),
      );
      return Array.isArray(raw) ? raw : [];
    },

    createOrder: (payload: CreateOrderRequest, idempotencyKey: string, init?: any) =>
      unwrap(
        apiPost<ApiEnvelope<OrderSummary>>(basePath, payload, {
          ...(init || {}),
          headers: {
            ...(init?.headers || {}),
            "Idempotency-Key": idempotencyKey,
          },
        }),
      ),
  };
}
