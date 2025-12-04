<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">补货与安全库存</h1>
        <p class="text-gray-500 dark:text-gray-400">
          监控安全库存阈值，生成补货计划并协调供应链。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton variant="ghost" color="neutral" icon="i-heroicons-arrow-path">
          同步 ERP
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">创建补货计划</UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in planStats" :key="card.title">
        <div class="space-y-1">
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
          <div class="text-3xl font-semibold text-gray-900 dark:text-white">
            {{ card.value }}
          </div>
          <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
            {{ card.trend >= 0 ? "+" : "" }}{{ card.trend }}% 较昨日
          </p>
        </div>
      </UCard>
    </div>

    <div class="grid gap-6 lg:grid-cols-3">
      <UCard class="lg:col-span-2">
        <template #header>
          <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">补货优先级</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                基于安全库存与最近 7 天销量计算。
              </p>
            </div>
            <div class="flex flex-wrap gap-2">
              <USelect v-model="priorityFilter" class="w-40" :options="priorityOptions" />
              <USelect v-model="warehouseFilter" class="w-40" :options="warehouseOptions" />
              <UInput
                v-model="keyword"
                class="w-48"
                placeholder="搜索商品/SKU"
                icon="i-heroicons-magnifying-glass"
              />
            </div>
          </div>
        </template>

        <UTable :columns="columns" :data="filteredPlans">
          <template #priority-cell="{ getValue }">
            <UBadge :color="priorityMeta(getValue()).color" variant="subtle">
              {{ priorityMeta(getValue()).label }}
            </UBadge>
          </template>
          <template #status-cell="{ getValue }">
            <UBadge :color="getValue() === '已确认' ? 'success' : 'warning'" variant="subtle">
              {{ getValue() }}
            </UBadge>
          </template>
          <template #actions-cell>
            <div class="flex gap-2">
              <UButton size="xs" variant="ghost">调整数量</UButton>
              <UButton size="xs" variant="ghost" color="primary">发起采购</UButton>
            </div>
          </template>
        </UTable>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">补货建议</h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                已根据预测销量生成自动建议。
              </p>
            </div>
            <UBadge color="info" variant="subtle">
              {{ suggestions.length }} 条
            </UBadge>
          </div>
        </template>

        <ul class="space-y-4">
          <li
            v-for="item in suggestions"
            :key="item.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-start justify-between">
              <div>
                <p class="font-semibold text-gray-900 dark:text-white">{{ item.product }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ item.sku }}</p>
              </div>
              <UBadge :color="priorityMeta(item.priority).color" variant="subtle">
                {{ priorityMeta(item.priority).label }}
              </UBadge>
            </div>
            <dl class="mt-3 grid grid-cols-2 gap-3 text-xs text-gray-500 dark:text-gray-400">
              <div>
                <dt>当前库存</dt>
                <dd class="text-sm text-gray-900 dark:text-white">
                  {{ item.currentStock }}
                </dd>
              </div>
              <div>
                <dt>建议补货</dt>
                <dd class="text-sm text-gray-900 dark:text-white">
                  {{ item.recommended }}
                </dd>
              </div>
              <div>
                <dt>目标仓</dt>
                <dd class="text-sm text-gray-900 dark:text-white">
                  {{ item.target }}
                </dd>
              </div>
              <div>
                <dt>预计到货</dt>
                <dd class="text-sm text-gray-900 dark:text-white">
                  {{ item.eta }}
                </dd>
              </div>
            </dl>
            <div class="mt-4 flex gap-2">
              <UButton color="primary" size="xs" variant="soft">采纳</UButton>
              <UButton color="neutral" size="xs" variant="ghost">忽略</UButton>
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
  name: "inventory-replenishment",
});

type Priority = "high" | "medium" | "low";

type RestockPlan = {
  id: string;
  product: string;
  sku: string;
  warehouse: string;
  currentStock: number;
  minStock: number;
  recommended: number;
  priority: Priority;
  eta: string;
  status: "待确认" | "已确认";
};

