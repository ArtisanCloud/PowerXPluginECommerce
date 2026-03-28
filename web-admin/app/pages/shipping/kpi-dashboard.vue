<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">物流 KPI 大屏</h1>
        <p class="text-gray-500 dark:text-gray-400">按租户/承运商/仓库/区域查看履约指标，支持趋势与异常钻取。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <USelect v-model="filters.dimension" class="w-36" :options="dimensionOptions" />
        <UInput v-model.number="filters.windowHours" class="w-32" type="number" placeholder="窗口(h)" />
        <UInput v-model="filters.carrierId" class="w-40" placeholder="承运商 ID" />
        <UInput v-model="filters.warehouseId" class="w-40" placeholder="仓库 ID" />
        <UInput v-model="filters.destinationZone" class="w-40" placeholder="目的区域" />
        <UButton color="primary" :loading="loading" @click="loadAll">刷新</UButton>
        <UButton color="neutral" variant="soft" :loading="exporting" @click="exportCSV">导出</UButton>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-5">
      <UCard>
        <p class="text-xs text-gray-500">总运单</p>
        <p class="mt-1 text-xl font-semibold">{{ overview.totalWaybills }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">妥投数</p>
        <p class="mt-1 text-xl font-semibold">{{ overview.deliveredCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">异常数</p>
        <p class="mt-1 text-xl font-semibold text-warning-600">{{ overview.exceptionCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">准时率</p>
        <p class="mt-1 text-xl font-semibold">{{ overview.onTimeRate.toFixed(2) }}%</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">单均成本</p>
        <p class="mt-1 text-xl font-semibold">¥{{ overview.avgCost.toFixed(2) }}</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">维度趋势</h3>
      </template>
      <UTable :columns="trendColumns" :data="trends" />
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">异常钻取</h3>
      </template>
      <UTable :columns="drilldownColumns" :data="drilldowns" />
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import {
  type LogisticsKPIDashboardDrilldown,
  type LogisticsKPIDashboardOverview,
  type LogisticsKPIDashboardTrend,
  useLogisticsApi,
} from "~/composables/api";

definePageMeta({
  name: "shipping-kpi-dashboard",
});

const logisticsApi = useLogisticsApi();
const loading = ref(false);
const exporting = ref(false);

const dimensionOptions = [
  { label: "承运商", value: "carrier" },
  { label: "仓库", value: "warehouse" },
  { label: "区域", value: "destination_zone" },
  { label: "租户", value: "tenant" },
];

const filters = reactive({
  dimension: "carrier" as "tenant" | "carrier" | "warehouse" | "destination_zone",
  windowHours: 24,
  carrierId: "",
  warehouseId: "",
  destinationZone: "",
});

const overview = ref<LogisticsKPIDashboardOverview>({
  windowHours: 24,
  dimension: "carrier",
  totalWaybills: 0,
  deliveredCount: 0,
  exceptionCount: 0,
  timeoutCount: 0,
  onTimeRate: 0,
  deliverySuccessRate: 0,
  avgTransitHours: 0,
  totalCost: 0,
  avgCost: 0,
});
const trends = ref<LogisticsKPIDashboardTrend[]>([]);
const drilldowns = ref<LogisticsKPIDashboardDrilldown[]>([]);

const trendColumns = computed<TableColumn<LogisticsKPIDashboardTrend>[]>(() => [
  { accessorKey: "dimensionKey", header: "维度" },
  { accessorKey: "totalWaybills", header: "总运单" },
  { accessorKey: "deliveredCount", header: "妥投数" },
  { accessorKey: "exceptionCount", header: "异常数" },
  { accessorKey: "onTimeRate", header: "准时率(%)" },
  { accessorKey: "deliverySuccessRate", header: "妥投率(%)" },
  { accessorKey: "avgTransitHours", header: "平均耗时(h)" },
  { accessorKey: "avgCost", header: "单均成本" },
]);

const drilldownColumns = computed<TableColumn<LogisticsKPIDashboardDrilldown>[]>(() => [
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "warehouseId", header: "仓库" },
  { accessorKey: "destinationZone", header: "区域" },
  { accessorKey: "elapsedHours", header: "耗时(h)" },
  { accessorKey: "timeoutRiskLevel", header: "超时风险" },
  { accessorKey: "costAmount", header: "成本" },
]);

const buildQuery = () => ({
  dimension: filters.dimension,
  window_hours: Number(filters.windowHours || 24),
  carrier_id: filters.carrierId || undefined,
  warehouse_id: filters.warehouseId || undefined,
  destination_zone: filters.destinationZone || undefined,
});

const loadAll = async () => {
  loading.value = true;
  try {
    const query = buildQuery();
    const [ov, trendRows, drillRows] = await Promise.all([
      logisticsApi.getKPIDashboardOverview(query),
      logisticsApi.getKPIDashboardTrends(query),
      logisticsApi.getKPIDashboardDrilldown({ ...query, limit: 100 }),
    ]);
    overview.value = ov;
    trends.value = trendRows;
    drilldowns.value = drillRows;
  } finally {
    loading.value = false;
  }
};

const exportCSV = async () => {
  exporting.value = true;
  try {
    const payload = await logisticsApi.exportKPIDashboard(buildQuery());
    const blob = new Blob([payload.content], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `logistics-kpi-${Date.now()}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  } finally {
    exporting.value = false;
  }
};

onMounted(() => {
  loadAll();
});
</script>
