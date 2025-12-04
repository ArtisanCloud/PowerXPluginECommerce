<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">转化与增长</h1>
        <p class="text-gray-500 dark:text-gray-400">
          追踪漏斗、留存与营销转化，为增长策略提供依据。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <USelect v-model="timeframe" :options="timeframeOptions" class="w-44" />
        <USelect v-model="segment" :options="segmentOptions" class="w-48" />
        <UButton variant="ghost" color="neutral" icon="i-heroicons-arrow-path">
          更新
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
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}%
        </p>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">交易转化漏斗</h3>
        </template>
        <div class="space-y-4">
          <div
            v-for="stage in funnel"
            :key="stage.name"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ stage.name }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ stage.count }}</span>
            </div>
            <div class="mt-2 flex items-center gap-3">
              <div class="flex-1 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-800">
                <div
                  class="h-2 rounded-full bg-primary-500"
                  :style="{ width: stage.rate + '%' }"
                />
              </div>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ stage.rate }}%</span>
            </div>
          </div>
        </div>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">用户留存</h3>
            <span class="text-sm text-gray-500 dark:text-gray-400">按 {{ timeframeLabel }}</span>
          </div>
        </template>
        <UTable :columns="retentionColumns" :data="retentionData">
          <template #cohort-cell="{ row }">
            <span class="font-medium text-gray-900 dark:text-white">{{ row.original.cohort }}</span>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ row.original.size }} 用户</p>
          </template>
          <template #day7-cell="{ getValue }">
            {{ getValue() }}%
          </template>
          <template #day30-cell="{ getValue }">
            {{ getValue() }}%
          </template>
        </UTable>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">营销活动</h3>
            <USelect v-model="campaignFilter" :options="campaignOptions" class="w-44" />
          </div>
        </template>
        <UTable :columns="campaignColumns" :data="filteredCampaigns">
          <template #cvrs-cell="{ getValue }">
            {{ getValue() }}%
          </template>
          <template #roi-cell="{ getValue }">
            {{ getValue() }}x
          </template>
        </UTable>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">增长任务提醒</h3>
            <UBadge color="warning" variant="subtle">{{ tasks.length }} 条</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="task in tasks"
            :key="task.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ task.title }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ task.due }}</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ task.desc }}</p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">查看</UButton>
              <UButton size="xs" variant="ghost">标记完成</UButton>
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
  name: "reports-conversion",
});

const timeframe = ref("30d");
const segment = ref("all");

const timeframeOptions = [
  { label: "近 7 天", value: "7d" },
  { label: "近 30 天", value: "30d" },
  { label: "近 90 天", value: "90d" },
];

const segmentOptions = [
  { label: "全部用户", value: "all" },
  { label: "新客", value: "new" },
  { label: "会员", value: "vip" },
  { label: "复购用户", value: "repeat" },
];

const summaryCards = computed(() => [
  { title: "整体转化率", value: "3.2%", trend: 0.6 },
  { title: "新客首购率", value: "42%", trend: -1.1 },
  { title: "会员复购率", value: "58%", trend: 3.4 },
  { title: "平均获客成本", value: "¥ 126", trend: -2.5 },
]);

const timeframeLabel = computed(
  () => timeframeOptions.find((item) => item.value === timeframe.value)?.label || "",
);

const funnel = computed(() => [
  { name: "访客数", count: "1,260,000", rate: 100 },
  { name: "加入购物车", count: "186,000", rate: 14.8 },
  { name: "提交订单", count: "54,300", rate: 4.3 },
  { name: "成功支付", count: "40,320", rate: 3.2 },
]);

type RetentionRow = {
  cohort: string;
  size: number;
  day7: number;
  day30: number;
};

const retentionData = ref<RetentionRow[]>([
  { cohort: "Jan W4", size: 12300, day7: 32, day30: 18 },
  { cohort: "Feb W1", size: 9800, day7: 30, day30: 17 },
  { cohort: "Feb W2", size: 11200, day7: 34, day30: 20 },
]);

const retentionColumns = computed<TableColumn<RetentionRow>[]>(() => [
  { accessorKey: "cohort", header: "Cohort" },
  { accessorKey: "day7", header: "7 日留存" },
  { accessorKey: "day30", header: "30 日留存" },
]);

type Campaign = {
  id: string;
  name: string;
  channel: string;
  cvrs: number;
  roi: number;
};

const campaigns = ref<Campaign[]>([
  { id: "CP-01", name: "女王节拉新", channel: "抖音投放", cvrs: 2.8, roi: 3.5 },
  { id: "CP-02", name: "京东会员日", channel: "京东站内", cvrs: 3.6, roi: 4.2 },
  { id: "CP-03", name: "APP Push 联动", channel: "自有渠道", cvrs: 1.9, roi: 2.4 },
]);

const campaignOptions = computed(() =>
  [{ label: "全部渠道", value: "" }].concat(
    Array.from(new Set(campaigns.value.map((it) => it.channel))).map((channel) => ({
      label: channel,
      value: channel,
    })),
  ),
);
const campaignFilter = ref("");

const campaignColumns = computed<TableColumn<Campaign>[]>(() => [
  { accessorKey: "name", header: "活动" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "cvrs", header: "转化率" },
  { accessorKey: "roi", header: "ROI" },
]);

const filteredCampaigns = computed(() =>
  campaigns.value.filter((campaign) => !campaignFilter.value || campaign.channel === campaignFilter.value),
);

const tasks = ref([
  {
    id: "TASK-01",
    title: "优化会员成长任务",
    desc: "日活低于目标 8%，调整激励机制。",
    due: "今天",
  },
  {
    id: "TASK-02",
    title: "补齐留存推送文案",
    desc: "针对 Feb W2 cohort 制定两轮 A/B 内容。",
    due: "周五",
  },
]);
</script>
