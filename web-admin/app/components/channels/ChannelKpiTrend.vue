<template>
  <UCard>
    <template #header>
      <div class="flex items-center justify-between">
        <div>
          <h3 class="text-lg font-semibold">KPI 指标</h3>
          <p class="text-sm text-gray-500">展示 d1/d7/d30 指标快照</p>
        </div>
        <UBadge variant="subtle" color="neutral">{{ metrics?.length ?? 0 }} 条</UBadge>
      </div>
    </template>
    <div v-if="metrics?.length">
      <UTable :columns="columns" :data="tableData" />
    </div>
    <div v-else class="py-6 text-center text-sm text-gray-500">
      暂无 KPI 数据
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { TableColumn } from '@nuxt/ui'
import type { ChannelMetric } from '~/types/channels'

const props = defineProps<{
  metrics?: ChannelMetric[]
}>()

const columns: TableColumn[] = [
  { accessorKey: 'window', header: '窗口' },
  { accessorKey: 'gmv', header: 'GMV' },
  { accessorKey: 'orders', header: '订单量' },
  { accessorKey: 'gmvGrowthRate', header: 'GMV 环比' },
  { accessorKey: 'inventoryCoverage', header: '库存覆盖' },
  { accessorKey: 'errorRate', header: '错误率' },
  { accessorKey: 'syncSuccessRate', header: '同步成功率' },
]

const tableData = computed(() =>
  (props.metrics ?? []).map((metric) => ({
    window: metric.window?.toUpperCase(),
    gmv: formatCurrency(metric.gmv),
    orders: metric.orders.toLocaleString(),
    gmvGrowthRate: formatPercent(metric.gmvGrowthRate),
    inventoryCoverage: formatPercent(metric.inventoryCoverage),
    errorRate: formatPercent(metric.errorRate),
    syncSuccessRate: formatPercent(metric.syncSuccessRate),
  })),
)

const formatCurrency = (value?: number) => {
  if (!value) return '¥0'
  return `¥${value.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`
}

const formatPercent = (value?: number) => {
  if (value === undefined || Number.isNaN(value)) return '0%'
  return `${(value * 100).toFixed(1)}%`
}
</script>
