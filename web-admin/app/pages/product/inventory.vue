<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ $t("nav.inventory") }}
      </h1>
      <div class="flex space-x-2">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-heroicons-arrow-down-tray"
        >
          {{ $t("common.import") }}
        </UButton>
        <UButton
          color="primary"
          variant="solid"
          icon="i-heroicons-arrow-up-tray"
        >
          {{ $t("common.export") }}
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">库存列表</h3>
          <div class="flex space-x-2">
            <UInput
              v-model="searchQuery"
              :placeholder="$t('common.search')"
              icon="i-heroicons-magnifying-glass"
              class="w-64"
            />
            <USelect
              v-model="selectedStatus"
              :options="stockStatuses"
              option-attribute="label"
              value-attribute="value"
              :placeholder="$t('common.filter')"
              class="w-36"
            />
          </div>
        </div>
      </template>

      <UTable
        :data="filteredInventory"
        :columns="columns"
        :loading="loading"
        class="w-full"
      >
        <!-- v3: 单元格插槽用 -cell -->
        <template #stock-cell="{ row }">
          <div class="flex items-center gap-2">
            <span>{{ row.original.stock }}</span>
            <UBadge
              :color="getStockColor(row.original.stock, row.original.minStock)"
              variant="subtle"
              size="xs"
            >
              {{ getStockStatus(row.original.stock, row.original.minStock) }}
            </UBadge>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" color="warning" variant="ghost">
              调整库存
            </UButton>
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-eye"
            >
              {{ $t("common.view") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { h } from "vue";
import type { TableColumn } from "@nuxt/ui";
const { t } = useI18n();

// 响应式数据
const searchQuery = ref("");
const selectedStatus = ref<string | "">("");
const loading = ref(false);

// 状态筛选选项
const stockStatuses = [
  { label: "全部状态", value: "" },
  { label: "正常", value: "正常" },
  { label: "库存不足", value: "库存不足" },
  { label: "缺货", value: "缺货" },
];

// 数据类型
type Inventory = {
  productId: string;
  productName: string;
  sku: string;
  stock: number;
  minStock: number;
  warehouse: string;
  lastUpdate: string; // ISO/文本，展示时格式化
};

// 列（v3 TanStack 风格，用 computed 保证切换语言时表头更新）
const columns = computed<TableColumn<Inventory>[]>(() => [
  { accessorKey: "productId", header: t("inventory.productId") || "商品ID" },
  {
    accessorKey: "productName",
    header: t("inventory.productName") || "商品名称",
  },
  { accessorKey: "sku", header: t("inventory.sku") || "SKU" },
  { accessorKey: "stock", header: t("inventory.stock") || "当前库存" },
  { accessorKey: "minStock", header: t("inventory.minStock") || "最低库存" },
  { accessorKey: "warehouse", header: t("inventory.warehouse") || "仓库" },
  {
    accessorKey: "lastUpdate",
    header: t("inventory.lastUpdate") || "最后更新",
    cell: ({ getValue }) => {
      const v = String(getValue() || "");
      // 简单本地化（可改成 dayjs/Intl.DateTimeFormat）
      return h("span", {}, v.replace(" ", " "));
    },
  },
  { id: "actions", header: t("common.actions") || "操作" },
]);

// 模拟数据
const inventory = ref<Inventory[]>([
  {
    productId: "P001",
    productName: "iPhone 15 Pro",
    sku: "IP15P-256-BLK",
    stock: 45,
    minStock: 10,
    warehouse: "北京仓",
    lastUpdate: "2024-01-15 14:30",
  },
  {
    productId: "P002",
    productName: "MacBook Air M3",
    sku: "MBA-M3-512-SLV",
    stock: 8,
    minStock: 5,
    warehouse: "上海仓",
    lastUpdate: "2024-01-14 16:20",
  },
  {
    productId: "P003",
    productName: "AirPods Pro 2",
    sku: "APP2-WHT",
    stock: 2,
    minStock: 15,
    warehouse: "广州仓",
    lastUpdate: "2024-01-13 09:45",
  },
  {
    productId: "P004",
    productName: "iPad Pro 12.9",
    sku: "IPP-129-1TB-GRY",
    stock: 0,
    minStock: 3,
    warehouse: "深圳仓",
    lastUpdate: "2024-01-12 11:15",
  },
]);

// 过滤后的库存
const filteredInventory = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return inventory.value.filter((item) => {
    const passQ =
      !q ||
      item.productName.toLowerCase().includes(q) ||
      item.sku.toLowerCase().includes(q) ||
      item.productId.toLowerCase().includes(q);
    const status = getStockStatus(item.stock, item.minStock);
    const passStatus = !selectedStatus.value || status === selectedStatus.value;
    return passQ && passStatus;
  });
});

// 库存状态 & 语义色
const getStockStatus = (stock: number, minStock: number) => {
  if (stock === 0) return "缺货";
  if (stock <= minStock) return "库存不足";
  return "正常";
};
const getStockColor = (stock: number, minStock: number) => {
  if (stock === 0) return "error"; // 红
  if (stock <= minStock) return "warning"; // 黄
  return "success"; // 绿
};
</script>
