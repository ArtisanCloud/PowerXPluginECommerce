<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">面单打印</h1>
        <p class="text-gray-500 dark:text-gray-400">支持批量打印、失败回执与补打重试队列。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" @click="loadData">刷新</UButton>
        <UButton color="warning" icon="i-heroicons-arrow-path-rounded-square" @click="retryFailed">
          重试失败任务
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">批量打印</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">输入运单 ID（逗号分隔），执行批量打印或补打。</p>
          </div>
          <USelect v-model="statusFilter" class="w-40" :options="statusOptions" />
        </div>
      </template>

      <div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
        <UInput v-model="form.waybillIdsText" class="lg:col-span-2" placeholder="waybill_id_1,waybill_id_2" />
        <UInput v-model="form.idempotencyKey" placeholder="幂等键（可选）" />
        <UInput v-model="form.reprintReason" placeholder="补打原因（可选）" />
      </div>

      <div class="mt-3 flex flex-wrap gap-2">
        <UButton color="primary" icon="i-heroicons-printer" @click="batchPrint">批量打印</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">打印任务回执</h3>
          <UBadge color="info" variant="subtle">{{ filteredTasks.length }} 条</UBadge>
        </div>
      </template>
      <UTable :columns="columns" :data="filteredTasks">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #attemptCount-cell="{ row }">
          <span>{{ row.original.attemptCount }}/{{ row.original.maxAttempts }}</span>
        </template>
        <template #lastError-cell="{ getValue }">
          <span class="text-warning-600">{{ getValue() || "-" }}</span>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi, type LogisticsLabelPrintTask } from "~/composables/api";

definePageMeta({
  name: "shipping-labels",
});

const logisticsApi = useLogisticsApi();
const tasks = ref<LogisticsLabelPrintTask[]>([]);
const statusFilter = ref("");

const form = reactive({
  waybillIdsText: "",
  idempotencyKey: "",
  reprintReason: "",
});

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "成功", value: "success" },
  { label: "失败", value: "failed" },
  { label: "待处理", value: "pending" },
];

const columns = computed<TableColumn<LogisticsLabelPrintTask>[]>(() => [
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "attemptCount", header: "尝试次数" },
  { accessorKey: "lastError", header: "失败原因" },
  { accessorKey: "printedAt", header: "打印时间" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const filteredTasks = computed(() =>
  tasks.value.filter((item) => !statusFilter.value || item.status === statusFilter.value),
);

const statusMeta = (status: string) => {
  switch (status) {
    case "success":
      return { label: "成功", color: "success" as const };
    case "failed":
      return { label: "失败", color: "error" as const };
    default:
      return { label: "待处理", color: "neutral" as const };
  }
};

const parseIDs = (raw: string) =>
  raw
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

const formatTime = (value: string) => (value ? value.replace("T", " ").slice(0, 16) : "-");

const loadData = async () => {
  const rows = await logisticsApi.listLabelPrintTasks({
    status: statusFilter.value || undefined,
  });
  tasks.value = rows.map((item) => ({
    ...item,
    printedAt: formatTime(item.printedAt),
    createdAt: formatTime(item.createdAt),
  }));
};

const batchPrint = async () => {
  const waybillIds = parseIDs(form.waybillIdsText);
  if (!waybillIds.length) return;
  await logisticsApi.batchPrintLabels({
    waybill_ids: waybillIds,
    idempotency_key: form.idempotencyKey || undefined,
    reprint_reason: form.reprintReason || undefined,
  });
  await loadData();
};

const retryFailed = async () => {
  await logisticsApi.retryLabelPrint({});
  await loadData();
};

watch(statusFilter, () => {
  loadData();
});

onMounted(() => {
  loadData();
});
</script>
