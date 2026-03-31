import { computed, reactive, ref } from "vue"
import { defineStore } from "pinia"

export interface SubscriptionReconciliationFilter {
  from: string
  to: string
  channel: string
  plan: string
  region: string
  failureReason: string
}

export interface DashboardDeltaType {
  type: string
  count: number
}

export interface SubscriptionReconciliationDashboard {
  deltaRate: number
  recoveryRate: number
  avgHandleHours: number
  pendingTasks: number
  byDeltaType: DashboardDeltaType[]
}

const defaultDashboard = (): SubscriptionReconciliationDashboard => ({
  deltaRate: 0,
  recoveryRate: 0,
  avgHandleHours: 0,
  pendingTasks: 0,
  byDeltaType: [],
})

export const useSubscriptionReconciliationStore = defineStore("subscriptionReconciliation", () => {
  const config = useRuntimeConfig()
  const apiBase = computed(() => String(config.public.apiBaseUrl || ""))

  const loading = ref(false)
  const exporting = ref(false)
  const error = ref<string | null>(null)
  const dashboard = ref<SubscriptionReconciliationDashboard>(defaultDashboard())

  const filters = reactive<SubscriptionReconciliationFilter>({
    from: "",
    to: "",
    channel: "",
    plan: "",
    region: "",
    failureReason: "",
  })

  async function fetchDashboard(partial: Partial<SubscriptionReconciliationFilter> = {}) {
    Object.assign(filters, partial)
    loading.value = true
    error.value = null
    try {
      const response = await $fetch<{ data?: SubscriptionReconciliationDashboard }>(
        `${apiBase.value}/admin/subscription-reconciliation/dashboard`,
        { params: { ...filters } }
      )
      dashboard.value = response?.data ?? defaultDashboard()
      return dashboard.value
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      dashboard.value = defaultDashboard()
      throw e
    } finally {
      loading.value = false
    }
  }

  async function exportDashboard(format: "csv" | "json" = "csv") {
    exporting.value = true
    error.value = null
    try {
      if (format === "json") {
        return fetchDashboard()
      }
      const query = new URLSearchParams({ ...filters }).toString()
      const url = `${apiBase.value}/admin/subscription-reconciliation/dashboard/export?${query}`
      if (typeof window !== "undefined") {
        window.open(url, "_blank", "noopener,noreferrer")
      }
      return url
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e)
      throw e
    } finally {
      exporting.value = false
    }
  }

  return {
    loading,
    exporting,
    error,
    filters,
    dashboard,
    fetchDashboard,
    exportDashboard,
  }
})

