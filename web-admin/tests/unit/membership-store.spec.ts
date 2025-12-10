import { describe, it, beforeEach, expect, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useCustomerStore } from "../../app/stores/customer";

const fetchInsightsMock = vi.fn();
const submitReminderMock = vi.fn();
const recordReminderResultMock = vi.fn();

vi.mock("../../app/composables/useMembershipInsights", () => ({
  useMembershipInsights: () => ({
    fetchInsights: fetchInsightsMock,
  }),
}));

vi.mock("../../app/composables/useCustomerBulkActions", () => ({
  useCustomerBulkActions: () => ({
    submitReminder: submitReminderMock,
  }),
}));

vi.mock("../../app/composables/api/services/customerService", () => ({
  useCustomerService: () => ({
    listCustomers: vi.fn(),
    getCustomer: vi.fn(),
  }),
}));

vi.mock("../../app/composables/useCustomerMetrics", () => ({
  useCustomerMetrics: () => ({
    startLatencyTimer: () => () => undefined,
    trackJob: () => () => undefined,
    recordEvent: () => undefined,
    recordReminderResult: recordReminderResultMock,
  }),
}));

describe("useCustomerStore - membership", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    fetchInsightsMock.mockReset();
    submitReminderMock.mockReset();
    recordReminderResultMock.mockReset();
    (globalThis as any).window = {
      localStorage: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
      },
    };
  });

  it("loads membership insights and stats", async () => {
    fetchInsightsMock.mockResolvedValueOnce({
      insights: [
        {
          customer: { id: "cus-1", name: "张三" },
          snapshot: {
            customerId: "cus-1",
            tier: "gold",
            growthValue: 80,
            points: 120,
            retentionStatus: "safe",
          },
        },
      ],
      snapshots: [],
      stats: {
        total: 1,
        active: 1,
        warning: 0,
        downgrade: 0,
        averageGrowthValue: 80,
      },
      segments: { safe: 1, warning: 0, downgrade: 0 },
      total: 1,
    });

    const store = useCustomerStore();
    await store.fetchMemberships();

    expect(store.membershipInsights).toHaveLength(1);
    expect(store.membershipStats.total).toBe(1);
    expect(store.membershipSegments.safe).toBe(1);
  });

  it("submits bulk reminders and updates reminder state", async () => {
    submitReminderMock.mockResolvedValueOnce({ taskId: "job-001" });
    const store = useCustomerStore();

    await store.triggerMembershipReminder({
      ids: ["cus-1", "cus-2"],
      channel: "sms",
      templateId: "retain-tier-1",
    });

    expect(submitReminderMock).toHaveBeenCalledWith({
      ids: ["cus-1", "cus-2"],
      channel: "sms",
      templateId: "retain-tier-1",
      metadata: undefined,
    });
    expect(store.reminderState.lastTaskId).toBe("job-001");
    expect(store.reminderState.success).toBe(2);
    expect(recordReminderResultMock).toHaveBeenCalledWith({
      channel: "sms",
      total: 2,
      success: 2,
    });
  });
});
