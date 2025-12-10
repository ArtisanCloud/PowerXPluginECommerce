import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useCustomerStore } from "../../app/stores/customer";

const submitImportMock = vi.fn();
const submitExportMock = vi.fn();
const pollJobMock = vi.fn();

vi.mock("../../app/composables/api/services/customerService", () => ({
  useCustomerService: () => ({
    listCustomers: vi.fn(),
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

vi.mock("../../app/composables/useMembershipInsights", () => ({
  useMembershipInsights: () => ({
    fetchInsights: vi.fn(),
  }),
}));

vi.mock("../../app/composables/useCustomerBulkActions", () => ({
  useCustomerBulkActions: () => ({
    submitImport: submitImportMock,
    submitExport: submitExportMock,
    pollJobUntilFinished: pollJobMock,
    submitReminder: vi.fn(),
  }),
}));

describe("customer store import/export", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    submitImportMock.mockReset();
    submitExportMock.mockReset();
    pollJobMock.mockReset();
    (globalThis as any).window = {
      localStorage: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
      },
    };
  });

  it("submits import task and registers bulk task", async () => {
    submitImportMock.mockResolvedValueOnce({ taskId: "import-001" });
    pollJobMock.mockResolvedValueOnce({
      taskId: "import-001",
      status: "success",
    });
    const store = useCustomerStore();

    await store.submitImportTask({
      file: { name: "demo.csv" } as File,
      context: "directory",
    });

    expect(store.importState.lastTaskId).toBe("import-001");
    expect(store.bulkTasks[0]?.type).toBe("import");
    expect(pollJobMock).toHaveBeenCalledWith("import-001", expect.any(Object));
  });

  it("captures import errors", async () => {
    submitImportMock.mockRejectedValueOnce(new Error("upload fail"));
    const store = useCustomerStore();
    await expect(
      store.submitImportTask({
        file: { name: "bad.csv" } as File,
      })
    ).rejects.toThrow();
    expect(store.importState.error).toBeTruthy();
  });

  it("submits export task with context", async () => {
    submitExportMock.mockResolvedValueOnce({ taskId: "export-123" });
    pollJobMock.mockResolvedValueOnce({
      taskId: "export-123",
      status: "success",
    });
    const store = useCustomerStore();
    await store.submitExportTask({
      fields: ["id", "name"],
      filters: { tier: "gold" },
      context: "members",
    });
    expect(submitExportMock).toHaveBeenCalledWith(
      expect.objectContaining({
        filters: { tier: "gold" },
      })
    );
    expect(store.exportState.lastTaskId).toBe("export-123");
    expect(store.bulkTasks[0]?.type).toBe("export");
  });
});