const plans = ref<RestockPlan[]>([
  {
    id: "RP-1001",
    product: "Apple Watch S9",
    sku: "AW-S9-45-BLK",
    warehouse: "上海青浦保税仓",
    currentStock: 35,
    minStock: 80,
    recommended: 120,
    priority: "high",
    eta: "3 天",
    status: "待确认",
  },
  {
    id: "RP-1002",
    product: "Dyson V15 Detect",
    sku: "DY-V15-STD",
    warehouse: "广州黄埔前置仓",
    currentStock: 60,
    minStock: 50,
    recommended: 40,
    priority: "medium",
    eta: "5 天",
    status: "已确认",
  },
  {
    id: "RP-1003",
    product: "Nintendo Switch OLED",
    sku: "NS-OLED-WHT",
    warehouse: "北京顺义中心仓",
    currentStock: 22,
    minStock: 75,
    recommended: 100,
    priority: "high",
    eta: "2 天",
    status: "待确认",
  },
  {
    id: "RP-1004",
    product: "Kindle Scribe",
    sku: "KD-SCR-32",
    warehouse: "成都分拨中心",
    currentStock: 80,
    minStock: 60,
    recommended: 20,
    priority: "low",
    eta: "7 天",
    status: "已确认",
  },
]);

const suggestions = ref([
  {
    id: "SG-1",
    product: "PlayStation 5 Slim",
    sku: "PS5-SLIM-DIGI",
    currentStock: 15,
    recommended: 70,
    priority: "high" as Priority,
    target: "上海青浦保税仓",
    eta: "4 天",
  },
  {
    id: "SG-2",
    product: "DJI Air 3",
    sku: "DJI-A3-FLY",
    currentStock: 24,
    recommended: 40,
    priority: "medium" as Priority,
    target: "深圳宝安前置仓",
    eta: "5 天",
  },
]);

const keyword = ref("");
const priorityFilter = ref<Priority | "">("");
const warehouseFilter = ref("");

const warehouseOptions = computed(() =>
  [{ label: "全部仓库", value: "" }].concat(
    Array.from(new Set(plans.value.map((plan) => plan.warehouse))).map((warehouse) => ({
      label: warehouse,
      value: warehouse,
    })),
  ),
);

const priorityOptions = [
  { label: "全部优先级", value: "" },
  { label: "高", value: "high" },
  { label: "中", value: "medium" },
  { label: "低", value: "low" },
];

const columns = computed<TableColumn<RestockPlan>[]>(() => [
  { accessorKey: "product", header: "商品" },
  { accessorKey: "sku", header: "SKU" },
  { accessorKey: "warehouse", header: "仓库" },
  { accessorKey: "currentStock", header: "当前库存" },
  { accessorKey: "minStock", header: "安全库存" },
  { accessorKey: "recommended", header: "建议补货" },
  { accessorKey: "priority", header: "优先级" },
  { accessorKey: "eta", header: "预计到货" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

const filteredPlans = computed(() =>
  plans.value.filter((plan) => {
    const matchesKeyword =
      !keyword.value ||
      plan.product.toLowerCase().includes(keyword.value.toLowerCase()) ||
      plan.sku.toLowerCase().includes(keyword.value.toLowerCase());
    const matchesPriority =
      !priorityFilter.value || plan.priority === priorityFilter.value;
    const matchesWarehouse =
      !warehouseFilter.value || plan.warehouse === warehouseFilter.value;
    return matchesKeyword && matchesPriority && matchesWarehouse;
  }),
);

const priorityMeta = (priority: Priority | "") => {
  switch (priority) {
    case "high":
      return { label: "高", color: "error" as const };
    case "medium":
      return { label: "中", color: "warning" as const };
    case "low":
      return { label: "低", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const planStats = computed(() => [
  { title: "待确认计划", value: plans.value.filter((p) => p.status === "待确认").length, trend: 4.2 },
  { title: "今日补货总量", value: plans.value.reduce((sum, item) => sum + item.recommended, 0), trend: -1.3 },
  {
    title: "即将缺货 SKU",
    value: plans.value.filter((p) => p.currentStock <= p.minStock).length,
    trend: 2.8,
  },
]);
</script>
