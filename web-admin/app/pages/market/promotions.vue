<template>
  <div class="p-6">
    <!-- 标题区 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.promotions") }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          创建和管理促销活动，提升销售转化
        </p>
      </div>
      <div class="flex space-x-2">
        <UButton color="primary" variant="outline" icon="i-heroicons-plus"
          >创建促销</UButton
        >
        <UButton
          color="primary"
          variant="soft"
          icon="i-heroicons-chart-bar-square"
          >活动报告</UButton
        >
      </div>
    </div>

    <!-- 概览（原样保留） -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">进行中活动</p>
            <p class="text-2xl font-bold text-primary-600">12</p>
            <p class="text-xs text-success-500">+3 本周新增</p>
          </div>
          <UIcon
            name="i-heroicons-megaphone"
            class="w-8 h-8 text-primary-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">总参与用户</p>
            <p class="text-2xl font-bold text-success-600">8,420</p>
            <p class="text-xs text-success-500">+25.3% 较上周</p>
          </div>
          <UIcon name="i-heroicons-users" class="w-8 h-8 text-success-500" />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">促销收入</p>
            <p class="text-2xl font-bold text-warning-600">¥156K</p>
            <p class="text-xs text-success-500">+18.7% 较上周</p>
          </div>
          <UIcon
            name="i-heroicons-currency-yen"
            class="w-8 h-8 text-warning-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">平均转化率</p>
            <p class="text-2xl font-bold text-error-600">12.8%</p>
            <p class="text-xs text-success-500">+2.1% 较上周</p>
          </div>
          <UIcon name="i-heroicons-chart-bar" class="w-8 h-8 text-error-500" />
        </div>
      </UCard>
    </div>

    <!-- 筛选 -->
    <UCard class="mb-6">
      <div class="flex flex-wrap gap-4 items-center">
        <UInput
          v-model="searchQuery"
          placeholder="搜索活动名称..."
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
          v-model="selectedType"
          :options="typeOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="类型筛选"
        />
        <UButton variant="outline" icon="i-heroicons-funnel">高级筛选</UButton>
      </div>
    </UCard>

    <!-- 列表 -->
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold">促销活动列表</h3>
      </template>

      <UTable :data="filteredPromotions" :columns="columns" class="w-full">
        <!-- v3: 单元格插槽统一用 -cell -->
        <template #status-cell="{ getValue }">
          <UBadge :color="getStatusColor(String(getValue()))" variant="soft">
            {{ getValue() }}
          </UBadge>
        </template>

        <template #type-cell="{ getValue }">
          <UBadge
            :color="getTypeColor(String(getValue()))"
            variant="outline"
            size="sm"
          >
            {{ getValue() }}
          </UBadge>
        </template>

        <template #discount-cell="{ getValue }">
          <span class="font-medium text-success-600">{{ getValue() }}</span>
        </template>

        <template #participants-cell="{ getValue }">
          <span class="font-medium">{{
            Number(getValue() || 0).toLocaleString()
          }}</span>
        </template>

        <template #revenue-cell="{ getValue }">
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
          <div class="flex space-x-2">
            <UButton size="xs" variant="outline">查看</UButton>
            <UButton
              v-if="row.original.status === '进行中'"
              size="xs"
              color="warning"
              variant="soft"
              >暂停</UButton
            >
            <UButton
              v-else-if="row.original.status === '未开始'"
              size="xs"
              color="primary"
              >编辑</UButton
            >
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- 分析区（占位） -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
      <UCard>
        <template #header
          ><h3 class="text-lg font-semibold">活动效果趋势</h3></template
        >
        <div
          class="h-64 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded"
        >
          <p class="text-gray-500">活动效果趋势图表区域</p>
        </div>
      </UCard>

      <UCard>
        <template #header
          ><h3 class="text-lg font-semibold">活动类型分布</h3></template
        >
        <div class="space-y-4">
          <div
            v-for="type in promotionTypes"
            :key="type.name"
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded"
          >
            <div class="flex items-center space-x-3">
              <div
                class="w-4 h-4 rounded"
                :style="{ backgroundColor: type.color }"
              ></div>
              <span class="font-medium">{{ type.name }}</span>
            </div>
            <div class="text-right">
              <p class="font-semibold">{{ type.count }}</p>
              <p class="text-xs text-gray-500">个活动</p>
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
const { t } = useI18n();

