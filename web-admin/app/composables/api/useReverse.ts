import { apiGet, apiPost } from "./_client";
import type { ApiResponse } from "./_base";

type ApiEnvelope<T> = ApiResponse<T> & {
  request_id?: string;
  timestamp?: string;
  [key: string]: any;
};

type RawRecord = Record<string, any>;

const unwrap = async <T>(promise: Promise<ApiEnvelope<T> | T>): Promise<T> => {
  const resp = await promise;
  if (resp && typeof resp === "object" && "data" in (resp as Record<string, any>)) {
    return (resp as ApiEnvelope<T>).data;
  }
  return resp as T;
};

const pick = (raw: RawRecord, camel: string, snake: string) => raw[camel] ?? raw[snake];
const asArray = <T>(value: unknown): T[] => (Array.isArray(value) ? (value as T[]) : []);

export type ReverseWaybill = {
  id: string;
  orderId: string;
  afterSaleId: string;
  waybillNo: string;
  status: string;
  inspectionResult: string;
  disposition: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type ReverseTracking = {
  id: string;
  waybillId: string;
  status: string;
  description: string;
  occurredAt: string;
  payload: Record<string, any>;
  createdAt: string;
};

export type ReverseWarehouseResult = {
  id: string;
  waybillId: string;
  result: string;
  operatorId: string;
  notes: string;
  metadata: Record<string, any>;
  createdAt: string;
};

export type ReverseWaybillDetail = {
  waybill: ReverseWaybill;
  tracking: ReverseTracking[];
  warehouseResults: ReverseWarehouseResult[];
  compensation: Record<string, any>;
};

export type ReverseInspectionRule = {
  id: string;
  name: string;
  priority: number;
  enabled: boolean;
  condition: Record<string, any>;
  decision: string;
  recommendation: string;
  notes: string;
  createdAt: string;
  updatedAt: string;
};

export type ReverseInspectionDecision = {
  waybillId: string;
  rule: ReverseInspectionRule | null;
  decision: string;
  recommendation: string;
  reason: string;
  idempotencyState: string;
};

const normalizeWaybill = (raw: RawRecord): ReverseWaybill => ({
  id: String(pick(raw, "id", "id") || ""),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  afterSaleId: String(pick(raw, "afterSaleId", "after_sale_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  status: String(pick(raw, "status", "status") || "created"),
  inspectionResult: String(pick(raw, "inspectionResult", "inspection_result") || ""),
  disposition: String(pick(raw, "disposition", "disposition") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeTracking = (raw: RawRecord): ReverseTracking => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  description: String(pick(raw, "description", "description") || ""),
  occurredAt: String(pick(raw, "occurredAt", "occurred_at") || ""),
  payload: (pick(raw, "payload", "payload") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeWarehouseResult = (raw: RawRecord): ReverseWarehouseResult => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  result: String(pick(raw, "result", "result") || ""),
  operatorId: String(pick(raw, "operatorId", "operator_id") || ""),
  notes: String(pick(raw, "notes", "notes") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeInspectionRule = (raw: RawRecord): ReverseInspectionRule => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  priority: Number(pick(raw, "priority", "priority") || 100),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  condition: (pick(raw, "condition", "condition_json") || {}) as Record<string, any>,
  decision: String(pick(raw, "decision", "decision") || ""),
  recommendation: String(pick(raw, "recommendation", "recommendation") || ""),
  notes: String(pick(raw, "notes", "notes") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeInspectionDecision = (raw: RawRecord): ReverseInspectionDecision => ({
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  rule: raw?.rule ? normalizeInspectionRule(raw.rule as RawRecord) : null,
  decision: String(pick(raw, "decision", "decision") || ""),
  recommendation: String(pick(raw, "recommendation", "recommendation") || ""),
  reason: String(pick(raw, "reason", "reason") || ""),
  idempotencyState: String(pick(raw, "idempotencyState", "idempotency_state") || ""),
});

export function useReverseApi() {
  const basePath = "/admin/reverse";

  return {
    listWaybills: async (init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/waybills`, undefined, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeWaybill);
    },

    createWaybill: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/waybills`, payload, init),
      );
      return normalizeWaybill(raw || {});
    },

    getWaybillDetail: async (id: string, init?: any): Promise<ReverseWaybillDetail> => {
      const raw = await unwrap(
        apiGet<
          ApiEnvelope<{
            waybill: RawRecord;
            tracking: RawRecord[];
            warehouse_results: RawRecord[];
            compensation: Record<string, any>;
          }>
        >(`${basePath}/waybills/${id}`, undefined, init),
      );
      return {
        waybill: normalizeWaybill((raw?.waybill || {}) as RawRecord),
        tracking: asArray<RawRecord>(raw?.tracking).map(normalizeTracking),
        warehouseResults: asArray<RawRecord>(raw?.warehouse_results).map(normalizeWarehouseResult),
        compensation: (raw?.compensation || {}) as Record<string, any>,
      };
    },

    appendTracking: async (id: string, payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/waybills/${id}/track`, payload, init),
      );
      return normalizeTracking(raw || {});
    },

    recordWarehouseResult: async (id: string, payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ warehouse_result: RawRecord; waybill: RawRecord }>>(
          `${basePath}/waybills/${id}/warehouse-result`,
          payload,
          init,
        ),
      );
      return {
        warehouseResult: normalizeWarehouseResult((raw?.warehouse_result || {}) as RawRecord),
        waybill: normalizeWaybill((raw?.waybill || {}) as RawRecord),
      };
    },

    listInspectionRules: async (init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/inspection/rules`, undefined, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeInspectionRule);
    },

    createInspectionRule: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/inspection/rules`, payload, init),
      );
      return normalizeInspectionRule(raw || {});
    },

    evaluateInspection: async (waybillID: string, payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(
          `${basePath}/waybills/${waybillID}/inspection`,
          payload,
          init,
        ),
      );
      return normalizeInspectionDecision(raw || {});
    },
  };
}
