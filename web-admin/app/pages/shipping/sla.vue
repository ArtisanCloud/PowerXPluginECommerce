<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">承运商 SLA 看板</h1>
        <p class="text-gray-500 dark:text-gray-400">监控揽收时效、签收时效与异常率。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="carrierId" class="w-44" placeholder="承运商ID（可选）" />
        <UInput v-model.number="pickupSlaHours" class="w-36" type="number" placeholder="揽收 SLA(h)" />
        <UInput v-model.number="deliverySlaHours" class="w-36" type="number" placeholder="签收 SLA(h)" />
        <UButton color="info" variant="soft" icon="i-heroicons-bolt" :loading="syncJobLoading" @click="createSyncJob">
          创建批量同步任务
        </UButton>
        <UButton color="primary" icon="i-heroicons-arrow-path" @click="loadSnapshot">刷新</UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">网关健康</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">请求总数</p>
          <p class="mt-1 text-xl font-semibold">{{ gateway.summary.totalRequests }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">成功率</p>
          <p class="mt-1 text-xl font-semibold">{{ gateway.summary.successRate.toFixed(2) }}%</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">P95 延迟</p>
          <p class="mt-1 text-xl font-semibold">{{ gateway.summary.p95LatencyMS }} ms</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">失败请求</p>
          <p class="mt-1 text-xl font-semibold text-warning-600">{{ gateway.summary.failedRequests }}</p>
        </div>
      </div>
      <div class="mt-3 space-y-2">
        <p class="text-sm font-medium text-gray-900 dark:text-white">告警</p>
        <div v-if="!gateway.alerts.length" class="text-xs text-gray-500">暂无告警</div>
        <div v-for="item in gateway.alerts" :key="item.code" class="rounded border border-gray-200 p-2 text-xs dark:border-gray-800">
          <span class="font-semibold">{{ item.code }}</span> · {{ item.message }}
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">总体指标</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">运单总数</p>
          <p class="mt-1 text-xl font-semibold">{{ total.waybillCount }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">揽收准时率</p>
          <p class="mt-1 text-xl font-semibold">{{ total.pickupOnTimeRate.toFixed(2) }}%</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">签收准时率</p>
          <p class="mt-1 text-xl font-semibold">{{ total.signOnTimeRate.toFixed(2) }}%</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">异常率</p>
          <p class="mt-1 text-xl font-semibold text-warning-600">{{ total.exceptionRate.toFixed(2) }}%</p>
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">承运商分布</h3>
      </template>
      <UTable :columns="columns" :data="rows">
        <template #pickupOnTimeRate-cell="{ getValue }">
          <span>{{ Number(getValue()).toFixed(2) }}%</span>
        </template>
        <template #signOnTimeRate-cell="{ getValue }">
          <span>{{ Number(getValue()).toFixed(2) }}%</span>
        </template>
        <template #exceptionRate-cell="{ getValue }">
          <span class="text-warning-600">{{ Number(getValue()).toFixed(2) }}%</span>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi, type LogisticsGatewayHealthSnapshot, type LogisticsSLASummaryItem } from "~/composables/api";

definePageMeta({
  name: "shipping-sla",
});

const logisticsApi = useLogisticsApi();
const carrierId = ref("");
const pickupSlaHours = ref(24);
const deliverySlaHours = ref(72);
const rows = ref<LogisticsSLASummaryItem[]>([]);
const syncJobLoading = ref(false);
const gateway = ref<LogisticsGatewayHealthSnapshot>({
  summary: {
    windowHours: 24,
    totalRequests: 0,
    successRequests: 0,
    failedRequests: 0,
    successRate: 0,
    p95LatencyMS: 0,
  },
  carriers: [],
  alerts: [],
});
const total = ref<LogisticsSLASummaryItem>({
  carrierId: "all",
  carrierName: "全部承运商",
  waybillCount: 0,
  pickupOnTimeCount: 0,
  pickupOnTimeRate: 0,
  signOnTimeCount: 0,
  signOnTimeRate: 0,
  exceptionCount: 0,
  exceptionRate: 0,
  pickupSLAHours: 24,
  deliverySLAHours: 72,
});

const columns = computed<TableColumn<LogisticsSLASummaryItem>[]>(() => [
  { accessorKey: "carrierName", header: "承运商" },
  { accessorKey: "waybillCount", header: "运单数" },
  { accessorKey: "pickupOnTimeCount", header: "揽收准时单" },
  { accessorKey: "pickupOnTimeRate", header: "揽收准时率" },
  { accessorKey: "signOnTimeCount", header: "签收准时单" },
  { accessorKey: "signOnTimeRate", header: "签收准时率" },
  { accessorKey: "exceptionCount", header: "异常单数" },
  { accessorKey: "exceptionRate", header: "异常率" },
]);

const loadSnapshot = async () => {
  const snapshot = await logisticsApi.getSLADashboard({
    carrier_id: carrierId.value || undefined,
    pickup_sla_hours: pickupSlaHours.value || undefined,
    delivery_sla_hours: deliverySlaHours.value || undefined,
  });
  rows.value = snapshot.summary || [];
  total.value = snapshot.total || total.value;
  gateway.value = await logisticsApi.getGatewayHealth({ window_hours: 24 });
};

const createSyncJob = async () => {
  syncJobLoading.value = true;
  try {
    await logisticsApi.createTrackingSyncJob({
      carrier_id: carrierId.value || undefined,
      waybill_status: "in_transit",
      batch_limit: 20,
      event_limit: 20,
    });
    await loadSnapshot();
  } finally {
    syncJobLoading.value = false;
  }
};

onMounted(() => {
  loadSnapshot();
});
</script>
