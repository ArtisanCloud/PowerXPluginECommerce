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

export type FulfillmentTask = {
  id: string;
  orderId: string;
  warehouseId: string;
  status: string;
  assignedTo: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type FulfillmentException = {
  id: string;
  taskId: string;
  waybillId: string;
  type: string;
  status: string;
  reason: string;
  firstActionAt: string;
  escalatedAt: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type FulfillmentWave = {
  id: string;
  name: string;
  warehouseId: string;
  status: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type FulfillmentWaveTaskLink = {
  id: string;
  waveId: string;
  taskId: string;
  result: string;
  message: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type FulfillmentWaveDetail = {
  wave: FulfillmentWave;
  tasks: FulfillmentWaveTaskLink[];
};

const normalizeTask = (raw: RawRecord): FulfillmentTask => ({
  id: String(pick(raw, "id", "id") || ""),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  status: String(pick(raw, "status", "status") || "pending"),
  assignedTo: String(pick(raw, "assignedTo", "assigned_to") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeException = (raw: RawRecord): FulfillmentException => ({
  id: String(pick(raw, "id", "id") || ""),
  taskId: String(pick(raw, "taskId", "task_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  type: String(pick(raw, "type", "type") || "other"),
  status: String(pick(raw, "status", "status") || "open"),
  reason: String(pick(raw, "reason", "reason") || ""),
  firstActionAt: String(pick(raw, "firstActionAt", "first_action_at") || ""),
  escalatedAt: String(pick(raw, "escalatedAt", "escalated_at") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeWave = (raw: RawRecord): FulfillmentWave => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  status: String(pick(raw, "status", "status") || "pending"),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeWaveTaskLink = (raw: RawRecord): FulfillmentWaveTaskLink => ({
  id: String(pick(raw, "id", "id") || ""),
  waveId: String(pick(raw, "waveId", "wave_id") || ""),
  taskId: String(pick(raw, "taskId", "task_id") || ""),
  result: String(pick(raw, "result", "result") || "pending"),
  message: String(pick(raw, "message", "message") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

export function useFulfillmentApi() {
  const basePath = "/admin/fulfillment";

  return {
    listTasks: async (query?: { status?: string }, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/tasks`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeTask);
    },

    createTask: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tasks`, payload, init),
      );
      return normalizeTask(raw || {});
    },

    completeTask: async (id: string, payload?: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPatch<ApiEnvelope<RawRecord>>(
          `${basePath}/tasks/${id}/complete`,
          payload || {},
          init,
        ),
      );
      return normalizeTask(raw || {});
    },

    listExceptions: async (query?: { status?: string }, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/exceptions`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeException);
    },

    reportException: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/exceptions`, payload, init),
      );
      return normalizeException(raw || {});
    },

    listWaves: async (query?: { status?: string }, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/waves`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeWave);
    },

    createWave: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ wave: RawRecord; tasks: RawRecord[] }>>(`${basePath}/waves`, payload, init),
      );
      return {
        wave: normalizeWave((raw?.wave || {}) as RawRecord),
        tasks: asArray<RawRecord>(raw?.tasks).map(normalizeWaveTaskLink),
      };
    },

    getWaveDetail: async (id: string, init?: any): Promise<FulfillmentWaveDetail> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ wave: RawRecord; tasks: RawRecord[] }>>(`${basePath}/waves/${id}`, undefined, init),
      );
      return {
        wave: normalizeWave((raw?.wave || {}) as RawRecord),
        tasks: asArray<RawRecord>(raw?.tasks).map(normalizeWaveTaskLink),
      };
    },

    advanceWave: async (id: string, payload: Record<string, any>, init?: any) => {
      return unwrap(
        apiPatch<ApiEnvelope<{
          wave_id: string;
          wave_status: string;
          success: number;
          failed: number;
          task_results: RawRecord[];
        }>>(`${basePath}/waves/${id}/advance`, payload, init),
      );
    },

    reassignWaveTask: async (waveId: string, taskId: string, payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPatch<ApiEnvelope<RawRecord>>(`${basePath}/waves/${waveId}/tasks/${taskId}/reassign`, payload, init),
      );
      return normalizeTask(raw || {});
    },
  };
}
