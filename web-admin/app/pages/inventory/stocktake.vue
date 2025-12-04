<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">库存盘点</h1>
        <p class="text-gray-500 dark:text-gray-400">
          统一管理周期盘点、抽盘与差异复核，确保账实一致。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-document-arrow-down">
          下载模板
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          发起盘点任务
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in stocktakeStats" :key="card.title">
        <div class="space-y-1">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
          <div class="text-3xl font-semibold text-gray-900 dark:text-white">
            {{ card.value }}
          </div>
          <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
            {{ card.trend >= 0 ? "+" : "" }}{{ card.trend }}% 环比
          </p>
        </div>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-3">
      <UCard class="lg:col-span-2">
        <template #header>
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">盘点任务</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                根据优先级排序，支持查看差异率。
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <USelect class="w-32" v-model="statusFilter" :options="statusOptions" />
              <USelect class="w-40" v-model="cycleFilter" :options="cycleOptions" />
              <UInput
                v-model="taskKeyword"
                class="w-48"
                placeholder="搜索仓库/负责人"
                icon="i-heroicons-magnifying-glass"
              />
            </div>
          </div>
        </template>

        <UTable :data="filteredTasks" :columns="columns">
          <template #status-cell="{ getValue }">
            <UBadge :color="statusMeta(getValue()).color" variant="subtle">
              {{ statusMeta(getValue()).label }}
            </UBadge>
          </template>
          <template #progress-cell="{ getValue }">
            <div>
              <div class="flex items-center justify-between text-xs text-gray-500">
                <span>{{ getValue() }}%</span>
                <span>100%</span>
              </div>
              <UProgress :value="getValue()" size="xs" />
            </div>
          </template>
          <template #variance-cell="{ getValue }">
            <span :class="getValue() > 0.5 ? 'text-rose-500' : 'text-emerald-600'">
              {{ getValue() }}%
            </span>
          </template>
          <template #actions-cell>
            <div class="flex gap-2">
              <UButton size="xs" variant="ghost">查看</UButton>
              <UButton size="xs" variant="ghost" color="primary">导出差异</UButton>
            </div>
          </template>
        </UTable>
      </UCard>

      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">盘点日历</h3>
        </template>

        <ul class="space-y-4">
          <li
            v-for="event in upcomingEvents"
            :key="event.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">{{ event.title }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ event.warehouse }}</p>
              </div>
              <UBadge variant="subtle" :color="statusMeta(event.status).color">
                {{ statusMeta(event.status).label }}
              </UBadge>
            </div>
            <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">
              {{ event.date }} · 负责人 {{ event.owner }}
            </p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">提醒</UButton>
              <UButton size="xs" variant="ghost" color="neutral">详情</UButton>
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
  name: "inventory-stocktake",
});

type TaskStatus = "draft" | "in-progress" | "review" | "done";

type StocktakeTask = {
  id: string;
  warehouse: string;
  cycle: string;
  scheduledDate: string;
  owner: string;
  status: TaskStatus;
  progress: number;
  variance: number;
  type: "全盘" | "抽盘";
};

const tasks = ref<StocktakeTask[]>([
  {
    id: "ST-202402-01",
    warehouse: "北京顺义中心仓",
    cycle: "月度",
    scheduledDate: "2024-02-15",
    owner: "刘宇",
    status: "in-progress",
    progress: 68,
    variance: 0.34,
    type: "全盘",
  },
  {
    id: "ST-202402-02",
    warehouse: "上海青浦保税仓",
    cycle: "月度",
    scheduledDate: "2024-02-12",
    owner: "林晓",
    status: "review",
    progress: 100,
    variance: 0.82,
    type: "全盘",
  },
  {
    id: "ST-202402-03",
    warehouse: "深圳宝安前置仓",
    cycle: "周度",
    scheduledDate: "2024-02-10",
    owner: "王昊",
    status: "done",
    progress: 100,
    variance: 0.22,
    type: "抽盘",
  },
  {
    id: "ST-202402-04",
    warehouse: "广州黄埔前置仓",
    cycle: "周度",
    scheduledDate: "2024-02-09",
    owner: "陈敏",
    status: "draft",
    progress: 15,
    variance: 0,
    type: "抽盘",
  },
]);

const taskKeyword = ref("");
const statusFilter = ref<TaskStatus | "">("");
const cycleFilter = ref("");

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "待启动", value: "draft" },
  { label: "执行中", value: "in-progress" },
  { label: "差异复核", value: "review" },
  { label: "已完成", value: "done" },
];

const cycleOptions = [
  { label: "全部周期", value: "" },
  { label: "月度", value: "月度" },
  { label: "周度", value: "周度" },
];

const filteredTasks = computed(() =>
  tasks.value.filter((task) => {
    const matchesKeyword =
      !taskKeyword.value ||
      task.warehouse.includes(taskKeyword.value) ||
      task.owner.includes(taskKeyword.value);
    const matchesStatus =
      !statusFilter.value || task.status === statusFilter.value;
    const matchesCycle = !cycleFilter.value || task.cycle === cycleFilter.value;
    return matchesKeyword && matchesStatus && matchesCycle;
  }),
);

const columns = computed<TableColumn<StocktakeTask>[]>(() => [
  { accessorKey: "id", header: "任务编号" },
  { accessorKey: "warehouse", header: "仓库" },
  { accessorKey: "cycle", header: "周期" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "scheduledDate", header: "计划日期" },
  { accessorKey: "owner", header: "负责人" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "progress", header: "进度" },
  { accessorKey: "variance", header: "差异率" },
  { id: "actions", header: "操作" },
]);

const statusMeta = (status: TaskStatus | "") => {
  switch (status) {
    case "draft":
      return { label: "待启动", color: "neutral" as const };
    case "in-progress":
      return { label: "执行中", color: "info" as const };
    case "review":
      return { label: "差异复核", color: "warning" as const };
    case "done":
      return { label: "已完成", color: "success" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const stocktakeStats = computed(() => [
  {
    title: "本月计划任务",
    value: tasks.value.length,
    trend: 5.4,
  },
  {
    title: "平均差异率",
    value:
      (
        tasks.value.reduce((sum, task) => sum + task.variance, 0) /
        tasks.value.length
      ).toFixed(2) + "%",
    trend: -0.6,
  },
  {
    title: "待复核任务",
    value: tasks.value.filter((task) => task.status === "review").length,
    trend: 3.1,
  },
]);

const upcomingEvents = ref([
  {
    id: "EV-01",
    title: "华东区域抽盘",
    warehouse: "上海青浦保税仓",
    date: "2 月 12 日 10:00",
    owner: "林晓",
    status: "review" as TaskStatus,
  },
  {
    id: "EV-02",
    title: "华北月度盘点",
    warehouse: "北京顺义中心仓",
    date: "2 月 15 日 09:30",
    owner: "刘宇",
    status: "in-progress" as TaskStatus,
  },
  {
    id: "EV-03",
    title: "华南抽盘",
    warehouse: "深圳宝安前置仓",
    date: "2 月 18 日 14:00",
    owner: "王昊",
    status: "draft" as TaskStatus,
  },
]);
</script>