// 响应式数据
const searchQuery = ref("");
const selectedStatus = ref<string | "">("");
const selectedType = ref<string | "">("");

// 选项
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "未开始", value: "未开始" },
  { label: "进行中", value: "进行中" },
  { label: "已结束", value: "已结束" },
  { label: "已暂停", value: "已暂停" },
];
const typeOptions = [
  { label: "全部类型", value: "" },
  { label: "满减优惠", value: "满减优惠" },
  { label: "折扣促销", value: "折扣促销" },
  { label: "买赠活动", value: "买赠活动" },
  { label: "限时秒杀", value: "限时秒杀" },
];

// 类型 & 列定义（v3 TanStack 风格）
type Promotion = {
  id: number;
  name: string;
  type: "满减优惠" | "折扣促销" | "买赠活动" | "限时秒杀";
  status: "未开始" | "进行中" | "已结束" | "已暂停";
  discount: string;
  participants: number;
  revenue: number;
};

const columns = computed<TableColumn<Promotion>[]>(() => [
  { accessorKey: "name", header: t("promotions.name") || "活动名称" },
  { accessorKey: "type", header: t("promotions.type") || "类型" },
  { accessorKey: "status", header: t("promotions.status") || "状态" },
  { accessorKey: "discount", header: t("promotions.discount") || "优惠力度" },
  {
    accessorKey: "participants",
    header: t("promotions.participants") || "参与人数",
    meta: { class: { td: "text-right" } },
  },
  {
    accessorKey: "revenue",
    header: t("promotions.revenue") || "促销收入",
    meta: { class: { td: "text-right" } },
  },
  { id: "actions", header: t("common.actions") || "操作" },
]);

// 数据
const promotions = ref<Promotion[]>([
  {
    id: 1,
    name: "双11狂欢节",
    type: "折扣促销",
    status: "进行中",
    discount: "8折起",
    participants: 2580,
    revenue: 45600,
  },
  {
    id: 2,
    name: "新用户专享",
    type: "满减优惠",
    status: "进行中",
    discount: "满200减50",
    participants: 1200,
    revenue: 18900,
  },
  {
    id: 3,
    name: "限时秒杀",
    type: "限时秒杀",
    status: "进行中",
    discount: "5折",
    participants: 890,
    revenue: 12300,
  },
  {
    id: 4,
    name: "买二送一",
    type: "买赠活动",
    status: "未开始",
    discount: "买2送1",
    participants: 0,
    revenue: 0,
  },
  {
    id: 5,
    name: "会员专享日",
    type: "折扣促销",
    status: "已结束",
    discount: "7折",
    participants: 3200,
    revenue: 68500,
  },
]);

// 活动类型分布（演示）
const promotionTypes = ref([
  { name: "折扣促销", count: 8, color: "#3B82F6" },
  { name: "满减优惠", count: 6, color: "#10B981" },
  { name: "买赠活动", count: 4, color: "#F59E0B" },
  { name: "限时秒杀", count: 3, color: "#EF4444" },
]);

// 颜色映射（语义色）
const getStatusColor = (status: string) => {
  const map: Record<string, "neutral" | "success" | "warning" | "error"> = {
    未开始: "neutral",
    进行中: "success",
    已结束: "warning",
    已暂停: "error",
  };
  return map[status] || "neutral";
};
const getTypeColor = (type: string) => {
  const map: Record<
    string,
    "primary" | "success" | "warning" | "error" | "neutral"
  > = {
    折扣促销: "primary",
    满减优惠: "success",
    买赠活动: "warning",
    限时秒杀: "error",
  };
  return map[type] || "neutral";
};

// 过滤
const filteredPromotions = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return promotions.value.filter((p) => {
    const passQ = !q || p.name.toLowerCase().includes(q);
    const passStatus =
      !selectedStatus.value || p.status === selectedStatus.value;
    const passType = !selectedType.value || p.type === selectedType.value;
    return passQ && passStatus && passType;
  });
});
</script>
