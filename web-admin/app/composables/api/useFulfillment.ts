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

export type FulfillmentWaveStrategy = {
  id: string;
  name: string;
  warehouseId: string;
  carrierCode: string;
  timeWindow: string;
  priorityBand: string;
  maxTasksPerWave: number;
  enabled: boolean;
  rules: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type FulfillmentWaveStrategyPreviewGroup = {
  groupKey: string;
  warehouseId: string;
  carrierCode: string;
  timeSlot: string;
  priorityBand: string;
  taskIds: string[];
};

export type FulfillmentWaveStrategyPreview = {
  strategyId: string;
  strategyName: string;
  totalCandidate: number;
  groups: FulfillmentWaveStrategyPreviewGroup[];
  skippedTaskIds: string[];
};

export type FulfillmentOutbound = {
  id: string;
  taskId: string;
  orderId: string;
  warehouseId: string;
  waybillId: string;
  status: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
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

const normalizeWaveStrategy = (raw: RawRecord): FulfillmentWaveStrategy => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  carrierCode: String(pick(raw, "carrierCode", "carrier_code") || ""),
  timeWindow: String(pick(raw, "timeWindow", "time_window") || ""),
  priorityBand: String(pick(raw, "priorityBand", "priority_band") || ""),
  maxTasksPerWave: Number(pick(raw, "maxTasksPerWave", "max_tasks_per_wave") || 50),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  rules: (pick(raw, "rules", "rules") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeWaveStrategyPreviewGroup = (
  raw: RawRecord,
): FulfillmentWaveStrategyPreviewGroup => ({
  groupKey: String(pick(raw, "groupKey", "group_key") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  carrierCode: String(pick(raw, "carrierCode", "carrier_code") || ""),
  timeSlot: String(pick(raw, "timeSlot", "time_slot") || ""),
  priorityBand: String(pick(raw, "priorityBand", "priority_band") || ""),
  taskIds: asArray<string>(pick(raw, "taskIds", "task_ids") || []),
});

const normalizeOutbound = (raw: RawRecord): FulfillmentOutbound => ({
  id: String(pick(raw, "id", "id") || ""),
  taskId: String(pick(raw, "taskId", "task_id") || ""),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  status: String(pick(raw, "status", "status") || "reserved"),
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

    listWaveStrategies: async (init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/wave-strategies`, undefined, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeWaveStrategy);
    },

    createWaveStrategy: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/wave-strategies`, payload, init),
      );
      return normalizeWaveStrategy(raw || {});
    },

    previewWaveStrategy: async (strategyId: string, init?: any): Promise<FulfillmentWaveStrategyPreview> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(
          `${basePath}/wave-strategies/preview`,
          { strategy_id: strategyId },
          init,
        ),
      );
      return {
        strategyId: String(pick(raw || {}, "strategyId", "strategy_id") || ""),
        strategyName: String(pick(raw || {}, "strategyName", "strategy_name") || ""),
        totalCandidate: Number(pick(raw || {}, "totalCandidate", "total_candidate") || 0),
        groups: asArray<RawRecord>(pick(raw || {}, "groups", "groups") || []).map(
          normalizeWaveStrategyPreviewGroup,
        ),
        skippedTaskIds: asArray<string>(pick(raw || {}, "skippedTaskIds", "skipped_task_ids") || []),
      };
    },

    listOutbounds: async (query?: { status?: string }, init?: any): Promise<FulfillmentOutbound[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/warehouse/outbounds`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeOutbound);
    },

    createOutbound: async (payload: Record<string, any>, init?: any): Promise<FulfillmentOutbound> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/warehouse/outbounds`, payload, init),
      );
      return normalizeOutbound((raw || {}) as RawRecord);
    },

    executeOutbound: async (id: string, payload?: Record<string, any>, init?: any) => {
      return unwrap(
        apiPost<ApiEnvelope<{ outbound: RawRecord; task: RawRecord }>>(
          `${basePath}/warehouse/outbounds/${id}/execute`,
          payload || {},
          init,
        ),
      );
    },

    rollbackOutbound: async (id: string, payload?: Record<string, any>, init?: any) => {
      return unwrap(
        apiPost<ApiEnvelope<{ outbound: RawRecord; task: RawRecord }>>(
          `${basePath}/warehouse/outbounds/${id}/rollback`,
          payload || {},
          init,
        ),
      );
    },
  };
}
