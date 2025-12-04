<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">仓库</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理全国仓库的容量、运营状态与负责人。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-down-tray">
          导出
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          新建仓库
        </UButton>
      </div>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <UCard v-for="card in summaryCards" :key="card.label">
        <template #header>
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ card.label }}</div>
          <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ card.value }}
          </div>
          <div class="mt-1 text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
            {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% vs 上周
          </div>
        </template>
        <UProgress v-if="card.progress !== undefined" :value="card.progress" size="sm" />
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">仓库列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              支持按区域与状态快速筛选。
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="searchQuery"
              class="w-52"
              placeholder="搜索仓库/负责人"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="regionFilter"
              class="w-40"
              :options="regionOptions"
              placeholder="全部区域"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :options="statusOptions"
              placeholder="运营状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredWarehouses">
        <template #status-cell="{ getValue }">
          <UBadge :color="getStatusMeta(getValue()).color" variant="subtle">
            {{ getStatusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #utilization-cell="{ getValue }">
          <div>
            <div class="flex items-center justify-between text-xs text-gray-500">
              <span>{{ getValue() }}%</span>
              <span>100%</span>
            </div>
            <UProgress :value="getValue()" size="xs" />
          </div>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton color="primary" size="xs" variant="ghost">编辑</UButton>
            <UButton color="neutral" size="xs" variant="ghost">盘点计划</UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "inventory-warehouses",
});

type WarehouseStatus = "operational" | "maintenance" | "paused";

type Warehouse = {
  id: string;
  name: string;
  region: string;
  city: string;
  capacity: number;
  utilization: number;
  status: WarehouseStatus;
  manager: string;
  contact: string;
  lastAudit: string;
};

const warehouses = ref<Warehouse[]>([
  {
    id: "WH-BJ-01",
    name: "北京顺义中心仓",
    region: "华北",
    city: "北京",
    capacity: 28000,
    utilization: 72,
    status: "operational",
    manager: "张强",
    contact: "13800001001",
    lastAudit: "2024-02-10",
  },
  {
    id: "WH-SH-02",
    name: "上海青浦保税仓",
    region: "华东",
    city: "上海",
    capacity: 35000,
    utilization: 83,
    status: "operational",
    manager: "李琳",
    contact: "13800001002",
    lastAudit: "2024-01-28",
  },
  {
    id: "WH-GZ-03",
    name: "广州黄埔前置仓",
    region: "华南",
    city: "广州",
    capacity: 18000,
    utilization: 65,
    status: "maintenance",
    manager: "陈凯",
    contact: "13800001003",
    lastAudit: "2023-12-18",
  },
  {
    id: "WH-SZ-04",
    name: "深圳宝安前置仓",
    region: "华南",
    city: "深圳",
    capacity: 12000,
    utilization: 92,
    status: "operational",
    manager: "王欣",
    contact: "13800001004",
    lastAudit: "2024-01-05",
  },
  {
    id: "WH-CD-05",
    name: "成都分拨中心",
    region: "西南",
    city: "成都",
    capacity: 15000,
    utilization: 58,
    status: "paused",
    manager: "赵云",
    contact: "13800001005",
    lastAudit: "2023-11-22",
  },
]);

const searchQuery = ref("");
const regionFilter = ref("");
const statusFilter = ref<WarehouseStatus | "">("");

const regionOptions = computed(() =>
  [{ label: "全部区域", value: "" }].concat(
    Array.from(new Set(warehouses.value.map((item) => item.region))).map((region) => ({
      label: region,
      value: region,
    })),
  ),
);

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "运营中", value: "operational" },
  { label: "维护中", value: "maintenance" },
  { label: "暂停", value: "paused" },
];

const getStatusMeta = (status: WarehouseStatus | "") => {
  switch (status) {
    case "operational":
      return { label: "运营中", color: "success" as const };
    case "maintenance":
      return { label: "维护中", color: "warning" as const };
    case "paused":
      return { label: "暂停", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const filteredWarehouses = computed(() => {
  return warehouses.value.filter((item) => {
    const matchesSearch =
      !searchQuery.value ||
      item.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      item.manager.includes(searchQuery.value);
    const matchesRegion =
      !regionFilter.value || item.region === regionFilter.value;
    const matchesStatus =
      !statusFilter.value || item.status === statusFilter.value;
    return matchesSearch && matchesRegion && matchesStatus;
  });
});

const columns = computed<TableColumn<Warehouse>[]>(() => [
  { accessorKey: "name", header: "仓库" },
  { accessorKey: "region", header: "区域" },
  { accessorKey: "city", header: "城市" },
  {
    accessorKey: "capacity",
    header: "容量 (m²)",
    cell: ({ getValue }) => getValue()?.toLocaleString(),
  },
  { accessorKey: "utilization", header: "利用率" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "manager", header: "负责人" },
  { accessorKey: "lastAudit", header: "最近盘点" },
  { id: "actions", header: "操作" },
]);

const summaryCards = computed(() => [
  {
    label: "仓库总数",
    value: warehouses.value.length,
    trend: 2.4,
  },
  {
    label: "总可用面积",
    value:
      warehouses.value
        .reduce((total, current) => total + current.capacity, 0)
        .toLocaleString() + " m²",
    trend: 0.8,
    progress: 72,
  },
  {
    label: "平均利用率",
    value:
      Math.round(
        warehouses.value.reduce((total, item) => total + item.utilization, 0) /
          warehouses.value.length,
      ) + "%",
    trend: -1.2,
    progress: Math.round(
      warehouses.value.reduce((total, item) => total + item.utilization, 0) /
        warehouses.value.length,
    ),
  },
  {
    label: "维护/暂停仓",
    value: warehouses.value.filter((item) => item.status !== "operational").length,
    trend: 1.1,
  },
]);
</script>
