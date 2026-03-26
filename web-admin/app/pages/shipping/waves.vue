<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约波次</h1>
        <p class="text-gray-500 dark:text-gray-400">
          聚合任务批量推进，支持部分失败回执与任务重分配。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="newWave.name" class="w-40" placeholder="波次名称" />
        <UInput v-model="newWave.warehouseId" class="w-36" placeholder="仓ID" />
        <UInput v-model="newWave.taskIdsText" class="w-80" placeholder="任务ID（逗号分隔）" />
        <UButton color="primary" icon="i-heroicons-plus" @click="createWave">创建波次</UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">波次列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">按仓库聚合任务并批量推进。</p>
          </div>
          <USelect v-model="statusFilter" class="w-44" :options="statusOptions" />
        </div>
      </template>
      <UTable :columns="columns" :data="filteredWaves">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" color="primary" @click="advanceWave(row.original.id)">
              批量推进
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="reassignFirstTask(row.original.id)">
              重分配首任务
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useFulfillmentApi } from "~/composables/api";

definePageMeta({
  name: "shipping-waves",
});

type WaveRow = {
  id: string;
  name: string;
  warehouseId: string;
  status: string;
  createdAt: string;
  updatedAt: string;
};

const fulfillmentApi = useFulfillmentApi();
const waves = ref<WaveRow[]>([]);
const statusFilter = ref("");

const newWave = reactive({
  name: "",
  warehouseId: "",
  taskIdsText: "",
});

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "待处理", value: "pending" },
  { label: "执行中", value: "running" },
  { label: "部分失败", value: "partially_failed" },
  { label: "失败", value: "failed" },
  { label: "已完成", value: "completed" },
];

const columns = computed<TableColumn<WaveRow>[]>(() => [
  { accessorKey: "name", header: "波次名称" },
  { accessorKey: "warehouseId", header: "仓ID" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const filteredWaves = computed(() =>
  waves.value.filter((item) => !statusFilter.value || item.status === statusFilter.value),
);

const statusMeta = (status: string) => {
  switch (status) {
    case "completed":
      return { label: "已完成", color: "success" as const };
    case "partially_failed":
      return { label: "部分失败", color: "warning" as const };
    case "failed":
      return { label: "失败", color: "error" as const };
    case "running":
      return { label: "执行中", color: "info" as const };
    default:
      return { label: "待处理", color: "neutral" as const };
  }
};

const formatTime = (value: string) => (value ? value.replace("T", " ").slice(0, 16) : "-");

const loadWaves = async () => {
  const rows = await fulfillmentApi.listWaves();
  waves.value = rows.map((item) => ({
    id: item.id,
    name: item.name || item.id,
    warehouseId: item.warehouseId,
    status: item.status,
    createdAt: formatTime(item.createdAt),
    updatedAt: formatTime(item.updatedAt),
  }));
};

const parseTaskIDs = (raw: string) =>
  raw
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

const createWave = async () => {
  const taskIds = parseTaskIDs(newWave.taskIdsText);
  if (!newWave.warehouseId || taskIds.length === 0) return;
  await fulfillmentApi.createWave({
    name: newWave.name,
    warehouse_id: newWave.warehouseId,
    task_ids: taskIds,
  });
  newWave.name = "";
  newWave.warehouseId = "";
  newWave.taskIdsText = "";
  await loadWaves();
};

const advanceWave = async (waveId: string) => {
  await fulfillmentApi.advanceWave(waveId, {
    status: "picking",
    operator_id: "admin",
  });
  await loadWaves();
};

const reassignFirstTask = async (waveId: string) => {
  const waveDetail = await fulfillmentApi.getWaveDetail(waveId);
  const taskID = waveDetail.tasks[0]?.taskId;
  if (!taskID) return;
  await fulfillmentApi.reassignWaveTask(waveId, taskID, {
    assigned_to: "picker-reassigned",
    operator_id: "admin",
  });
  await loadWaves();
};

onMounted(() => {
  loadWaves();
});
</script>
