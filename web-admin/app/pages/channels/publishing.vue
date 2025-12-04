<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">上架发布</h1>
        <p class="text-gray-500 dark:text-gray-400">
          统一管理渠道发布任务、审核进度与风险提醒。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-down-tray">
          导出任务
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          创建发布任务
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-3">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较昨日
        </p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">发布任务</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              跟踪各渠道发布、审核、上线状态。
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-52"
              placeholder="商品/任务 ID"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="channelFilter"
              class="w-40"
              :options="channelOptions"
              placeholder="全部渠道"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :options="statusOptions"
              placeholder="任务状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredTasks">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #progress-cell="{ getValue }">
          <div class="flex items-center gap-2">
            <UProgress :value="getValue()" size="xs" class="flex-1" />
            <span class="text-xs text-gray-500">{{ getValue() }}%</span>
          </div>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">详情</UButton>
            <UButton size="xs" variant="ghost" color="primary">同步结果</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">发布阶段</h3>
            <UBadge color="info" variant="subtle">自动化流程</UBadge>
          </div>
        </template>
        <ol class="space-y-4">
          <li
            v-for="step in publishingSteps"
            :key="step.id"
            class="flex gap-3 rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div
              class="flex h-10 w-10 items-center justify-center rounded-full text-sm font-semibold"
              :class="
                step.done
                  ? 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/40 dark:text-emerald-300'
                  : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
              "
            >
              {{ step.id }}
            </div>
            <div>
              <p class="font-medium text-gray-900 dark:text-white">{{ step.title }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ step.desc }}</p>
            </div>
          </li>
        </ol>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">异常提醒</h3>
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
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ alert.message }}</p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">查看任务</UButton>
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
  name: "channels-publishing",
});

type PublishingStatus = "draft" | "syncing" | "pending" | "approved" | "failed";

type PublishingTask = {
  id: string;
  product: string;
  channel: string;
  status: PublishingStatus;
  progress: number;
  owner: string;
  plannedOnlineAt: string;
};

const tasks = ref<PublishingTask[]>([
  {
    id: "PUB-202402-01",
    product: "iPhone 15 Pro",
    channel: "天猫旗舰店",
    status: "pending",
    progress: 72,
    owner: "王帆",
    plannedOnlineAt: "2024-02-14",
  },
  {
    id: "PUB-202402-02",
    product: "Dyson V15",
    channel: "京东自营",
    status: "approved",
    progress: 100,
    owner: "陈曦",
    plannedOnlineAt: "2024-02-10",
  },
  {
    id: "PUB-202402-03",
    product: "Nintendo Switch",
    channel: "抖音旗舰店",
    status: "failed",
    progress: 38,
    owner: "李倩",
    plannedOnlineAt: "2024-02-12",
  },
  {
    id: "PUB-202402-04",
    product: "AirPods Pro",
    channel: "小红书店铺",
    status: "syncing",
    progress: 55,
    owner: "周楠",
    plannedOnlineAt: "2024-02-15",
  },
]);

const keyword = ref("");
const channelFilter = ref("");
const statusFilter = ref<PublishingStatus | "">("");

const channelOptions = computed(() =>
  [{ label: "全部渠道", value: "" }].concat(
    Array.from(new Set(tasks.value.map((task) => task.channel))).map((channel) => ({
      label: channel,
      value: channel,
    })),
  ),
);

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "草稿", value: "draft" },
  { label: "同步中", value: "syncing" },
  { label: "待审核", value: "pending" },
  { label: "审核通过", value: "approved" },
  { label: "失败", value: "failed" },
];

const columns = computed<TableColumn<PublishingTask>[]>(() => [
  { accessorKey: "id", header: "任务号" },
  { accessorKey: "product", header: "商品" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "progress", header: "进度" },
  { accessorKey: "owner", header: "负责人" },
  { accessorKey: "plannedOnlineAt", header: "计划上线" },
  { id: "actions", header: "操作" },
]);

const filteredTasks = computed(() =>
  tasks.value.filter((task) => {
    const matchesKeyword =
      !keyword.value ||
      task.product.toLowerCase().includes(keyword.value.toLowerCase()) ||
      task.id.includes(keyword.value);
    const matchesChannel = !channelFilter.value || task.channel === channelFilter.value;
    const matchesStatus = !statusFilter.value || task.status === statusFilter.value;
    return matchesKeyword && matchesChannel && matchesStatus;
  }),
);

const statusMeta = (status: PublishingStatus | "") => {
  switch (status) {
    case "draft":
      return { label: "草稿", color: "neutral" as const };
    case "syncing":
      return { label: "同步中", color: "info" as const };
    case "pending":
      return { label: "待审核", color: "warning" as const };
    case "approved":
      return { label: "已上线", color: "success" as const };
    case "failed":
      return { label: "失败", color: "error" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => [
  { title: "待审核任务", value: tasks.value.filter((task) => task.status === "pending").length, trend: 2.1 },
  { title: "今日上线", value: tasks.value.filter((task) => task.status === "approved").length, trend: 0.5 },
  { title: "异常任务", value: tasks.value.filter((task) => task.status === "failed").length, trend: -1.8 },
]);

const publishingSteps = ref([
  { id: 1, title: "素材检查", desc: "确认素材、文案、类目均符合平台规范", done: true },
  { id: 2, title: "价格同步", desc: "比对价格策略，提交渠道价变更", done: true },
  { id: 3, title: "发布与审核", desc: "调用渠道 API，等待审核通过", done: false },
  { id: 4, title: "结果回传", desc: "同步上线结果，通知运营同学", done: false },
]);

const alerts = ref([
  {
    id: "ALERT-1",
    title: "抖音任务审核失败",
    message: "SKU 12345 因关键词违规审核未通过，需修改文案重提。",
    time: "今天 11:20",
  },
  {
    id: "ALERT-2",
    title: "小红书接口限流",
    message: "平台返回 429，请在 30 分钟后重新推送。",
    time: "昨天 22:45",
  },
]);
</script>
