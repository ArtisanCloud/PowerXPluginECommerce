import { defineStore } from "pinia";
import { useCustomerService } from "~/composables/api/services/customerService";
import { useCustomerMetrics } from "~/composables/useCustomerMetrics";
import type {
  BulkTask,
  Customer,
  CustomerListFilters,
  MembershipFilters,
  MembershipSnapshot,
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
    membershipSnapshots: [] as MembershipSnapshot[],
    membershipFilters: createDefaultMemberFilters(),
    membershipStats: createDefaultStats(),
    membershipLoading: false,
    membershipError: "" as string | null,
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
    updateMembershipStats(stats: Partial<MembershipStats>) {
      this.membershipStats = { ...this.membershipStats, ...stats };
    },
    setMembershipFilters(payload: Partial<MembershipFilters>) {
      this.membershipFilters = { ...this.membershipFilters, ...payload };
    },
    resetMembershipFilters() {
      this.membershipFilters = createDefaultMemberFilters();
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
    async fetchMemberships() {
      const service = useCustomerService();
      this.membershipLoading = true;
      this.membershipError = null;
      try {
        const response = await service.listMembers(this.membershipFilters);
        this.membershipSnapshots =
          response.data?.map((entry) => entry.snapshot) || [];
        if (response.stats) {
          this.membershipStats = response.stats;
        } else {
          this.membershipStats.total = response.meta?.total ?? 0;
        }
      } catch (err: any) {
        this.membershipError =
          err?.data?.message || err?.message || "加载会员数据失败";
        throw err;
      } finally {
        this.membershipLoading = false;
      }
    },
  },
});
