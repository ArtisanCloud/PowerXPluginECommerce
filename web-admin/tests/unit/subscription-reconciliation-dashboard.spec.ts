import { describe, it, beforeEach, expect, vi } from "vitest"
import { mount } from "@vue/test-utils"
import { reactive, ref } from "vue"
import SubscriptionReconciliationDashboardPage from "../../app/pages/finance/subscription-reconciliation.vue"

const fetchDashboardMock = vi.fn()
const exportDashboardMock = vi.fn()

const storeState = {
  dashboard: ref({
    deltaRate: 0.12,
    recoveryRate: 0.35,
    avgHandleHours: 10.5,
    pendingTasks: 6,
    byDeltaType: [{ type: "missing_payment", count: 4 }],
  }),
  filters: reactive({
    from: "",
    to: "",
    channel: "",
    plan: "",
    region: "",
    failureReason: "",
  }),
  loading: ref(false),
  exporting: ref(false),
  error: ref<string | null>(null),
}

vi.mock("../../app/stores/subscription-reconciliation", () => ({
  useSubscriptionReconciliationStore: () => ({
    ...storeState,
    fetchDashboard: fetchDashboardMock,
    exportDashboard: exportDashboardMock,
  }),
}))

vi.mock("pinia", async () => {
  const actual = await vi.importActual<typeof import("pinia")>("pinia")
  return {
    ...actual,
    storeToRefs: (store: any) => store,
  }
})

describe("subscription reconciliation dashboard page", () => {
  beforeEach(() => {
    fetchDashboardMock.mockReset()
    exportDashboardMock.mockReset()
  })

  it("renders filters and triggers query/export actions", async () => {
    fetchDashboardMock.mockResolvedValueOnce(storeState.dashboard.value)
    exportDashboardMock.mockResolvedValueOnce("/export.csv")

    const wrapper = mount(SubscriptionReconciliationDashboardPage)
    await Promise.resolve()

    expect(fetchDashboardMock).toHaveBeenCalledTimes(1)
    expect(wrapper.get("[data-testid='metric-delta-rate']").text()).toContain("0.12")
    expect(wrapper.findAll("[data-testid='delta-type-row']")).toHaveLength(1)

    await wrapper.get("[data-testid='btn-query']").trigger("click")
    expect(fetchDashboardMock).toHaveBeenCalledTimes(2)

    await wrapper.get("[data-testid='btn-export']").trigger("click")
    expect(exportDashboardMock).toHaveBeenCalledWith("csv")
  })
})

