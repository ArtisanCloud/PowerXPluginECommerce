<template>
  <div class="p-6">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
        {{ $t("nav.marketing") }}
      </h1>
      <UButton color="primary" icon="i-heroicons-plus"> 创建活动 </UButton>
    </div>

    <!-- 指标卡，保持不变 -->

    <UCard class="mt-6">
      <template #header>
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-semibold">营销活动列表</h3>
          <div class="flex space-x-2">
            <UInput
              v-model="searchQuery"
              :placeholder="$t('common.search')"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="selectedStatus"
              :options="campaignStatuses"
              option-attribute="label"
              value-attribute="value"
              :placeholder="$t('common.filter')"
            />
          </div>
        </div>
      </template>

      <UTable :data="filteredCampaigns" :columns="columns" :loading="loading">
        <!-- v3: 用 -cell 插槽 -->
        <template #type-cell="{ getValue }">
          <UBadge :color="typeColor(String(getValue()))" variant="soft">
            {{ getValue() }}
          </UBadge>
        </template>

        <template #status-cell="{ getValue }">
          <UBadge :color="statusColor(String(getValue()))" variant="soft">
            {{ getValue() }}
          </UBadge>
        </template>

        <template #budget-cell="{ getValue }">
          <div class="text-right font-medium">
            {{
              new Intl.NumberFormat("zh-CN", {
                style: "currency",
                currency: "CNY",
                maximumFractionDigits: 0,
              }).format(Number(getValue() || 0))
            }}
          </div>
        </template>

        <template #conversion-cell="{ getValue }">
          <div class="text-right">
            {{ (Number(getValue() || 0) * 100).toFixed(1) }}%
          </div>
        </template>

        <template #startDate-cell="{ getValue }">
          {{ new Date(String(getValue())).toLocaleDateString("zh-CN") }}
        </template>
        <template #endDate-cell="{ getValue }">
          {{ new Date(String(getValue())).toLocaleDateString("zh-CN") }}
        </template>

        <template #actions-cell>
          <div class="flex space-x-2">
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              icon="i-heroicons-eye"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              size="xs"
              color="primary"
              variant="ghost"
              icon="i-heroicons-pencil-square"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton size="xs" color="error" variant="soft"> 停止 </UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
const { t } = useI18n();

// 状态
const searchQuery = ref("");
const selectedStatus = ref<string | "">("");
const loading = ref(false);

type Campaign = {
  campaignId: string;
  name: string;
  type: "促销活动" | "拉新活动" | "会员活动" | "节日活动";
  startDate: string; // 可被 Date 解析
  endDate: string;
  budget: number; // 用 number 存储
  conversion: number; // 0.152 = 15.2%
  status: "进行中" | "待开始" | "已结束" | "已暂停";
};

const campaignStatuses = [
  { label: "全部状态", value: "" },
  { label: "进行中", value: "进行中" },
  { label: "待开始", value: "待开始" },
  { label: "已结束", value: "已结束" },
  { label: "已暂停", value: "已暂停" },
];

// 列定义（v3 TanStack）
const columns = computed<TableColumn<Campaign>[]>(() => [
  { accessorKey: "campaignId", header: "活动ID" },
  { accessorKey: "name", header: "活动名称" },
  { accessorKey: "type", header: "活动类型" },
  { accessorKey: "startDate", header: "开始时间" },
  { accessorKey: "endDate", header: "结束时间" },
  {
    accessorKey: "budget",
    header: "预算",
    meta: { class: { td: "text-right" } },
  },
  {
    accessorKey: "conversion",
    header: "转化率",
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

// 示例数据（budget/convert 改为 number）
const campaigns = ref<Campaign[]>([
  {
    campaignId: "C001",
    name: "双11大促销",
    type: "促销活动",
    startDate: "2024-11-01",
    endDate: "2024-11-11",
    budget: 50000,
    conversion: 0.152,
    status: "进行中",
  },
  {
    campaignId: "C002",
    name: "新用户注册礼",
    type: "拉新活动",
    startDate: "2024-01-01",
    endDate: "2024-12-31",
    budget: 30000,
    conversion: 0.087,
    status: "进行中",
  },
  {
    campaignId: "C003",
    name: "会员专享日",
    type: "会员活动",
    startDate: "2024-02-01",
    endDate: "2024-02-03",
    budget: 20000,
    conversion: 0.221,
    status: "已结束",
  },
  {
    campaignId: "C004",
    name: "春节特惠",
    type: "节日活动",
    startDate: "2024-02-10",
    endDate: "2024-02-17",
    budget: 80000,
    conversion: 0.0,
    status: "待开始",
  },
]);

// 过滤
const filteredCampaigns = computed(() => {
  const q = searchQuery.value.trim().toLowerCase();
  return campaigns.value.filter((c) => {
    const passQ = !q || c.name.toLowerCase().includes(q);
    const passS = !selectedStatus.value || c.status === selectedStatus.value;
    return passQ && passS;
  });
});

// 颜色映射（语义色）
const typeColor = (type: string) => {
  const map: Record<
    string,
    "error" | "primary" | "purple" | "success" | "neutral"
  > = {
    促销活动: "error",
    拉新活动: "primary",
    会员活动: "purple",
    节日活动: "success",
  };
  return map[type] || "neutral";
};
const statusColor = (status: string) => {
  const map: Record<
    string,
    "success" | "info" | "neutral" | "warning" | "error"
  > = {
    进行中: "success",
    待开始: "info",
    已结束: "neutral",
    已暂停: "warning",
  };
  return map[status] || "neutral";
};
</script>
