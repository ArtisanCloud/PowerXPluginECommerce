import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useCustomerStore } from "../../app/stores/customer";

const listCustomersMock = vi.fn();
const listMembersMock = vi.fn();

vi.mock("../../app/composables/api/services/customerService", () => ({
  useCustomerService: () => ({
    listCustomers: listCustomersMock,
    listMembers: listMembersMock,
  }),
}));

vi.mock("../../app/composables/useCustomerMetrics", () => ({
  useCustomerMetrics: () => ({
    startLatencyTimer: () => () => undefined,
    trackJob: () => () => undefined,
    recordEvent: () => undefined,
    recordReminderResult: () => undefined,
  }),
}));

describe("useCustomerStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    listCustomersMock.mockReset();
    listMembersMock.mockReset();
    (globalThis as any).window = {
      localStorage: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
      },
    };
  });

  it("fetches customers and updates meta", async () => {
    listCustomersMock.mockResolvedValueOnce({
      data: [
        {
          id: "cus-1",
          name: "张三",
          status: "active",
        },
      ],
      meta: {
        total: 1,
        savedViewId: null,
      },
    });

    const store = useCustomerStore();
    await store.fetchCustomers();

    expect(store.list).toHaveLength(1);
    expect(store.meta.total).toBe(1);
    expect(store.error).toBeNull();
  });

  it("saves and applies views", async () => {
    const store = useCustomerStore();
    store.loadSavedViews();
    expect(store.savedViews).toEqual([]);

    store.saveCurrentView("全部客户");
    expect(store.savedViews).toHaveLength(1);
    expect(store.activeViewId).toBeTruthy();

    store.setFilters({ keyword: "abc" });
    const savedId = store.savedViews[0].id;
    store.applySavedView(savedId);
    expect(store.filters.keyword).toBe("");
  });

  it("handles fetch errors", async () => {
    listCustomersMock.mockRejectedValueOnce(new Error("network error"));
    const store = useCustomerStore();
    await expect(store.fetchCustomers()).rejects.toThrowError();
    expect(store.error).toBeTruthy();
  });
});
