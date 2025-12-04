<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.suppliers") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理供应商信息，维护供应链合作关系
        </p>
      </div>
      <div class="flex space-x-2">
        <UButton color="primary" variant="outline" icon="i-heroicons-plus">
          添加供应商
        </UButton>
        <UButton
          color="primary"
          variant="soft"
          icon="i-heroicons-arrow-down-tray"
        >
          导出数据
        </UButton>
      </div>
    </div>

    <!-- 概览（原样保留） -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">总供应商</p>
            <p class="text-2xl font-bold text-primary-600">156</p>
            <p class="text-xs text-success-500">+8 本月新增</p>
          </div>
          <UIcon
            name="i-heroicons-building-office"
            class="w-8 h-8 text-primary-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">活跃供应商</p>
            <p class="text-2xl font-bold text-success-600">142</p>
            <p class="text-xs text-success-500">91.0% 活跃率</p>
          </div>
          <UIcon
            name="i-heroicons-check-circle"
            class="w-8 h-8 text-success-500"
          />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">平均评分</p>
            <p class="text-2xl font-bold text-warning-600">4.2</p>
            <p class="text-xs text-success-500">+0.1 较上月</p>
          </div>
          <UIcon name="i-heroicons-star" class="w-8 h-8 text-warning-500" />
        </div>
      </UCard>

      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">合作金额</p>
            <p class="text-2xl font-bold text-error-600">¥2.8M</p>
            <p class="text-xs text-success-500">+15.2% 较上月</p>
          </div>
          <UIcon
            name="i-heroicons-currency-yen"
            class="w-8 h-8 text-error-500"
          />
        </div>
      </UCard>
    </div>

    <!-- 筛选和搜索 -->
    <UCard class="mb-6">
      <div class="flex flex-wrap gap-4 items-center">
        <UInput
          v-model="searchQuery"
          placeholder="搜索供应商名称、联系人..."
          icon="i-heroicons-magnifying-glass"
          class="flex-1 min-w-64"
        />
        <USelect
          v-model="selectedStatus"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="状态筛选"
        />
        <USelect
          v-model="selectedCategory"
          :options="categoryOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="分类筛选"
        />
        <UButton variant="outline" icon="i-heroicons-funnel">
          高级筛选
        </UButton>
      </div>
    </UCard>

    <!-- 列表 -->
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold">供应商列表</h3>
      </template>

      <UTable :data="filteredSuppliers" :columns="columns" class="w-full">
        <!-- v3: 用 -cell 插槽 -->
        <template #status-cell="{ getValue }">
          <UBadge
            :color="
              getValue() === '活跃'
                ? 'success'
                : getValue() === '暂停'
                  ? 'warning'
                  : 'neutral'
            "
            variant="subtle"
          >
            {{ getValue() }}
          </UBadge>
        </template>

        <template #rating-cell="{ getValue }">
          <div class="flex items-center gap-1">
            <UIcon
              v-for="i in 5"
              :key="i"
              name="i-heroicons-star"
              :class="
                i <= Math.round(+getValue())
                  ? 'text-warning-400'
                  : 'text-gray-300'
              "
              class="w-4 h-4"
            />
            <span class="text-sm text-gray-500 ml-1"
              >({{ (+getValue()).toFixed(1) }})</span
            >
          </div>
        </template>

        <template #cooperation-cell="{ getValue }">
          <span class="font-medium">
            {{
              new Intl.NumberFormat("zh-CN", {
                style: "currency",
                currency: "CNY",
                maximumFractionDigits: 0,
              }).format(Number(getValue() || 0))
            }}
          </span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="outline" icon="i-heroicons-eye">
              查看
            </UButton>
            <UButton size="xs" color="primary" icon="i-heroicons-pencil-square">
              编辑
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
const selectedCategory = ref<string | "">("");

// 筛选项
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "活跃", value: "活跃" },
  { label: "暂停", value: "暂停" },
  { label: "待审核", value: "待审核" },
];
const categoryOptions = [
  { label: "全部分类", value: "" },
  { label: "服装配饰", value: "服装配饰" },
  { label: "电子产品", value: "电子产品" },
  { label: "家居用品", value: "家居用品" },
  { label: "食品饮料", value: "食品饮料" },
];

// 类型
type Supplier = {
  id: number;
  name: string;
  contact: string;
  category: string;
  status: "活跃" | "暂停" | "待审核";
  rating: number;
  cooperation: number;
};

// 列定义（v3 TanStack 风格）
const columns = computed<TableColumn<Supplier>[]>(() => [
  { accessorKey: "name", header: t("suppliers.name") || "供应商名称" },
  { accessorKey: "contact", header: t("suppliers.contact") || "联系人" },
  { accessorKey: "category", header: t("suppliers.category") || "主营类目" },
  { accessorKey: "status", header: t("suppliers.status") || "状态" },
  { accessorKey: "rating", header: t("suppliers.rating") || "评分" },
  {
    accessorKey: "cooperation",
    header: t("suppliers.cooperation") || "合作金额",
  },
  { id: "actions", header: t("common.actions") || "操作" },
]);

// 数据
const suppliers = ref<Supplier[]>([
  {
    id: 1,
    name: "优质服装供应商",
    contact: "张经理 (138****8888)",
    category: "服装配饰",
    status: "活跃",
    rating: 4.5,
    cooperation: 280000,
  },
  {
    id: 2,
    name: "科技电子有限公司",
    contact: "李总 (139****9999)",
    category: "电子产品",
    status: "活跃",
    rating: 4.2,
    cooperation: 450000,
  },
  {
    id: 3,
    name: "家居生活用品厂",
    contact: "王主管 (137****7777)",
    category: "家居用品",
    status: "暂停",
    rating: 3.8,
    cooperation: 120000,
  },
  {
    id: 4,
    name: "绿色食品供应链",
    contact: "陈经理 (136****6666)",
    category: "食品饮料",
    status: "活跃",
    rating: 4.7,
    cooperation: 380000,
  },
  {
    id: 5,
    name: "时尚潮流服饰",
    contact: "赵总监 (135****5555)",
    category: "服装配饰",
    status: "活跃",
    rating: 4.1,
    cooperation: 220000,
  },
]);

// 过滤
const filteredSuppliers = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return suppliers.value.filter((s) => {
    const passQ =
      !q ||
      s.name.toLowerCase().includes(q) ||
      s.contact.toLowerCase().includes(q);
    const passStatus =
      !selectedStatus.value || s.status === selectedStatus.value;
    const passCategory =
      !selectedCategory.value || s.category === selectedCategory.value;
    return passQ && passStatus && passCategory;
  });
});
</script>
