<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">商品报表</h1>
        <p class="text-gray-500 dark:text-gray-400">
          分析商品销售、库存与毛利表现，支持 Top 商品与类目洞察。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <USelect v-model="categoryFilter" :options="categoryOptions" class="w-48" />
        <USelect v-model="sortKey" :options="sortOptions" class="w-48" />
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

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Top 商品</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">按销售额与毛利排序。</p>
          </div>
          <UInput
            v-model="keyword"
            class="w-64"
            placeholder="搜索 SKU / 商品名称"
            icon="i-heroicons-magnifying-glass"
          />
        </div>
      </template>

      <UTable :columns="columns" :data="filteredProducts">
        <template #price-cell="{ getValue }">
          ¥{{ getValue().toLocaleString() }}
        </template>
        <template #gmv-cell="{ getValue }">
          ¥{{ getValue().toLocaleString() }}
        </template>
        <template #grossMargin-cell="{ getValue }">
          {{ getValue() }}%
        </template>
        <template #inventoryDays-cell="{ getValue }">
          {{ getValue() }} 天
        </template>
        <template #status-cell="{ getValue }">
          <UBadge :color="getValue() === '充足' ? 'success' : 'warning'" variant="subtle">
            {{ getValue() }}
          </UBadge>
        </template>
      </UTable>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">类目表现</h3>
        </template>
        <ul class="space-y-4">
          <li
            v-for="category in categoryPerformance"
            :key="category.name"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ category.name }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ category.skuCount }} 个 SKU</p>
              </div>
              <UBadge :color="category.trend >= 0 ? 'success' : 'neutral'" variant="subtle">
                {{ category.trend >= 0 ? '+' : '' }}{{ category.trend }}%
              </UBadge>
            </div>
            <div class="mt-3 space-y-2 text-sm text-gray-500 dark:text-gray-400">
              <div class="flex justify-between">
                <span>GMV</span>
                <span class="text-gray-900 dark:text-white">¥{{ category.gmv.toLocaleString() }}</span>
              </div>
              <div class="flex justify-between">
                <span>毛利率</span>
                <span class="text-gray-900 dark:text-white">{{ category.margin }}%</span>
              </div>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">库存健康度</h3>
            <UBadge color="info" variant="subtle">{{ stockAlerts.length }} 预警</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="alert in stockAlerts"
            :key="alert.id"
            class="rounded-xl border border-amber-100 p-4 dark:border-amber-900/30"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ alert.product }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ alert.type }}</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ alert.message }}
            </p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">立即处理</UButton>
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
  name: "reports-products",
});

const summaryCards = computed(() => [
  { title: "在售 SKU", value: "1,248", trend: 2.3 },
  { title: "热销 SKU", value: "186", trend: 5.8 },
  { title: "平均毛利率", value: "32.4%", trend: 0.9 },
  { title: "动销率", value: "78%", trend: -1.1 },
]);

type Product = {
  sku: string;
  name: string;
  category: string;
  gmv: number;
  price: number;
  grossMargin: number;
  inventoryDays: number;
  status: "充足" | "紧张";
};

const products = ref<Product[]>([
  {
    sku: "IP15P-256-BLK",
    name: "iPhone 15 Pro 256G",
    category: "手机",
    gmv: 860000,
    price: 7999,
    grossMargin: 28,
    inventoryDays: 12,
    status: "充足",
  },
  {
    sku: "MBA-M3-16G",
    name: "MacBook Air M3 16G",
    category: "电脑",
    gmv: 420000,
    price: 9999,
    grossMargin: 35,
    inventoryDays: 25,
    status: "紧张",
  },
  {
    sku: "DJI-A3-FLY",
    name: "DJI Air 3 Fly More",
    category: "智能硬件",
    gmv: 210000,
    price: 7699,
    grossMargin: 22,
    inventoryDays: 18,
    status: "充足",
  },
  {
    sku: "ROG16-4070",
    name: "ROG 幻 16 RTX4070",
    category: "电脑",
    gmv: 310000,
    price: 11999,
    grossMargin: 38,
    inventoryDays: 9,
    status: "紧张",
  },
]);

const keyword = ref("");
const categoryFilter = ref("");
const sortKey = ref("gmv");

const categoryOptions = computed(() =>
  [{ label: "全部类目", value: "" }].concat(
    Array.from(new Set(products.value.map((p) => p.category))).map((category) => ({
      label: category,
      value: category,
    })),
  ),
);

const sortOptions = [
  { label: "按 GMV 排序", value: "gmv" },
  { label: "按毛利率排序", value: "grossMargin" },
  { label: "按库存周转天数", value: "inventoryDays" },
];

const columns = computed<TableColumn<Product>[]>(() => [
  { accessorKey: "name", header: "商品" },
  { accessorKey: "sku", header: "SKU" },
  { accessorKey: "category", header: "类目" },
  { accessorKey: "price", header: "价格" },
  { accessorKey: "gmv", header: "GMV" },
  { accessorKey: "grossMargin", header: "毛利率" },
  { accessorKey: "inventoryDays", header: "库存周转" },
  { accessorKey: "status", header: "库存状态" },
]);

const sortedProducts = computed(() =>
  [...products.value]
    .filter((product) => !categoryFilter.value || product.category === categoryFilter.value)
    .sort((a, b) => {
      if (sortKey.value === "inventoryDays") {
        return a.inventoryDays - b.inventoryDays;
      }
      return (b as any)[sortKey.value] - (a as any)[sortKey.value];
    }),
);

const filteredProducts = computed(() =>
  sortedProducts.value.filter(
    (product) =>
      product.name.toLowerCase().includes(keyword.value.trim().toLowerCase()) ||
      product.sku.toLowerCase().includes(keyword.value.trim().toLowerCase()),
  ),
);

const categoryPerformance = computed(() => [
  { name: "手机", skuCount: 120, gmv: 1800000, margin: 26, trend: 4.1 },
  { name: "电脑", skuCount: 220, gmv: 960000, margin: 34, trend: 1.2 },
  { name: "智能硬件", skuCount: 80, gmv: 540000, margin: 29, trend: -0.8 },
]);

const stockAlerts = ref([
  {
    id: "SA-1",
    product: "MacBook Air M3 16G",
    type: "库存紧张",
    message: "预计 7 天后售罄，请加速补货。",
  },
  {
    id: "SA-2",
    product: "ROG 幻 16 RTX4070",
    type: "库存偏低",
    message: "当前周转仅 9 天，建议调拨深圳仓 80 台。",
  },
]);
</script>
