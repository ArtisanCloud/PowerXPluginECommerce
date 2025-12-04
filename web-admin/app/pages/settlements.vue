<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">结算与分账</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理渠道对账、分账配置与结算进度，及时发现异常账期。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-down-tray">
          导出对账单
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">创建结算批次</UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-4">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较上周
        </p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">结算批次</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">按渠道、账期和状态筛选。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-56"
              placeholder="批次/渠道"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="channelFilter"
              class="w-44"
              :options="channelOptions"
              placeholder="全部渠道"
            />
            <USelect
              v-model="stateFilter"
              class="w-40"
              :options="stateOptions"
              placeholder="状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredBatches">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #amount-cell="{ getValue }">
          <span class="font-medium text-gray-900 dark:text-white">¥{{ getValue().toLocaleString() }}</span>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">详情</UButton>
            <UButton size="xs" variant="ghost" color="primary">下载账单</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">分账配置</h3>
            <UBadge color="info" variant="subtle">{{ splitRules.length }} 条</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="rule in splitRules"
            :key="rule.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ rule.name }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ rule.channel }}</p>
              </div>
              <UBadge color="primary" variant="subtle">{{ rule.mode }}</UBadge>
            </div>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ rule.desc }}</p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">编辑</UButton>
              <UButton size="xs" variant="ghost">复制</UButton>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">对账预警</h3>
            <UBadge color="warning" variant="subtle">{{ alerts.length }} 条</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="alert in alerts"
            :key="alert.id"
            class="rounded-xl border border-rose-100 p-4 dark:border-rose-900/40"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ alert.title }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ alert.time }}</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ alert.desc }}</p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">去处理</UButton>
              <UButton size="xs" variant="ghost">忽略</UButton>
            </div>
          </li>
        </ul>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "settlements",
});

type BatchStatus = "processing" | "completed" | "failed";

type SettlementBatch = {
  id: string;
  channel: string;
  period: string;
  amount: number;
  status: BatchStatus;
  updatedAt: string;
};

const batches = ref<SettlementBatch[]>([
  {
    id: "SET-20240208",
    channel: "天猫旗舰店",
    period: "2024/02/01 - 2024/02/07",
    amount: 1280000,
    status: "processing",
    updatedAt: "今天 10:12",
  },
  {
    id: "SET-20240205",
    channel: "京东自营",
    period: "2024/01/29 - 2024/02/04",
    amount: 980000,
    status: "completed",
    updatedAt: "昨天 18:40",
  },
  {
    id: "SET-20240131",
    channel: "抖音旗舰店",
    period: "2024/01/22 - 2024/01/28",
    amount: 305000,
    status: "failed",
    updatedAt: "2 天前",
  },
]);

const keyword = ref("");
const channelFilter = ref("");
const stateFilter = ref<BatchStatus | "">("");

const channelOptions = computed(() =>
  [{ label: "全部渠道", value: "" }].concat(
    Array.from(new Set(batches.value.map((batch) => batch.channel))).map((channel) => ({
      label: channel,
      value: channel,
    })),
  ),
);

const stateOptions = [
  { label: "全部状态", value: "" },
  { label: "结算中", value: "processing" },
  { label: "已完成", value: "completed" },
  { label: "异常", value: "failed" },
];

const columns = computed<TableColumn<SettlementBatch>[]>(() => [
  { accessorKey: "id", header: "批次号" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "period", header: "账期" },
  { accessorKey: "amount", header: "结算金额" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const filteredBatches = computed(() =>
  batches.value.filter((batch) => {
    const matchesKeyword = !keyword.value || batch.id.includes(keyword.value);
    const matchesChannel = !channelFilter.value || batch.channel === channelFilter.value;
    const matchesStatus = !stateFilter.value || batch.status === stateFilter.value;
    return matchesKeyword && matchesChannel && matchesStatus;
  }),
);

const statusMeta = (status: BatchStatus | "") => {
  switch (status) {
    case "processing":
      return { label: "结算中", color: "info" as const };
    case "completed":
      return { label: "已完成", color: "success" as const };
    case "failed":
      return { label: "异常", color: "error" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => [
  {
    title: "本周待结算",
    value: "¥" + (batches.value[0]?.amount ?? 0).toLocaleString(),
    trend: 2.4,
  },
  {
    title: "在途批次",
    value: batches.value.filter((b) => b.status === "processing").length,
    trend: 1.1,
  },
  {
    title: "异常批次",
    value: batches.value.filter((b) => b.status === "failed").length,
    trend: -0.8,
  },
  {
    title: "平均账期",
    value: "T+7",
    trend: 0.2,
  },
]);

const splitRules = ref([
  {
    id: "RULE-01",
    name: "直播分账（主播+仓储）",
    channel: "抖音旗舰店",
    mode: "比例",
    desc: "主播 5%，仓储服务 2%，支付完成后 T+1 分账。",
  },
  {
    id: "RULE-02",
    name: "跨境保税仓分账",
    channel: "天猫国际",
    mode: "阶梯",
    desc: "固定成本 ¥15 + 销售额 3%，满足 KPI 后降至 2.5%。",
  },
]);

const alerts = ref([
  {
    id: "ALERT-SET-1",
    title: "抖音批次差额预警",
    desc: "平台回传金额较本地订单少 ¥12,300，需核对退款记录。",
    time: "今天 09:15",
  },
  {
    id: "ALERT-SET-2",
    title: "京东账期延长",
    desc: "因年节物流延迟，账期顺延至 T+10，请关注资金计划。",
    time: "昨天 18:00",
  },
]);
</script>
