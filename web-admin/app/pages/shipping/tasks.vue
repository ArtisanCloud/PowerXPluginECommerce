<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约任务</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理仓内任务状态并登记异常，支持自动升级追踪。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-queue-list" to="/shipping/waves">
          波次管理
        </UButton>
        <UInput v-model="newTask.orderId" class="w-40" placeholder="订单ID" />
        <UInput v-model="newTask.warehouseId" class="w-40" placeholder="仓ID" />
        <UButton color="primary" icon="i-heroicons-plus" @click="createTask">
          新建任务
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">任务看板</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">支持按状态筛选并一键完成。</p>
          </div>
          <USelect v-model="statusFilter" class="w-44" :options="statusOptions" />
        </div>
      </template>

      <UTable :columns="taskColumns" :data="filteredTasks">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              size="xs"
              variant="ghost"
              color="primary"
              :disabled="row.original.status === 'completed'"
              @click="completeTask(row.original.id)"
            >
              完成
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="prepareException(row.original)">
              报异常
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">异常记录</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">Open/已升级异常一览。</p>
          </div>
          <UBadge color="warning" variant="subtle">
            {{ exceptions.length }} 条
          </UBadge>
        </div>
      </template>

      <UTable :columns="exceptionColumns" :data="exceptions">
        <template #status-cell="{ getValue }">
          <UBadge :color="getValue() === 'escalated' ? 'error' : 'warning'" variant="subtle">
            {{ getValue() === "escalated" ? "已升级" : "处理中" }}
          </UBadge>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useFulfillmentApi } from "~/composables/api";

definePageMeta({
  name: "shipping-tasks",
});

type TaskRow = {
  id: string;
  orderId: string;
  warehouseId: string;
  status: string;
  assignedTo: string;
  updatedAt: string;
};

const fulfillmentApi = useFulfillmentApi();

const tasks = ref<TaskRow[]>([]);
const exceptions = ref<any[]>([]);
const statusFilter = ref("");

const newTask = reactive({
  orderId: "",
  warehouseId: "",
});

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "待处理", value: "pending" },
  { label: "拣货中", value: "picking" },
  { label: "已打包", value: "packed" },
  { label: "已完成", value: "completed" },
  { label: "异常", value: "exception" },
];

const taskColumns = computed<TableColumn<TaskRow>[]>(() => [
  { accessorKey: "orderId", header: "订单ID" },
  { accessorKey: "warehouseId", header: "仓ID" },
  { accessorKey: "assignedTo", header: "负责人" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const exceptionColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "taskId", header: "任务ID" },
  { accessorKey: "type", header: "异常类型" },
  { accessorKey: "reason", header: "原因" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
]);

const filteredTasks = computed(() =>
  tasks.value.filter((item) => !statusFilter.value || item.status === statusFilter.value),
);

const statusMeta = (status: string) => {
  switch (status) {
    case "completed":
      return { label: "已完成", color: "success" as const };
    case "exception":
      return { label: "异常", color: "warning" as const };
    case "picking":
      return { label: "拣货中", color: "info" as const };
    case "packed":
      return { label: "已打包", color: "primary" as const };
    default:
      return { label: "待处理", color: "neutral" as const };
  }
};

const loadAll = async () => {
  const [taskRows, exRows] = await Promise.all([
    fulfillmentApi.listTasks(),
    fulfillmentApi.listExceptions(),
  ]);
  tasks.value = taskRows.map((item) => ({
    id: item.id,
    orderId: item.orderId,
    warehouseId: item.warehouseId,
    status: item.status,
    assignedTo: item.assignedTo || "-",
    updatedAt: item.updatedAt ? item.updatedAt.replace("T", " ").slice(0, 16) : "-",
  }));
  exceptions.value = exRows.map((item) => ({
    ...item,
    createdAt: item.createdAt ? item.createdAt.replace("T", " ").slice(0, 16) : "-",
  }));
};

const createTask = async () => {
  if (!newTask.orderId || !newTask.warehouseId) return;
  await fulfillmentApi.createTask({
    order_id: newTask.orderId,
    warehouse_id: newTask.warehouseId,
  });
  newTask.orderId = "";
  newTask.warehouseId = "";
  await loadAll();
};

const completeTask = async (id: string) => {
  await fulfillmentApi.completeTask(id, { operator_id: "admin" });
  await loadAll();
};

const prepareException = async (task: TaskRow) => {
  await fulfillmentApi.reportException({
    task_id: task.id,
    type: "stockout",
    reason: "库存不足，待补货",
  });
  await loadAll();
};

onMounted(() => {
  loadAll();
});
</script>
