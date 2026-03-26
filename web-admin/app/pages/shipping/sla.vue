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
        <UButton color="primary" icon="i-heroicons-arrow-path" @click="loadSnapshot">刷新</UButton>
      </div>
    </div>

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
import { useLogisticsApi, type LogisticsSLASummaryItem } from "~/composables/api";

definePageMeta({
  name: "shipping-sla",
});

const logisticsApi = useLogisticsApi();
const carrierId = ref("");
const pickupSlaHours = ref(24);
const deliverySlaHours = ref(72);
const rows = ref<LogisticsSLASummaryItem[]>([]);
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
};

onMounted(() => {
  loadSnapshot();
});
</script>
