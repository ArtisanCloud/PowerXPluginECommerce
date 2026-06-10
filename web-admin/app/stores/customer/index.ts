import { defineStore } from "pinia";
import { useCustomerApi } from "~/composables/api";
import { useApiClient } from "~/composables/api/_client";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import { useMembershipInsights } from "~/composables/useMembershipInsights";
import { useCustomerBulkActions } from "~/composables/useCustomerBulkActions";
import { useWsBusClient, type WsBusEvent } from "~/composables/useWsBusClient";
import type {
	BulkReminderPayload,
	BulkTask,
	Customer,
	CustomerCreatePayload,
	CustomerListFilters,
	CustomerUpdatePayload,
	MembershipFilters,
	MembershipInsight,
	MembershipReminderState,
	MembershipSnapshot,
	MembershipSegments,
	MembershipStats,
	SavedView,
} from "~/types/customer";

const SAVED_VIEWS_KEY = "px_customer_saved_views";

const TASK_PROGRESS_TOPICS = [
  "task.progress",
  "powerx.task.progress.v1",
  "worker.task.updated",
] as const;

const TERMINAL_TASK_STATUSES = new Set(["success", "failed", "error", "cancelled"]);

const normalizeTaskStatus = (raw?: string) => String(raw || "").trim().toLowerCase();


const normalizeProgress = (raw: any, fallback?: number): number | undefined => {
  const value = Number(raw);
  if (!Number.isFinite(value)) {
    return typeof fallback === "number" ? fallback : undefined;
  }
  if (value < 0) return 0;
  if (value > 100) return 100;
  return Math.round(value);
};

const resolveTaskIDFromEvent = (event: WsBusEvent): string => {
  const payload = (event?.payload || {}) as Record<string, any>;
  return String(
    payload.taskId || payload.task_id || payload.id || event.taskId || "",
  ).trim();
};

const pendingTaskEventPatches = new Map<string, Partial<BulkTask>>();

const mergeTaskPatch = (current: Partial<BulkTask>, incoming: Partial<BulkTask>): Partial<BulkTask> => {
  return {
    ...current,
    ...incoming,
    progress: normalizeProgress(incoming.progress, normalizeProgress(current.progress)),
    completedAt: incoming.completedAt || current.completedAt,
  };
};

const taskBusBindingState: {
  bound: boolean;
  handler: ((event: WsBusEvent) => void) | null;
} = {
  bound: false,
  handler: null,
};

const createDefaultFilters = (): CustomerListFilters => ({
  keyword: "",
  tier: "",
  source: "",
  region: "",
  riskLevel: "",
  type: undefined,
  tags: [],
  page: 1,
  pageSize: 20,
  sort: "-createdAt",
});

const createDefaultMemberFilters = (): MembershipFilters => ({
  page: 1,
  pageSize: 20,
});

const createDefaultStats = (): MembershipStats => ({
  total: 0,
  active: 0,
  warning: 0,
  downgrade: 0,
  averageGrowthValue: 0,
});

const createDefaultSegments = (): MembershipSegments => ({
  safe: 0,
  warning: 0,
  downgrade: 0,
});

const createReminderState = (): MembershipReminderState => ({
  submitting: false,
  error: null,
  lastTaskId: null,
  channel: "sms",
  templateId: "",
  total: 0,
  success: 0,
});

const createImportState = () => ({
  submitting: false,
  error: null as string | null,
  lastTaskId: null as string | null,
});

const createExportState = () => ({
  submitting: false,
  error: null as string | null,
  lastTaskId: null as string | null,
});

const createMutationState = () => ({
  creating: false,
  updating: false,
  deleting: false,
  error: null as string | null,
  lastAction: null as "create" | "update" | "delete" | null,
  lastCustomerId: null as string | null,
});

