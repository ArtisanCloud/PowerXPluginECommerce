<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">销售报表</h1>
        <p class="text-gray-500 dark:text-gray-400">
          聚合 GMV、订单与渠道贡献，支持按时间段与渠道筛选。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <USelect v-model="range" :options="rangeOptions" class="w-44" />
        <USelect v-model="channel" :options="channelOptions" class="w-48" />
        <UButton variant="ghost" color="neutral" icon="i-heroicons-arrow-path">
          刷新
        </UButton>
        <UButton color="primary" icon="i-heroicons-document-arrow-down">
          导出
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较上周
        </p>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">GMV 趋势</h3>
            <span class="text-sm text-gray-500 dark:text-gray-400">{{ rangeLabel }}</span>
          </div>
        </template>
        <div class="h-64 w-full rounded-lg bg-gradient-to-r from-primary-50 to-primary-100 dark:from-primary-900/40 dark:to-primary-900/20 p-6 text-sm text-gray-500 dark:text-gray-400">
          模拟折线图：近 14 天 GMV 变化
        </div>
      </UCard>

      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道贡献</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="channel in channelBreakdown"
            :key="channel.name"
            class="space-y-1"
          >
            <div class="flex justify-between text-sm text-gray-500 dark:text-gray-400">
              <span>{{ channel.name }}</span>
              <span>{{ channel.value }} · {{ channel.percent }}%</span>
            </div>
            <UProgress :value="channel.percent" size="xs" />
          </div>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">订单明细</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">按渠道、客单价与转化率对比。</p>
          </div>
          <UInput
            v-model="orderKeyword"
            class="w-64"
            placeholder="搜索渠道/来源"
            icon="i-heroicons-magnifying-glass"
          />
        </div>
      </template>

      <UTable :columns="orderColumns" :data="filteredOrders">
        <template #avgOrderValue-cell="{ getValue }">
          ¥{{ getValue().toLocaleString() }}
        </template>
        <template #conversionRate-cell="{ getValue }">
          {{ getValue() }}%
        </template>
        <template #trend-cell="{ getValue }">
          <UBadge :color="getValue() >= 0 ? 'success' : 'error'" variant="subtle">
            {{ getValue() >= 0 ? '+' : '' }}{{ getValue() }}%
          </UBadge>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "reports-sales",
});

const range = ref("7d");
const channel = ref("all");
const rangeOptions = [
  { label: "近 7 天", value: "7d" },
  { label: "近 14 天", value: "14d" },
  { label: "近 30 天", value: "30d" },
];
const channelOptions = [
  { label: "全部渠道", value: "all" },
  { label: "天猫旗舰店", value: "tmall" },
  { label: "京东自营", value: "jd" },
  { label: "抖音旗舰店", value: "douyin" },
];

const summaryCards = computed(() => [
  { title: "GMV", value: "¥3.2M", trend: 8.4 },
  { title: "订单量", value: "41,230", trend: 5.1 },
  { title: "客单价", value: "¥780", trend: 2.6 },
  { title: "退款率", value: "2.1%", trend: -0.3 },
]);

const rangeLabel = computed(() => rangeOptions.find((item) => item.value === range.value)?.label);

const channelBreakdown = computed(() => [
  { name: "天猫旗舰店", value: "¥1.3M", percent: 41 },
  { name: "京东自营", value: "¥960K", percent: 30 },
  { name: "抖音旗舰店", value: "¥560K", percent: 18 },
  { name: "小红书旗舰店", value: "¥240K", percent: 7 },
  { name: "其他", value: "¥140K", percent: 4 },
]);

type OrderRow = {
  channel: string;
  orders: number;
  avgOrderValue: number;
  conversionRate: number;
  trend: number;
};

const orderData = ref<OrderRow[]>([
  { channel: "天猫旗舰店", orders: 16890, avgOrderValue: 780, conversionRate: 3.5, trend: 4.6 },
  { channel: "京东自营", orders: 12030, avgOrderValue: 820, conversionRate: 4.1, trend: 2.3 },
  { channel: "抖音旗舰店", orders: 8120, avgOrderValue: 420, conversionRate: 2.8, trend: -1.2 },
  { channel: "小红书旗舰店", orders: 4190, avgOrderValue: 360, conversionRate: 1.9, trend: 0.8 },
]);

const orderKeyword = ref("");

const orderColumns = computed<TableColumn<OrderRow>[]>(() => [
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "orders", header: "订单量" },
  { accessorKey: "avgOrderValue", header: "客单价" },
  { accessorKey: "conversionRate", header: "转化率" },
  { accessorKey: "trend", header: "环比" },
]);

const filteredOrders = computed(() =>
  orderData.value.filter((row) =>
    row.channel.toLowerCase().includes(orderKeyword.value.trim().toLowerCase()),
  ),
);
</script>
