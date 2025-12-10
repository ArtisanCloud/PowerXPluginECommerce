import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useCustomerStore } from "../../app/stores/customer";

const listCustomersMock = vi.fn();
const createCustomerMock = vi.fn();
const updateCustomerMock = vi.fn();
const deleteCustomerMock = vi.fn();
const bulkActionMocks = {
  submitReminder: vi.fn(),
  submitImport: vi.fn(),
  submitExport: vi.fn(),
  pollJobUntilFinished: vi.fn(),
};

vi.mock("../../app/composables/api/services/customerService", () => ({
  useCustomerService: () => ({
    listCustomers: listCustomersMock,
    createCustomer: createCustomerMock,
    updateCustomer: updateCustomerMock,
    deleteCustomer: deleteCustomerMock,
  }),
}));

vi.mock("../../app/composables/useCustomerMetrics", () => ({
  useCustomerMetrics: () => ({
    startLatencyTimer: () => () => undefined,
    recordEvent: () => undefined,
    recordReminderResult: () => undefined,
  }),
}));

vi.mock("../../app/composables/useCustomerBulkActions", () => ({
  useCustomerBulkActions: () => bulkActionMocks,
}));

describe("customer CRUD store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    listCustomersMock.mockResolvedValue({ data: [], meta: { total: 0 } });
    createCustomerMock.mockResolvedValue({ id: "cust-1", name: "Demo" });
    updateCustomerMock.mockResolvedValue({ id: "cust-1", name: "Updated" });
    deleteCustomerMock.mockResolvedValue({ success: true });
  });

  it("creates customer and records task", async () => {
    const store = useCustomerStore();
    await store.createCustomer({
      name: "Demo",
      type: "individual",
      phone: "+8613800000000",
      membershipTier: "gold",
      source: "website",
      tags: [],
    });
    expect(createCustomerMock).toHaveBeenCalled();
    expect(store.mutationState.lastAction).toBe("create");
    expect(store.bulkTasks[0]?.type).toBe("customerChanged");
  });

  it("updates customer metadata", async () => {
    const store = useCustomerStore();
    await store.updateCustomer("cust-1", { name: "New" });
    expect(updateCustomerMock).toHaveBeenCalledWith("cust-1", { name: "New" });
    expect(store.mutationState.lastAction).toBe("update");
  });

  it("deletes customer with reason", async () => {
    const store = useCustomerStore();
    store.replaceSelection(["cust-1"]);
    await store.deleteCustomer({ id: "cust-1", reason: "cleanup" });
    expect(deleteCustomerMock).toHaveBeenCalled();
    expect(store.mutationState.lastAction).toBe("delete");
    expect(store.selection).toHaveLength(0);
  });
});
