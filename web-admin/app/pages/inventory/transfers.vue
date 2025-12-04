<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">库存调拨</h1>
        <p class="text-gray-500 dark:text-gray-400">
          协调跨仓调拨与调拨在途监控，避免区域缺货。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton variant="ghost" color="neutral" icon="i-heroicons-document-duplicate">
          调拨模板
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">发起调拨</UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-4">
      <UCard v-for="card in transferStats" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较上周
        </p>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-3">
      <UCard class="lg:col-span-2">
        <template #header>
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">调拨单</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                支持按状态、区域快速筛选。
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <USelect v-model="statusFilter" class="w-40" :options="transferStatusOptions" />
              <USelect v-model="regionFilter" class="w-40" :options="regionOptions" />
              <UInput
                v-model="transferKeyword"
                class="w-48"
                placeholder="调拨单/商品"
                icon="i-heroicons-magnifying-glass"
              />
            </div>
          </div>
        </template>

        <UTable :columns="columns" :data="filteredTransfers">
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
              <UButton size="xs" variant="ghost" color="primary">跟踪物流</UButton>
            </div>
          </template>
        </UTable>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">在途监控</h3>
            <UBadge color="info" variant="subtle">{{ tracking.length }} 条</UBadge>
          </div>
        </template>

        <ul class="space-y-4">
          <li
            v-for="item in tracking"
            :key="item.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-start justify-between">
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">
                  {{ item.product }}
                </p>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ item.from }} → {{ item.to }}
                </p>
              </div>
              <UBadge :color="statusMeta(item.status).color" variant="subtle">
                {{ statusMeta(item.status).label }}
              </UBadge>
            </div>
            <div class="mt-3 space-y-2 text-xs text-gray-500 dark:text-gray-400">
              <div class="flex justify-between">
                <span>数量</span>
                <span>{{ item.quantity }}</span>
              </div>
              <div class="flex justify-between">
                <span>预计到达</span>
                <span>{{ item.eta }}</span>
              </div>
            </div>
            <div class="mt-3">
              <UProgress :value="item.progress" size="xs" />
            </div>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">异常上报</UButton>
              <UButton size="xs" variant="ghost">回传 ERP</UButton>
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
  name: "inventory-transfers",
});

type TransferStatus = "draft" | "shipping" | "arrived" | "in-review";

type TransferOrder = {
  id: string;
  product: string;
  sku: string;
  from: string;
  to: string;
  quantity: number;
  region: string;
  status: TransferStatus;
  progress: number;
  eta: string;
};

const transfers = ref<TransferOrder[]>([
  {
    id: "TR-202402-01",
    product: "MacBook Air M3",
    sku: "MBA-M3-512-SLV",
    from: "上海青浦保税仓",
    to: "北京顺义中心仓",
    quantity: 80,
    region: "华东",
    status: "shipping",
    progress: 64,
    eta: "2 天",
  },
  {
    id: "TR-202402-02",
    product: "Nintendo Switch OLED",
    sku: "NS-OLED-WHT",
    from: "广州黄埔前置仓",
    to: "成都分拨中心",
    quantity: 45,
    region: "华南",
    status: "arrived",
    progress: 100,
    eta: "已签收",
  },
  {
    id: "TR-202402-03",
    product: "iPhone 15 Pro",
    sku: "IP15P-256-BLK",
    from: "北京顺义中心仓",
    to: "深圳宝安前置仓",
    quantity: 120,
    region: "华北",
    status: "in-review",
    progress: 88,
    eta: "1 天",
  },
  {
    id: "TR-202402-04",
    product: "DJI Air 3",
    sku: "DJI-A3-FLY",
    from: "深圳宝安前置仓",
    to: "上海青浦保税仓",
    quantity: 30,
    region: "华南",
    status: "draft",
    progress: 15,
    eta: "-",
  },
]);

const tracking = ref([
  {
    id: "TT-1",
    product: "PlayStation 5 Slim",
    from: "上海青浦保税仓",
    to: "北京顺义中心仓",
    quantity: 60,
    progress: 72,
    status: "shipping" as TransferStatus,
    eta: "2 月 13 日",
  },
  {
    id: "TT-2",
    product: "Surface Laptop Studio 2",
    from: "深圳宝安前置仓",
    to: "成都分拨中心",
    quantity: 25,
    progress: 38,
    status: "shipping" as TransferStatus,
    eta: "2 月 15 日",
  },
]);

const transferKeyword = ref("");
const statusFilter = ref<TransferStatus | "">("");
const regionFilter = ref("");

const transferStatusOptions = [
  { label: "全部状态", value: "" },
  { label: "草稿", value: "draft" },
  { label: "在途", value: "shipping" },
  { label: "已到达", value: "arrived" },
  { label: "入库核对", value: "in-review" },
];

const regionOptions = computed(() =>
  [{ label: "全部区域", value: "" }].concat(
    Array.from(new Set(transfers.value.map((transfer) => transfer.region))).map(
      (region) => ({ label: region, value: region }),
    ),
  ),
);

const columns = computed<TableColumn<TransferOrder>[]>(() => [
  { accessorKey: "id", header: "调拨单号" },
  { accessorKey: "product", header: "商品" },
  { accessorKey: "sku", header: "SKU" },
  {
    accessorKey: "from",
    header: "调出仓",
    cell: ({ row }) => `${row.original.from} → ${row.original.to}`,
  },
  { accessorKey: "quantity", header: "数量" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "progress", header: "进度" },
  { accessorKey: "eta", header: "预计到达" },
  { id: "actions", header: "操作" },
]);

const statusMeta = (status: TransferStatus | "") => {
  switch (status) {
    case "draft":
      return { label: "草稿", color: "neutral" as const };
    case "shipping":
      return { label: "在途", color: "info" as const };
    case "arrived":
      return { label: "已到达", color: "success" as const };
    case "in-review":
      return { label: "入库核对", color: "warning" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const filteredTransfers = computed(() =>
  transfers.value.filter((transfer) => {
    const matchesKeyword =
      !transferKeyword.value ||
      transfer.id.includes(transferKeyword.value) ||
      transfer.product.toLowerCase().includes(transferKeyword.value.toLowerCase());
    const matchesStatus =
      !statusFilter.value || transfer.status === statusFilter.value;
    const matchesRegion =
      !regionFilter.value || transfer.region === regionFilter.value;
    return matchesKeyword && matchesStatus && matchesRegion;
  }),
);

const transferStats = computed(() => [
  { title: "本周发起调拨", value: transfers.value.length, trend: 6.5 },
  {
    title: "在途调拨",
    value: transfers.value.filter((item) => item.status === "shipping").length,
    trend: 2.1,
  },
  {
    title: "平均到货时长",
    value: "2.6 天",
    trend: -0.4,
  },
  {
    title: "调拨完成率",
    value:
      Math.round(
        (transfers.value.filter((item) => item.status === "arrived").length /
          transfers.value.length) *
          100,
      ) + "%",
    trend: 1.7,
  },
]);
</script>
