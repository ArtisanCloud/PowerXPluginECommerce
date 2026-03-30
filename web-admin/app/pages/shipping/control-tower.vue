<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约控制塔</h1>
        <p class="text-gray-500 dark:text-gray-400">统一监控在途、异常、超时、SLA 与成本信号。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="filters.carrierId" class="w-40" placeholder="承运商 ID" />
        <UInput v-model="filters.warehouseId" class="w-40" placeholder="仓库 ID" />
        <UInput v-model="filters.destinationZone" class="w-40" placeholder="目的区域" />
        <USelect v-model="filters.status" class="w-36" :options="statusOptions" />
        <UInput v-model.number="filters.windowHours" class="w-32" type="number" placeholder="窗口(h)" />
        <UButton color="primary" icon="i-heroicons-arrow-path" :loading="loading" @click="loadAll">刷新</UButton>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-5">
      <UCard>
        <p class="text-xs text-gray-500">总运单</p>
        <p class="mt-1 text-xl font-semibold">{{ overview.summary.totalWaybills }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">在途</p>
        <p class="mt-1 text-xl font-semibold">{{ overview.summary.inTransitCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">异常</p>
        <p class="mt-1 text-xl font-semibold text-warning-600">{{ overview.summary.exceptionCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">超时</p>
        <p class="mt-1 text-xl font-semibold text-error-600">{{ overview.summary.timeoutCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">准时率</p>
        <p class="mt-1 text-xl font-semibold">{{ overview.summary.onTimeRate.toFixed(2) }}%</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">告警与成本</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div class="rounded-lg border border-gray-200 p-3 dark:border-gray-800">
          <p class="text-xs text-gray-500">成本</p>
          <p class="mt-1 text-lg font-semibold">¥{{ overview.summary.totalCost.toFixed(2) }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-3 dark:border-gray-800">
          <p class="text-xs text-gray-500">告警数</p>
          <p class="mt-1 text-lg font-semibold">{{ overview.summary.alertCount }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-3 dark:border-gray-800">
          <p class="text-xs text-gray-500">窗口</p>
          <p class="mt-1 text-lg font-semibold">{{ overview.summary.windowHours }}h</p>
        </div>
      </div>
      <div class="mt-3 space-y-2 text-xs">
        <div v-if="!overview.alerts.length" class="text-gray-500">暂无告警</div>
        <div
          v-for="item in overview.alerts"
          :key="item.code"
          class="rounded border border-gray-200 p-2 dark:border-gray-800"
        >
          <span class="font-semibold">{{ item.code }}</span> · {{ item.message }}
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">异常钻取</h3>
      </template>
      <UTable :columns="columns" :data="drilldownRows">
        <template #costAmount-cell="{ getValue }">
          <span>¥{{ Number(getValue() || 0).toFixed(2) }}</span>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">告警订阅</h3>
      </template>
      <div class="grid gap-2 md:grid-cols-6">
        <UInput v-model="subForm.name" placeholder="订阅名称" />
        <UInput v-model.number="subForm.minOnTimeRate" type="number" placeholder="最小准时率" />
        <UInput v-model.number="subForm.maxTimeoutCount" type="number" placeholder="最大超时单量" />
        <UInput v-model.number="subForm.maxCostAmount" type="number" placeholder="最大成本" />
        <UInput v-model="subForm.carrierId" placeholder="承运商ID（可选）" />
        <UButton color="primary" :loading="submitting" @click="createSubscription">保存订阅</UButton>
      </div>
      <ul class="mt-3 space-y-2 text-xs">
        <li
          v-for="item in subscriptions"
          :key="item.id"
          class="rounded border border-gray-200 p-2 dark:border-gray-800"
        >
          <div class="flex items-center justify-between gap-2">
            <span>
              {{ item.name }} · 准时率≥{{ item.minOnTimeRate }}% · 超时≤{{ item.maxTimeoutCount }} · 成本≤¥{{ item.maxCostAmount }}
            </span>
            <UButton size="xs" variant="ghost" @click="toggleSubscription(item)">
              {{ item.enabled ? "停用" : "启用" }}
            </UButton>
          </div>
        </li>
        <li v-if="!subscriptions.length" class="text-gray-500">暂无订阅</li>
      </ul>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import {
  useLogisticsApi,
  type LogisticsControlTowerDrilldownItem,
  type LogisticsControlTowerOverview,
  type LogisticsControlTowerSubscription,
} from "~/composables/api";

definePageMeta({
  name: "shipping-control-tower",
});

const logisticsApi = useLogisticsApi();
const loading = ref(false);
const submitting = ref(false);

const filters = reactive({
  carrierId: "",
  warehouseId: "",
  destinationZone: "",
  status: "",
  windowHours: 24,
});

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "在途", value: "in_transit" },
  { label: "异常", value: "exception" },
  { label: "超时", value: "timeout" },
  { label: "已妥投", value: "delivered" },
];

const overview = ref<LogisticsControlTowerOverview>({
  summary: {
    windowHours: 24,
    totalWaybills: 0,
    inTransitCount: 0,
    exceptionCount: 0,
    timeoutCount: 0,
    deliveredCount: 0,
    onTimeRate: 0,
    totalCost: 0,
    alertCount: 0,
  },
  alerts: [],
});
const drilldownRows = ref<LogisticsControlTowerDrilldownItem[]>([]);
const subscriptions = ref<LogisticsControlTowerSubscription[]>([]);

const subForm = reactive({
  name: "",
  minOnTimeRate: 95,
  maxTimeoutCount: 5,
  maxCostAmount: 5000,
  carrierId: "",
});

const columns = computed<TableColumn<LogisticsControlTowerDrilldownItem>[]>(() => [
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "warehouseId", header: "仓库" },
  { accessorKey: "destinationZone", header: "目的区域" },
  { accessorKey: "elapsedHours", header: "耗时(h)" },
  { accessorKey: "timeoutRiskLevel", header: "超时风险" },
  { accessorKey: "costAmount", header: "成本" },
]);

const loadAll = async () => {
  loading.value = true;
  try {
    const query = {
      carrier_id: filters.carrierId || undefined,
      warehouse_id: filters.warehouseId || undefined,
      destination_zone: filters.destinationZone || undefined,
      window_hours: Number(filters.windowHours || 24),
    };
    overview.value = await logisticsApi.getControlTowerOverview(query);
    drilldownRows.value = await logisticsApi.getControlTowerDrilldown({
      ...query,
      status: filters.status || undefined,
      limit: 100,
    });
    subscriptions.value = await logisticsApi.listControlTowerSubscriptions();
  } finally {
    loading.value = false;
  }
};

const createSubscription = async () => {
  submitting.value = true;
  try {
    await logisticsApi.upsertControlTowerSubscription({
      name: subForm.name || "默认告警订阅",
      min_on_time_rate: Number(subForm.minOnTimeRate || 95),
      max_timeout_count: Number(subForm.maxTimeoutCount || 5),
      max_cost_amount: Number(subForm.maxCostAmount || 5000),
      carrier_id: subForm.carrierId || undefined,
      enabled: true,
    });
    await loadAll();
  } finally {
    submitting.value = false;
  }
};

const toggleSubscription = async (row: LogisticsControlTowerSubscription) => {
  submitting.value = true;
  try {
    await logisticsApi.upsertControlTowerSubscription({
      id: row.id,
      name: row.name,
      carrier_id: row.carrierId || undefined,
      warehouse_id: row.warehouseId || undefined,
      destination_zone: row.destinationZone || undefined,
      min_on_time_rate: row.minOnTimeRate,
      max_timeout_count: row.maxTimeoutCount,
      max_cost_amount: row.maxCostAmount,
      enabled: !row.enabled,
      config: row.config,
    });
    await loadAll();
  } finally {
    submitting.value = false;
  }
};

onMounted(() => {
  loadAll();
});
</script>
