<script setup lang="ts">
import { onMounted } from "vue"
import { storeToRefs } from "pinia"
import { useSubscriptionReconciliationStore } from "~/stores/subscription-reconciliation"

const store = useSubscriptionReconciliationStore()
const { dashboard, filters, loading, exporting, error } = storeToRefs(store)

const runQuery = async () => {
  await store.fetchDashboard()
}

const exportCSV = async () => {
  await store.exportDashboard("csv")
}

onMounted(async () => {
  await runQuery()
})
</script>

<template>
  <div class="space-y-4 p-4" data-testid="reconciliation-dashboard-page">
    <h1 class="text-xl font-semibold">订阅对账与续费治理看板</h1>

    <div class="grid grid-cols-2 gap-3 md:grid-cols-3">
      <label class="space-y-1">
        <span class="text-xs">开始日期</span>
        <input v-model="filters.from" data-testid="filter-from" type="date" class="w-full rounded border px-2 py-1" />
      </label>
      <label class="space-y-1">
        <span class="text-xs">结束日期</span>
        <input v-model="filters.to" data-testid="filter-to" type="date" class="w-full rounded border px-2 py-1" />
      </label>
      <label class="space-y-1">
        <span class="text-xs">渠道</span>
        <input v-model="filters.channel" data-testid="filter-channel" type="text" class="w-full rounded border px-2 py-1" placeholder="app/web" />
      </label>
      <label class="space-y-1">
        <span class="text-xs">套餐</span>
        <input v-model="filters.plan" data-testid="filter-plan" type="text" class="w-full rounded border px-2 py-1" placeholder="pro/basic" />
      </label>
      <label class="space-y-1">
        <span class="text-xs">地区</span>
        <input v-model="filters.region" data-testid="filter-region" type="text" class="w-full rounded border px-2 py-1" placeholder="CN/US" />
      </label>
      <label class="space-y-1">
        <span class="text-xs">失败原因</span>
        <input
          v-model="filters.failureReason"
          data-testid="filter-failure-reason"
          type="text"
          class="w-full rounded border px-2 py-1"
          placeholder="always_fail"
        />
      </label>
    </div>

    <div class="flex gap-2">
      <button data-testid="btn-query" class="rounded bg-black px-3 py-1 text-white disabled:opacity-60" :disabled="loading" @click="runQuery">
        {{ loading ? "查询中..." : "查询看板" }}
      </button>
      <button data-testid="btn-export" class="rounded border px-3 py-1 disabled:opacity-60" :disabled="exporting" @click="exportCSV">
        {{ exporting ? "导出中..." : "导出 CSV" }}
      </button>
    </div>

    <p v-if="error" class="text-sm text-red-600" data-testid="dashboard-error">{{ error }}</p>

    <div class="grid grid-cols-2 gap-3 md:grid-cols-4">
      <div class="rounded border p-3">
        <p class="text-xs text-gray-500">差异率</p>
        <p data-testid="metric-delta-rate" class="text-lg font-medium">{{ dashboard.deltaRate }}</p>
      </div>
      <div class="rounded border p-3">
        <p class="text-xs text-gray-500">恢复率</p>
        <p data-testid="metric-recovery-rate" class="text-lg font-medium">{{ dashboard.recoveryRate }}</p>
      </div>
      <div class="rounded border p-3">
        <p class="text-xs text-gray-500">平均处理时长（小时）</p>
        <p data-testid="metric-avg-handle-hours" class="text-lg font-medium">{{ dashboard.avgHandleHours }}</p>
      </div>
      <div class="rounded border p-3">
        <p class="text-xs text-gray-500">待处理积压</p>
        <p data-testid="metric-pending-tasks" class="text-lg font-medium">{{ dashboard.pendingTasks }}</p>
      </div>
    </div>

    <div class="rounded border p-3">
      <h2 class="mb-2 text-sm font-semibold">差异责任归因</h2>
      <table class="w-full text-sm">
        <thead>
          <tr class="text-left text-gray-500">
            <th class="pb-1">差异类型</th>
            <th class="pb-1">数量</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in dashboard.byDeltaType" :key="item.type" data-testid="delta-type-row">
            <td class="py-1">{{ item.type }}</td>
            <td class="py-1">{{ item.count }}</td>
          </tr>
          <tr v-if="dashboard.byDeltaType.length === 0">
            <td class="py-2 text-gray-400" colspan="2">暂无数据</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
