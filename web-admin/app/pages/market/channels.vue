<template>
  <div class="p-6">
    <!-- 标题区 -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.channels") || "渠道商" }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理渠道合作伙伴、返佣与结算
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="primary" variant="outline" icon="i-heroicons-plus">
          新增渠道商
        </UButton>
        <UButton
          color="primary"
          variant="soft"
          icon="i-heroicons-arrow-down-tray"
        >
          导出
        </UButton>
      </div>
    </div>

    <!-- 概览卡片（可选） -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">渠道商总数</p>
            <p class="text-2xl font-bold text-primary-600">
              {{ overview.total }}
            </p>
            <p class="text-xs text-success-500">
              +{{ overview.newThisMonth }} 本月新增
            </p>
          </div>
          <UIcon name="i-heroicons-users" class="w-8 h-8 text-primary-500" />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">活跃率</p>
            <p class="text-2xl font-bold text-success-600">
              {{ (overview.activeRate * 100).toFixed(1) }}%
            </p>
            <p class="text-xs text-success-500">
              上月 {{ (overview.lastActiveRate * 100).toFixed(1) }}%
            </p>
          </div>
          <UIcon
            name="i-heroicons-check-badge"
            class="w-8 h-8 text-success-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">近30天GMV</p>
            <p class="text-2xl font-bold text-warning-600">
              {{
                new Intl.NumberFormat("zh-CN", {
                  style: "currency",
                  currency: "CNY",
                  maximumFractionDigits: 0,
                }).format(overview.gmv30d)
              }}
            </p>
            <p class="text-xs text-success-500">
              环比 +{{ (overview.gmvMoM * 100).toFixed(1) }}%
            </p>
          </div>
          <UIcon
            name="i-heroicons-chart-bar"
            class="w-8 h-8 text-warning-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">平均返佣</p>
            <p
              class="text-2xl font-bold text-neutral-700 dark:text-neutral-200"
            >
              {{ (overview.avgCommission * 100).toFixed(1) }}%
            </p>
            <p class="text-xs text-muted">按有效订单加权</p>
          </div>
          <UIcon
            name="i-heroicons-banknotes"
            class="w-8 h-8 text-neutral-500"
          />
        </div>
      </UCard>
    </div>

    <!-- 筛选区 -->
    <UCard class="mb-6">
      <div class="flex flex-wrap items-center gap-3">
        <UInput
          v-model="search"
          placeholder="搜索渠道商名称、联系人、编号…"
          icon="i-heroicons-magnifying-glass"
          class="min-w-64 flex-1"
        />
        <USelect
          v-model="selStatus"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="状态"
          class="w-36"
        />
        <USelect
          v-model="selLevel"
          :options="levelOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="等级"
          class="w-36"
        />
        <USelect
          v-model="selRegion"
          :options="regionOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="区域"
          class="w-40"
        />
        <UButton variant="outline" icon="i-heroicons-funnel">
          高级筛选
        </UButton>
      </div>
    </UCard>

    <!-- 列表 -->
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold">渠道商列表</h3>
      </template>

      <UTable :data="filteredChannels" :columns="columns" class="w-full">
        <!-- v3: 单元格插槽统一用 -cell -->
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

        <template #commissionRate-cell="{ getValue }">
          <div class="text-right font-medium">
            {{ (Number(getValue() || 0) * 100).toFixed(1) }}%
          </div>
        </template>

        <template #settlement-cycle-cell="{ getValue }">
          <UBadge color="neutral" variant="soft">{{ getValue() }}</UBadge>
        </template>

        <template #gmv30d-cell="{ getValue }">
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

// ---- 概览（演示数据） ----
const overview = reactive({
  total: 86,
  newThisMonth: 5,
  activeRate: 0.87,
  lastActiveRate: 0.83,
  gmv30d: 2680000,
  gmvMoM: 0.152,
  avgCommission: 0.082,
});

