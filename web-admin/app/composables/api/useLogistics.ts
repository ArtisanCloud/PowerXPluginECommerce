import { apiGet, apiPatch, apiPost } from "./_client";
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

export type LogisticsCarrier = {
  id: string;
  name: string;
  code: string;
  type: string;
  status: string;
  contactName: string;
  contactPhone: string;
  capabilities: Record<string, any>;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsTemplate = {
  id: string;
  name: string;
  currency: string;
  status: string;
  version: number;
  channels: any[];
  rules: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsWaybill = {
  id: string;
  orderId: string;
  carrierId: string;
  serviceCode: string;
  waybillNo: string;
  status: string;
  feeAmount: number;
  labelUrl: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsTracking = {
  id: string;
  waybillId: string;
  waybillNo: string;
  eventId: string;
  status: string;
  source: string;
  description: string;
  occurredAt: string;
  payload: Record<string, any>;
  createdAt: string;
};

export type LogisticsWaybillDetail = {
  waybill: LogisticsWaybill;
  tracking: LogisticsTracking[];
};

const normalizeCarrier = (raw: RawRecord): LogisticsCarrier => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  code: String(pick(raw, "code", "code") || ""),
  type: String(pick(raw, "type", "type") || ""),
  status: String(pick(raw, "status", "status") || ""),
  contactName: String(pick(raw, "contactName", "contact_name") || ""),
  contactPhone: String(pick(raw, "contactPhone", "contact_phone") || ""),
  capabilities: (pick(raw, "capabilities", "capabilities") || {}) as Record<string, any>,
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeTemplate = (raw: RawRecord): LogisticsTemplate => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  currency: String(pick(raw, "currency", "currency") || "CNY"),
  status: String(pick(raw, "status", "status") || "draft"),
  version: Number(pick(raw, "version", "version") || 1),
  channels: asArray<any>(pick(raw, "channels", "channels") || []),
  rules: (pick(raw, "rules", "rules") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeWaybill = (raw: RawRecord): LogisticsWaybill => ({
  id: String(pick(raw, "id", "id") || ""),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  serviceCode: String(pick(raw, "serviceCode", "service_code") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  status: String(pick(raw, "status", "status") || "created"),
  feeAmount: Number(pick(raw, "feeAmount", "fee_amount") || 0),
  labelUrl: String(pick(raw, "labelUrl", "label_url") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeTracking = (raw: RawRecord): LogisticsTracking => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  eventId: String(pick(raw, "eventId", "event_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  source: String(pick(raw, "source", "source") || ""),
  description: String(pick(raw, "description", "description") || ""),
  occurredAt: String(pick(raw, "occurredAt", "occurred_at") || ""),
  payload: (pick(raw, "payload", "payload") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

export function useLogisticsApi() {
  const basePath = "/admin/logistics";

  return {
    listCarriers: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/carriers`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCarrier);
    },

    upsertCarrier: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/carriers`, payload, init));
      return normalizeCarrier(raw || {});
    },

    disableCarrier: async (id: string, init?: any) => {
      const raw = await unwrap(apiPatch<ApiEnvelope<RawRecord>>(`${basePath}/carriers/${id}`, {}, init));
      return normalizeCarrier(raw || {});
    },

    testCarrier: (id: string, init?: any) =>
      unwrap(apiPost<ApiEnvelope<{ carrier_id: string; reachable: boolean; message: string }>>(`${basePath}/carriers/${id}/test`, {}, init)),

    listTemplates: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/templates`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeTemplate);
    },

    upsertTemplate: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/templates`, payload, init));
      return normalizeTemplate(raw || {});
    },

    publishTemplate: async (id: string, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/templates/${id}/publish`, {}, init));
      return normalizeTemplate(raw || {});
    },

    listWaybills: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/waybills`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeWaybill);
    },

    createWaybill: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<{ waybill: RawRecord; idempotency_status: string }>>(`${basePath}/waybills`, payload, init));
      return {
        waybill: normalizeWaybill((raw?.waybill || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    getWaybillDetail: async (id: string, init?: any): Promise<LogisticsWaybillDetail> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ waybill: RawRecord; tracking: RawRecord[] }>>(`${basePath}/waybills/${id}`, undefined, init));
      return {
        waybill: normalizeWaybill((raw?.waybill || {}) as RawRecord),
        tracking: asArray<RawRecord>(raw?.tracking).map(normalizeTracking),
      };
    },
  };
}
