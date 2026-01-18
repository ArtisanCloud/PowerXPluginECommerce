import { apiGet, apiPost } from "./_client";
import type { ApiResponse } from "./_base";
import type {
  PaymentProvider,
  PaymentRiskEvent,
  PaymentReconciliation,
  PaymentReconciliationItem,
  PaymentSplitResult,
  PaymentTransaction,
  PaymentRefund,
  ManualPaymentReview,
} from "~/types/payments";

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

type RawRecord = Record<string, any>;

const pick = (raw: RawRecord, camel: string, snake: string) =>
  raw[camel] ?? raw[snake];

const normalizeProvider = (raw: RawRecord): PaymentProvider => ({
  id: Number(pick(raw, "id", "id") || 0),
  name: String(pick(raw, "name", "name") || ""),
  type: String(pick(raw, "type", "type") || ""),
  status: String(pick(raw, "status", "status") || ""),
  feeRate: Number(pick(raw, "feeRate", "fee_rate") || 0),
  settlementCycle: String(pick(raw, "settlementCycle", "settlement_cycle") || ""),
  currency: String(pick(raw, "currency", "currency") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeTransaction = (raw: RawRecord): PaymentTransaction => ({
  id: Number(pick(raw, "id", "id") || 0),
  transactionNo: String(pick(raw, "transactionNo", "transaction_no") || ""),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  orderNo: String(pick(raw, "orderNo", "order_no") || ""),
  providerId: Number(pick(raw, "providerId", "provider_id") || 0),
  payMethod: String(pick(raw, "payMethod", "pay_method") || ""),
  amountTotal: Number(pick(raw, "amountTotal", "amount_total") || 0),
  amountCurrency: String(pick(raw, "amountCurrency", "amount_currency") || ""),
  feeAmount: Number(pick(raw, "feeAmount", "fee_amount") || 0),
  status: String(pick(raw, "status", "status") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  completedAt: pick(raw, "completedAt", "completed_at") || null,
  failureReason: String(pick(raw, "failureReason", "failure_reason") || ""),
});

const normalizeRefund = (raw: RawRecord): PaymentRefund => ({
  id: Number(pick(raw, "id", "id") || 0),
  refundNo: String(pick(raw, "refundNo", "refund_no") || ""),
  refundAmount: Number(pick(raw, "refundAmount", "refund_amount") || 0),
  refundCurrency: String(pick(raw, "refundCurrency", "refund_currency") || ""),
  status: String(pick(raw, "status", "status") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  completedAt: pick(raw, "completedAt", "completed_at") || null,
});

const normalizeRiskEvent = (raw: RawRecord): PaymentRiskEvent => ({
  id: Number(pick(raw, "id", "id") || 0),
  riskType: String(pick(raw, "riskType", "risk_type") || ""),
  riskScore: Number(pick(raw, "riskScore", "risk_score") || 0),
  action: String(pick(raw, "action", "action") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  resolvedAt: pick(raw, "resolvedAt", "resolved_at") || null,
});

const normalizeSplitResult = (raw: RawRecord): PaymentSplitResult => ({
  id: Number(pick(raw, "id", "id") || 0),
  ruleId: Number(pick(raw, "ruleId", "rule_id") || 0),
  participant: String(pick(raw, "participant", "participant") || ""),
  amount: Number(pick(raw, "amount", "amount") || 0),
  status: String(pick(raw, "status", "status") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeReconciliation = (raw: RawRecord): PaymentReconciliation => ({
  id: Number(pick(raw, "id", "id") || 0),
  periodType: String(pick(raw, "periodType", "period_type") || ""),
  periodStart: String(pick(raw, "periodStart", "period_start") || ""),
  periodEnd: String(pick(raw, "periodEnd", "period_end") || ""),
  diffCount: Number(pick(raw, "diffCount", "diff_count") || 0),
  diffTotalAmount: Number(pick(raw, "diffTotalAmount", "diff_total_amount") || 0),
  status: String(pick(raw, "status", "status") || ""),
  processedBy: String(pick(raw, "processedBy", "processed_by") || ""),
  processedAt: pick(raw, "processedAt", "processed_at") || null,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeReconciliationItem = (raw: RawRecord): PaymentReconciliationItem => ({
  id: Number(pick(raw, "id", "id") || 0),
  reconciliationId: Number(pick(raw, "reconciliationId", "reconciliation_id") || 0),
  transactionId: Number(pick(raw, "transactionId", "transaction_id") || 0),
  diffType: String(pick(raw, "diffType", "diff_type") || ""),
  diffAmount: Number(pick(raw, "diffAmount", "diff_amount") || 0),
  resolution: String(pick(raw, "resolution", "resolution") || ""),
  resolvedBy: String(pick(raw, "resolvedBy", "resolved_by") || ""),
  resolvedAt: pick(raw, "resolvedAt", "resolved_at") || null,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeManualReview = (raw: RawRecord): ManualPaymentReview => ({
  id: Number(pick(raw, "id", "id") || 0),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  orderNo: String(pick(raw, "orderNo", "order_no") || ""),
  transactionId: pick(raw, "transactionId", "transaction_id") || null,
  providerId: Number(pick(raw, "providerId", "provider_id") || 0),
  payMethod: String(pick(raw, "payMethod", "pay_method") || ""),
  amountMinor: Number(pick(raw, "amountMinor", "amount_minor") || 0),
  currency: String(pick(raw, "currency", "currency") || ""),
  status: String(pick(raw, "status", "status") || ""),
  submittedBy: String(pick(raw, "submittedBy", "submitted_by") || ""),
  submittedAt: String(pick(raw, "submittedAt", "submitted_at") || ""),
  reviewedBy: String(pick(raw, "reviewedBy", "reviewed_by") || ""),
  reviewedAt: pick(raw, "reviewedAt", "reviewed_at") || null,
  reviewReason: String(pick(raw, "reviewReason", "review_reason") || ""),
  proofNo: String(pick(raw, "proofNo", "proof_no") || ""),
  note: String(pick(raw, "note", "note") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

export type ProviderListQuery = {
  keyword?: string;
  type?: string;
  status?: string;
};

export type TransactionListQuery = {
  status?: string;
  providerId?: number;
  from?: string;
  to?: string;
};

export type RefundCreateRequest = {
  amountMinor: number;
  reason?: string;
};

export type ManualPaymentCreateRequest = {
  orderId: string;
  amountMinor: number;
  currency?: string;
  payMethod: string;
  providerId?: number;
  proofNo?: string;
  note?: string;
};

export type ManualPaymentReviewRequest = {
  reason?: string;
};

export type ReconciliationCreateRequest = {
  periodType: string;
  periodStart: string;
  periodEnd: string;
  items: Array<{
    transactionId: number;
    diffType: string;
    diffAmount: number;
  }>;
};

export type ReconciliationResolveRequest = {
  resolution?: string;
};

export function usePaymentsApi() {
  const basePath = "/admin/payments";

  return {
    listProviders: async (query?: ProviderListQuery, init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/providers`, query, init));
      return Array.isArray(raw) ? raw.map(normalizeProvider) : [];
    },
    listTransactions: async (query?: TransactionListQuery, init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/transactions`, query, init));
      return Array.isArray(raw) ? raw.map(normalizeTransaction) : [];
    },
    getTransaction: async (id: number | string, init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord>>(`${basePath}/transactions/${id}`, undefined, init));
      return normalizeTransaction(raw || {});
    },
    createRefund: async (transactionId: number | string, payload: RefundCreateRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/transactions/${transactionId}/refund`, payload, init),
      );
      return normalizeRefund(raw || {});
    },
    listRiskEvents: async (transactionId?: number, init?: any) => {
      const query = transactionId ? { transactionId } : undefined;
      const raw = await unwrap(
        apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/risk-events`, query, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeRiskEvent) : [];
    },
    listSplitResults: async (transactionId?: number, init?: any) => {
      const query = transactionId ? { transactionId } : undefined;
      const raw = await unwrap(
        apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/split-results`, query, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeSplitResult) : [];
    },
    listReconciliations: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/reconciliations`, undefined, init));
      return Array.isArray(raw) ? raw.map(normalizeReconciliation) : [];
    },
    createReconciliation: async (payload: ReconciliationCreateRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/reconciliations`, payload, init),
      );
      return normalizeReconciliation(raw || {});
    },
    listReconciliationItems: async (reconciliationId: number | string, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/reconciliations/${reconciliationId}/items`, undefined, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeReconciliationItem) : [];
    },
    resolveReconciliationItem: async (
      reconciliationId: number | string,
      itemId: number | string,
      payload?: ReconciliationResolveRequest,
      init?: any,
    ) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(
          `${basePath}/reconciliations/${reconciliationId}/items/${itemId}/resolve`,
          payload || {},
          init,
        ),
      );
      return normalizeReconciliationItem(raw || {});
    },
    listManualPayments: async (orderId?: string, init?: any) => {
      const query = orderId ? { orderId } : undefined;
      const raw = await unwrap(
        apiGet<ApiEnvelope<RawRecord[]>>(`${basePath}/manual-payments`, query, init),
      );
      return Array.isArray(raw) ? raw.map(normalizeManualReview) : [];
    },
    createManualPayment: async (payload: ManualPaymentCreateRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/manual-payments`, payload, init),
      );
      return normalizeManualReview(raw || {});
    },
    approveManualPayment: async (reviewId: number | string, payload?: ManualPaymentReviewRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/manual-payments/${reviewId}/approve`, payload || {}, init),
      );
      return normalizeManualReview(raw || {});
    },
    rejectManualPayment: async (reviewId: number | string, payload: ManualPaymentReviewRequest, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/manual-payments/${reviewId}/reject`, payload, init),
      );
      return normalizeManualReview(raw || {});
    },
  };
}