// ---- 筛选状态 ----
const search = ref("");
const selStatus = ref<string | "">("");
const selLevel = ref<string | "">("");
const selRegion = ref<string | "">("");

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "活跃", value: "活跃" },
  { label: "暂停", value: "暂停" },
  { label: "待审核", value: "待审核" },
];
const levelOptions = [
  { label: "全部等级", value: "" },
  { label: "A 级", value: "A" },
  { label: "B 级", value: "B" },
  { label: "C 级", value: "C" },
];
const regionOptions = [
  { label: "全部区域", value: "" },
  { label: "华北", value: "华北" },
  { label: "华东", value: "华东" },
  { label: "华南", value: "华南" },
  { label: "西南", value: "西南" },
];

// ---- 类型 & 列定义（v3 TanStack 风格） ----
type Channel = {
  id: string;
  name: string;
  contact: string;
  level: "A" | "B" | "C";
  region: string;
  status: "活跃" | "暂停" | "待审核";
  commissionRate: number; // 0.08 = 8%
  settlement: "月结" | "半月结" | "季结";
  gmv30d: number; // 近30天GMV
};

const columns = computed<TableColumn<Channel>[]>(() => [
  { accessorKey: "name", header: t("channels.name") || "渠道商名称" },
  { accessorKey: "contact", header: t("channels.contact") || "联系人" },
  {
    accessorKey: "level",
    header: t("channels.level") || "等级",
    cell: ({ getValue }) =>
      h("span", { class: "font-medium" }, getValue() as string),
  },
  { accessorKey: "region", header: t("channels.region") || "区域" },
  { accessorKey: "status", header: t("channels.status") || "状态" },
  {
    accessorKey: "commissionRate",
    header: t("channels.commission") || "返佣",
    meta: { class: { td: "text-right" } },
  },
  {
    accessorKey: "settlement",
    header: t("channels.settlement") || "结算周期",
    id: "settlement-cycle",
  },
  {
    accessorKey: "gmv30d",
    header: "近30天GMV",
    meta: { class: { td: "text-right" } },
  },
  { id: "actions", header: t("common.actions") || "操作" },
]);

// ---- 示例数据 ----
const channels = ref<Channel[]>([
  {
    id: "CH001",
    name: "华北核心渠道",
    contact: "孙经理 (138****0001)",
    level: "A",
    region: "华北",
    status: "活跃",
    commissionRate: 0.1,
    settlement: "月结",
    gmv30d: 680000,
  },
  {
    id: "CH002",
    name: "华东优选代理",
    contact: "周总 (139****0002)",
    level: "A",
    region: "华东",
    status: "活跃",
    commissionRate: 0.08,
    settlement: "半月结",
    gmv30d: 520000,
  },
  {
    id: "CH003",
    name: "华南分销网络",
    contact: "刘主管 (137****0003)",
    level: "B",
    region: "华南",
    status: "暂停",
    commissionRate: 0.06,
    settlement: "月结",
    gmv30d: 120000,
  },
  {
    id: "CH004",
    name: "西南成长渠道",
    contact: "张经理 (136****0004)",
    level: "C",
    region: "西南",
    status: "待审核",
    commissionRate: 0.05,
    settlement: "季结",
    gmv30d: 80000,
  },
]);

// ---- 过滤逻辑 ----
const filteredChannels = computed(() => {
  const q = search.value.trim().toLowerCase();
  return channels.value.filter((c) => {
    const passQ =
      !q ||
      c.name.toLowerCase().includes(q) ||
      c.contact.toLowerCase().includes(q) ||
      c.id.toLowerCase().includes(q);
    const passStatus = !selStatus.value || c.status === selStatus.value;
    const passLevel = !selLevel.value || c.level === selLevel.value;
    const passRegion = !selRegion.value || c.region === selRegion.value;
    return passQ && passStatus && passLevel && passRegion;
  });
});
</script>