const readSavedViews = (): SavedView[] => {
  if (typeof window === "undefined") return [];
  try {
    const raw = window.localStorage.getItem(SAVED_VIEWS_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
};

const persistSavedViews = (views: SavedView[]) => {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(SAVED_VIEWS_KEY, JSON.stringify(views));
  } catch {
    // noop
  }
};

const createViewId = () => {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return `view_${Date.now()}_${Math.round(Math.random() * 1e6)}`;
};

export const useCustomerStore = defineStore("customer.directory", {
  state: () => ({
    list: [] as Customer[],
    meta: {
      total: 0,
      savedViewId: null as string | null,
    },
    loading: false,
    error: "" as string | null,
    filters: createDefaultFilters(),
    savedViews: [] as SavedView[],
    activeViewId: null as string | null,
    selection: [] as string[],
    bulkTasks: [] as BulkTask[],
    lastFetchedAt: "" as string | null,
    membershipInsights: [] as MembershipInsight[],
    membershipSnapshots: [] as MembershipSnapshot[],
    membershipFilters: createDefaultMemberFilters(),
    membershipStats: createDefaultStats(),
    membershipSegments: createDefaultSegments(),
    membershipLoading: false,
    membershipError: "" as string | null,
    membershipLastFetchedAt: "" as string | null,
    reminderState: createReminderState(),
    importState: createImportState(),
    exportState: createExportState(),
    taskStreamError: "" as string | null,
    mutationState: createMutationState(),
    visibleColumns: [
      "name",
      "contact",
      "tier",
      "source",
      "tags",
      "lastOrder",
      "status",
      "owner",
    ] as string[],
  }),
  getters: {
    hasSelection: (state) => state.selection.length > 0,
  },
  actions: {
    setFilters(payload: Partial<CustomerListFilters>) {
      this.filters = { ...this.filters, ...payload };
    },
    resetFilters() {
      this.filters = createDefaultFilters();
    },
    setSavedViews(views: SavedView[]) {
      this.savedViews = views;
    },
    setActiveView(id: string | null) {
      this.activeViewId = id;
    },
    toggleSelection(id: string) {
      if (!id) return;
      if (this.selection.includes(id)) {
        this.selection = this.selection.filter((item) => item !== id);
        return;
      }
      this.selection = [...this.selection, id];
    },
    replaceSelection(ids: string[]) {
      this.selection = Array.from(new Set(ids));
    },
    clearSelection() {
      this.selection = [];
    },
    registerTask(task: BulkTask) {
      const pendingPatch = pendingTaskEventPatches.get(task.taskId);
      const mergedTask = pendingPatch ? ({ ...task, ...pendingPatch } as BulkTask) : task;
      this.bulkTasks = [
        mergedTask,
        ...this.bulkTasks.filter((item) => item.taskId !== task.taskId),
      ];
      if (pendingPatch) {
        pendingTaskEventPatches.delete(task.taskId);
      }
    },
    recordCustomerChange(action: "created" | "updated" | "deleted", customer: Customer | { id: string }) {
      if (!customer?.id) return;
      this.registerTask({
        taskId: `customer-${action}-${customer.id}-${Date.now()}`,
        type: "customerChanged",
        status: "success",
        createdAt: new Date().toISOString(),
        scope: { ids: [customer.id] },
        context: "directory",
        message: action,
      });
    },
    updateTask(taskId: string, patch: Partial<BulkTask>) {
      this.bulkTasks = this.bulkTasks.map((task) =>
        task.taskId === taskId ? { ...task, ...patch } : task
      );
    },
    updateMembershipStats(stats: Partial<MembershipStats>) {
      this.membershipStats = { ...this.membershipStats, ...stats };
    },
    updateMembershipSegments(segments: Partial<MembershipSegments>) {
      this.membershipSegments = { ...this.membershipSegments, ...segments };
    },
    setMembershipFilters(payload: Partial<MembershipFilters>) {
      this.membershipFilters = { ...this.membershipFilters, ...payload };
    },
    resetMembershipFilters() {
      this.membershipFilters = createDefaultMemberFilters();
    },
    clearMembershipError() {
      this.membershipError = null;
    },
    loadSavedViews() {
      this.savedViews = readSavedViews();
    },
    saveCurrentView(name: string) {
      if (!name?.trim()) {
        throw new Error("视图名称不能为空");
      }
      const id = createViewId();
      const view: SavedView = {
        id,
        name: name.trim(),
        filters: { ...this.filters },
        columns: [...this.visibleColumns],
        owner: "local-user",
        shared: false,
        createdAt: new Date().toISOString(),
      };
      this.savedViews = [view, ...this.savedViews].slice(0, 15);
      persistSavedViews(this.savedViews);
      this.activeViewId = id;
    },
    deleteSavedView(id: string) {
      this.savedViews = this.savedViews.filter((view) => view.id !== id);
      persistSavedViews(this.savedViews);
      if (this.activeViewId === id) {
        this.activeViewId = null;
      }
    },
    applySavedView(id: string) {
      const target = this.savedViews.find((view) => view.id === id);
      if (!target) return;
      this.filters = { ...target.filters };
      if (target.columns?.length) {
        this.visibleColumns = [...target.columns];
      }
      this.activeViewId = id;
    },
    setPage(page: number) {
      this.filters = { ...this.filters, page };
    },
    setPageSize(pageSize: number) {
      this.filters = { ...this.filters, pageSize };
    },
    setSort(sort: string) {
      this.filters = { ...this.filters, sort };
    },
    clearError() {
      this.error = null;
    },
    async fetchCustomerById(id: string) {
      if (!id) {
        throw new Error("客户 ID 不能为空");
      }
      const api = useCustomerApi();
      return api.getCustomer(id);
    },
    isFieldMasked(customer: Customer, field: string) {
      return customer?.maskedFields?.includes(field);
    },
    async fetchCustomers(extra?: Partial<CustomerListFilters>) {
      const api = useCustomerApi();
      const metrics = useCustomerMetrics();
      const stopTimer = metrics.startLatencyTimer("customer_list_fetch", {
        source: "directory",
      });
      this.loading = true;
      this.error = null;
      try {
        const response = await api.listCustomers({
          ...this.filters,
          ...extra,
        });
        this.list = response.data || [];
        this.meta.total = response.meta?.total ?? 0;
        this.meta.savedViewId = response.meta?.savedViewId || null;
        this.lastFetchedAt = new Date().toISOString();
      } catch (err: any) {
        this.error =
          err?.data?.message || err?.message || "加载客户列表失败，请稍后重试";
        throw err;
      } finally {
        this.loading = false;
        stopTimer();
      }
    },
    async fetchMemberships(extra?: Partial<MembershipFilters>) {
      const membershipApi = useMembershipInsights();
      this.membershipLoading = true;
      this.membershipError = null;
      try {
        if (extra) {
          this.setMembershipFilters(extra);
        }
        const data = await membershipApi.fetchInsights(this.membershipFilters);
        this.membershipInsights = data.insights;
        this.membershipSnapshots = data.snapshots;
        this.membershipStats = data.stats;
        this.membershipSegments = data.segments;
        this.membershipLastFetchedAt = new Date().toISOString();
      } catch (err: any) {
        this.membershipError =
          err?.data?.message || err?.message || "加载会员数据失败";
        throw err;
      } finally {
        this.membershipLoading = false;
      }
    },
    setMembershipFilterAndFetch(payload: Partial<MembershipFilters>) {
      this.setMembershipFilters(payload);
      return this.fetchMemberships();
    },
    bindTaskStream() {
      const ws = useWsBusClient();
      const handler = taskBusBindingState.handler || ((event: WsBusEvent) => {
        this.applyTaskEvent(event);
      });

      taskBusBindingState.handler = handler;
      TASK_PROGRESS_TOPICS.forEach((topic) => ws.subscribe(topic, handler));
      ws.connect();
      this.taskStreamError = null;
      taskBusBindingState.bound = true;
    },
    applyTaskEvent(event: WsBusEvent) {
      const taskId = resolveTaskIDFromEvent(event);
      if (!taskId) {
        return;
      }
      const existingTask = this.bulkTasks.find((item) => item.taskId === taskId);

      const payload = (event?.payload || {}) as Record<string, any>;
      const status = normalizeTaskStatus(payload.status || payload.state || existingTask?.status || "running");
      const patch: Partial<BulkTask> = {
        status: (status || existingTask?.status || "running") as BulkTask["status"],
        message: String(payload.message || payload.error || existingTask?.message || ""),
        progress: normalizeProgress(payload.progress ?? payload.percent ?? payload.percentage, existingTask?.progress),
      };

      const isTerminalFailed = status === "failed" || status === "error" || status === "cancelled";
      const messageText = String(patch.message || payload.message || "");
      const isImportLike = messageText.includes("导入");
      const isExportLike = messageText.includes("导出") || Boolean(patch.downloadUrl);

      if (isImportLike && this.importState.lastTaskId !== taskId && this.importState.submitting) {
        this.importState.lastTaskId = taskId;
      }
      if (isExportLike && this.exportState.lastTaskId !== taskId && this.exportState.submitting) {
        this.exportState.lastTaskId = taskId;
      }
      if (taskId === this.importState.lastTaskId) {
        if (status === "success") {
          this.importState.error = null;
        } else if (isTerminalFailed) {
          this.importState.error = patch.message || "导入任务失败";
        }
      }
      if (taskId === this.exportState.lastTaskId) {
        if (status === "success") {
          this.exportState.error = null;
        } else if (isTerminalFailed) {
          this.exportState.error = patch.message || "导出任务失败";
        }
      }

      const downloadUrl = String(payload.downloadUrl || payload.download_url || "").trim();
      if (downloadUrl) {
        patch.downloadUrl = downloadUrl;
      }

      if (!patch.message && existingTask?.message) {
        patch.message = existingTask.message;
      }

      if (TERMINAL_TASK_STATUSES.has(status)) {
        patch.completedAt = String(payload.completedAt || payload.completed_at || new Date().toISOString());
        patch.progress = 100;
      }

      if (!existingTask) {
        const buffered = pendingTaskEventPatches.get(taskId) || {};
        const merged = mergeTaskPatch(buffered, patch);
        pendingTaskEventPatches.set(taskId, merged);

        const messageText = String(merged.message || payload.message || "");
        const isExport = messageText.includes("导出") || Boolean(merged.downloadUrl);
        const inferredType = isExport ? "export" : "import";

        this.registerTask({
          taskId,
          type: inferredType as BulkTask["type"],
          status: (merged.status || "running") as BulkTask["status"],
          progress: normalizeProgress(merged.progress, 0),
          message: String(merged.message || ""),
          downloadUrl: String(merged.downloadUrl || "") || undefined,
          createdAt: new Date().toISOString(),
          context: "directory",
          completedAt: merged.completedAt,
        });

        if (!this.importState.lastTaskId && !isExport) {
          this.importState.lastTaskId = taskId;
        }
        if (!this.exportState.lastTaskId && isExport) {
          this.exportState.lastTaskId = taskId;
        }
        return;
      }

      this.updateTask(taskId, patch);
    },
    async submitImportTask(params: {
      file: File;
      context?: "directory" | "members";
      conflictStrategy?: "fail" | "skip";
    }) {
      if (!params.file) {
        throw new Error("请提供导入文件");
      }
      const bulkActions = useCustomerBulkActions();
      this.bindTaskStream();
      this.importState.submitting = true;
      this.importState.error = null;
      try {
        const response = await bulkActions.submitImport({
          file: params.file,
          conflictStrategy: params.conflictStrategy || "fail",
        });
        this.importState.lastTaskId = response.taskId;
        this.registerTask({
          taskId: response.taskId,
          type: "import",
          status: "queued",
          progress: 0,
          createdAt: new Date().toISOString(),
          context: params.context || "directory",
        });
        this.bindTaskStream();
        return response;
      } catch (error: any) {
        this.importState.error =
          error?.data?.message || error?.message || "导入任务创建失败";
        throw error;
      } finally {
        this.importState.submitting = false;
      }
    },
    resetImportState() {
      this.importState = createImportState();
    },
    async fetchTaskStatus(taskId: string) {
      const normalizedTaskID = String(taskId || "").trim();
      if (!normalizedTaskID) {
        throw new Error("taskId is required");
      }
      const { client } = useApiClient();
      const resp = await client<{ success?: boolean; data?: any }>(`/admin/jobs/${normalizedTaskID}`, {
        method: "GET",
      });
      const data = (resp && typeof resp === "object" && "data" in (resp as Record<string, any>))
        ? (resp as Record<string, any>).data
        : resp;
      if (!data || typeof data !== "object") {
        return null;
      }

      const status = normalizeTaskStatus((data as Record<string, any>).status || "running");
      const patch: Partial<BulkTask> = {
        status: (status || "running") as BulkTask["status"],
        message: String((data as Record<string, any>).message || ""),
        progress: normalizeProgress(
          (data as Record<string, any>).progress ??
            (data as Record<string, any>).metadata?.progress,
          undefined,
        ),
        downloadUrl: String(
          (data as Record<string, any>).downloadUrl ||
            (data as Record<string, any>).download_url ||
            "",
        ) || undefined,
      };

      if (TERMINAL_TASK_STATUSES.has(status)) {
        patch.completedAt = String(
          (data as Record<string, any>).completedAt ||
            (data as Record<string, any>).completed_at ||
            new Date().toISOString(),
        );
      }

      const existing = this.bulkTasks.find((item) => item.taskId === normalizedTaskID);
      if (existing) {
        this.updateTask(normalizedTaskID, patch);
      } else {
        this.registerTask({
          taskId: normalizedTaskID,
          type: String((data as Record<string, any>).type || "").includes("export") ? "export" : "import",
          status: (patch.status || "running") as BulkTask["status"],
          progress: normalizeProgress(patch.progress, 0),
          message: patch.message,
          downloadUrl: patch.downloadUrl,
          createdAt: String((data as Record<string, any>).createdAt || new Date().toISOString()),
          completedAt: patch.completedAt,
          context: "directory",
        });
      }

      return this.bulkTasks.find((item) => item.taskId === normalizedTaskID) || null;
    },
    async submitExportTask(params: {
      fields: string[];
      filters?: CustomerListFilters;
      context?: "directory" | "members";
    }) {
      const bulkActions = useCustomerBulkActions();
      this.bindTaskStream();
      this.exportState.submitting = true;
      this.exportState.error = null;
      try {
        const response = await bulkActions.submitExport({
          filters: params.filters,
          fields: params.fields,
        });
        this.exportState.lastTaskId = response.taskId;
        this.registerTask({
          taskId: response.taskId,
          type: "export",
          status: "queued",
          progress: 0,
          createdAt: new Date().toISOString(),
          scope: { filters: params.filters },
          context: params.context || "directory",
        });
        this.bindTaskStream();
        return response;
      } catch (error: any) {
        this.exportState.error =
          error?.data?.message || error?.message || "导出任务创建失败";
        throw error;
      } finally {
        this.exportState.submitting = false;
      }
    },
    resetExportState() {
      this.exportState = createExportState();
    },
    clearReminderState() {
      this.reminderState = {
        ...createReminderState(),
        channel: this.reminderState.channel,
      };
    },
    async triggerMembershipReminder(payload: BulkReminderPayload) {
      if (!payload.ids?.length) {
        throw new Error("请至少选择一个目标客户");
      }
      const bulkActions = useCustomerBulkActions();
      const metrics = useCustomerMetrics();
      this.reminderState.submitting = true;
      this.reminderState.error = null;
      this.reminderState.channel = payload.channel;
      this.reminderState.templateId = payload.templateId;
      this.reminderState.total = payload.ids.length;
      try {
        const response = await bulkActions.submitReminder({
          ids: payload.ids,
          channel: payload.channel,
          templateId: payload.templateId,
          metadata: payload.metadata,
        });
        this.reminderState.lastTaskId = response.taskId;
        this.reminderState.success = payload.ids.length;
        this.registerTask({
          taskId: response.taskId,
          status: "queued",
          type: "reminder",
          createdAt: new Date().toISOString(),
          scope: { ids: [...payload.ids] },
          context: "members",
        });
        this.bindTaskStream();
        metrics.recordReminderResult({
          channel: payload.channel,
          total: payload.ids.length,
          success: payload.ids.length,
        });
        return response;
      } catch (error: any) {
        this.reminderState.error =
          error?.data?.message || error?.message || "批量提醒失败";
        this.reminderState.success = 0;
        metrics.recordReminderResult({
          channel: payload.channel,
          total: payload.ids.length,
          success: 0,
        });
        throw error;
      } finally {
        this.reminderState.submitting = false;
      }
    },
    resetMutationState() {
      this.mutationState = createMutationState();
    },
    async createCustomer(payload: CustomerCreatePayload) {
      if (!payload?.name) {
        throw new Error("缺少客户必填字段");
      }
      const api = useCustomerApi();
      this.mutationState.creating = true;
      this.mutationState.error = null;
      try {
        const customer = await api.createCustomer(payload);
        this.mutationState.lastAction = "create";
        this.mutationState.lastCustomerId = customer.id;
        if (this.filters.page !== 1) {
          this.setPage(1);
        }
        await this.fetchCustomers();
        if (this.membershipSnapshots.length) {
          await this.fetchMemberships();
        }
        this.recordCustomerChange("created", customer);
        return customer;
      } catch (error: any) {
        this.mutationState.error =
          error?.data?.message || error?.message || "创建客户失败";
        throw error;
      } finally {
        this.mutationState.creating = false;
      }
    },
    async updateCustomer(id: string, payload: CustomerUpdatePayload) {
      if (!id) {
        throw new Error("customerId is required");
      }
      const api = useCustomerApi();
      this.mutationState.updating = true;
      this.mutationState.error = null;
      try {
        const customer = await api.updateCustomer(id, payload);
        this.mutationState.lastAction = "update";
        this.mutationState.lastCustomerId = customer.id;
        await this.fetchCustomers();
        if (this.membershipSnapshots.length) {
          await this.fetchMemberships();
        }
        this.recordCustomerChange("updated", customer);
        return customer;
      } catch (error: any) {
        this.mutationState.error =
          error?.data?.message || error?.message || "更新客户失败";
        throw error;
      } finally {
        this.mutationState.updating = false;
      }
    },
    async deleteCustomer(params: { id: string; reason: string; customer?: Customer }) {
      if (!params?.id) {
        throw new Error("customerId is required");
      }
      const api = useCustomerApi();
      this.mutationState.deleting = true;
      this.mutationState.error = null;
      try {
        await api.deleteCustomer(params.id, { reason: params.reason });
        this.mutationState.lastAction = "delete";
        this.mutationState.lastCustomerId = params.id;
        this.selection = this.selection.filter((item) => item !== params.id);
        await this.fetchCustomers();
        if (this.membershipSnapshots.length) {
          await this.fetchMemberships();
        }
        this.recordCustomerChange("deleted", params.customer ?? { id: params.id });
      } catch (error: any) {
        this.mutationState.error =
          error?.data?.message || error?.message || "删除客户失败";
        throw error;
      } finally {
        this.mutationState.deleting = false;
      }
    },
  },
});
