import { defineStore } from "pinia";
import type {
  BulkTask,
  Customer,
  CustomerListFilters,
  MembershipFilters,
  MembershipSnapshot,
  MembershipStats,
  SavedView,
} from "~/types/customer";

const createDefaultFilters = (): CustomerListFilters => ({
  keyword: "",
  tier: "",
  source: "",
  region: "",
  riskLevel: "",
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
    activeViewId: "" as string | null,
    selection: [] as string[],
    bulkTasks: [] as BulkTask[],
    lastFetchedAt: "" as string | null,
    membershipSnapshots: [] as MembershipSnapshot[],
    membershipFilters: createDefaultMemberFilters(),
    membershipStats: createDefaultStats(),
    membershipLoading: false,
    membershipError: "" as string | null,
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
  },
});
