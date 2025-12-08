import { defineStore } from "pinia";
import { useCustomerService } from "~/composables/api/services/customerService";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import { useMembershipInsights } from "~/composables/useMembershipInsights";
import { useCustomerBulkActions } from "~/composables/useCustomerBulkActions";
import type {
  AuditContext,
  BulkReminderPayload,
  BulkTask,
  Customer,
  CustomerListFilters,
  MembershipFilters,
  MembershipInsight,
  MembershipReminderState,
  MembershipSnapshot,
  MembershipSegments,
  MembershipStats,
  SavedView,
} from "~/types/customer";

const SAVED_VIEWS_KEY = "px_customer_saved_views";

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
    taskPolling: {} as Record<string, boolean>,
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
      this.bulkTasks = [
        task,
        ...this.bulkTasks.filter((item) => item.taskId !== task.taskId),
      ];
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
    isFieldMasked(customer: Customer, field: string) {
      return customer?.maskedFields?.includes(field);
    },
    async fetchCustomers(extra?: Partial<CustomerListFilters>) {
      const service = useCustomerService();
      const metrics = useCustomerMetrics();
      const stopTimer = metrics.startLatencyTimer("customer_list_fetch", {
        source: "directory",
      });
      this.loading = true;
      this.error = null;
      try {
        const response = await service.listCustomers({
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
    async pollTaskStatus(taskId: string) {
      if (!taskId || this.taskPolling[taskId]) return;
      const bulkActions = useCustomerBulkActions();
      this.taskPolling[taskId] = true;
      try {
        const status = await bulkActions.pollJobUntilFinished(taskId, {
          intervalMs: 5_000,
        });
        this.updateTask(taskId, {
          status: status.status,
          message: status.message,
          completedAt: status.completedAt || new Date().toISOString(),
          downloadUrl: status.downloadUrl,
        });
      } catch (error: any) {
        this.updateTask(taskId, {
          status: "failed",
          message: error?.message || "任务执行失败",
        });
      } finally {
        delete this.taskPolling[taskId];
      }
    },
    async submitImportTask(params: {
      file: File;
      context?: "directory" | "members";
      audit?: AuditContext;
    }) {
      if (!params.file) {
        throw new Error("请提供导入文件");
      }
      const bulkActions = useCustomerBulkActions();
      this.importState.submitting = true;
      this.importState.error = null;
      try {
        const response = await bulkActions.submitImport({
          file: params.file,
          audit:
            params.audit ||
            (params.context === "members"
              ? {
                  action: "customer.membership.import",
                  resource: "customers:members",
                }
              : undefined),
        });
        this.importState.lastTaskId = response.taskId;
        this.registerTask({
          taskId: response.taskId,
          type: "import",
          status: "queued",
          createdAt: new Date().toISOString(),
          context: params.context || "directory",
        });
        this.pollTaskStatus(response.taskId);
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
    async submitExportTask(params: {
      fields: string[];
      filters?: CustomerListFilters;
      context?: "directory" | "members";
      audit?: AuditContext;
    }) {
      const bulkActions = useCustomerBulkActions();
      this.exportState.submitting = true;
      this.exportState.error = null;
      try {
        const response = await bulkActions.submitExport({
          filters: params.filters,
          fields: params.fields,
          audit:
            params.audit ||
            (params.context === "members"
              ? {
                  action: "customer.membership.export",
                  resource: "customers:members",
                }
              : undefined),
        });
        this.exportState.lastTaskId = response.taskId;
        this.registerTask({
          taskId: response.taskId,
          type: "export",
          status: "queued",
          createdAt: new Date().toISOString(),
          scope: { filters: params.filters },
          context: params.context || "directory",
        });
        this.pollTaskStatus(response.taskId);
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
        this.pollTaskStatus(response.taskId);
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
  },
});
