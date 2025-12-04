<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">税率与税则</h1>
        <p class="text-gray-500 dark:text-gray-400">
          统一维护国内外税率、类目税则以及跨境备案，避免申报风险。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-document-text">
          下载税率表
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">新增税则</UButton>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-3">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较上月
        </p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">税率列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">按类目与地区筛选。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-56"
              placeholder="搜索类目/编码"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="regionFilter"
              class="w-44"
              :options="regionOptions"
              placeholder="地区"
            />
            <USelect
              v-model="typeFilter"
              class="w-44"
              :options="typeOptions"
              placeholder="税种"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredRules">
        <template #rate-cell="{ getValue }">
          <span class="font-medium text-gray-900 dark:text-white">{{ getValue() }}%</span>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">编辑</UButton>
            <UButton size="xs" variant="ghost" color="primary">同步 ERP</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">跨境备案</h3>
            <UBadge color="info" variant="subtle">{{ filings.length }} 条</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="filing in filings"
            :key="filing.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ filing.country }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ filing.code }}</p>
              </div>
              <UBadge :color="filing.status === '有效' ? 'success' : 'warning'" variant="subtle">
                {{ filing.status }}
              </UBadge>
            </div>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              有效期至 {{ filing.expireAt }} · 负责人 {{ filing.owner }}
            </p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">续期</UButton>
              <UButton size="xs" variant="ghost">查看原件</UButton>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">风险提醒</h3>
            <UBadge color="warning" variant="subtle">{{ risks.length }} 条</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="risk in risks"
            :key="risk.id"
            class="rounded-xl border border-amber-100 p-4 dark:border-amber-900/40"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ risk.title }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ risk.time }}</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ risk.desc }}</p>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" variant="soft">查看详情</UButton>
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
  name: "tax",
});

type TaxRule = {
  id: string;
  category: string;
  hsCode: string;
  region: string;
  type: string;
  rate: number;
  updatedAt: string;
};

const rules = ref<TaxRule[]>([
  {
    id: "TAX-001",
    category: "3C 数码",
    hsCode: "84713020",
    region: "中国大陆",
    type: "增值税",
    rate: 13,
    updatedAt: "2024-02-05",
  },
  {
    id: "TAX-002",
    category: "服装",
    hsCode: "62044300",
    region: "欧盟",
    type: "关税",
    rate: 12,
    updatedAt: "2024-01-28",
  },
  {
    id: "TAX-003",
    category: "美妆",
    hsCode: "33049900",
    region: "美国",
    type: "消费税",
    rate: 10,
    updatedAt: "2024-02-02",
  },
  {
    id: "TAX-004",
    category: "食品",
    hsCode: "21069090",
    region: "中国香港",
    type: "进口税",
    rate: 5,
    updatedAt: "2024-01-16",
  },
]);

const keyword = ref("");
const regionFilter = ref("");
const typeFilter = ref("");

const regionOptions = computed(() =>
  [{ label: "全部地区", value: "" }].concat(
    Array.from(new Set(rules.value.map((rule) => rule.region))).map((region) => ({
      label: region,
      value: region,
    })),
  ),
);

const typeOptions = computed(() =>
  [{ label: "全部税种", value: "" }].concat(
    Array.from(new Set(rules.value.map((rule) => rule.type))).map((type) => ({
      label: type,
      value: type,
    })),
  ),
);

const columns = computed<TableColumn<TaxRule>[]>(() => [
  { accessorKey: "category", header: "类目" },
  { accessorKey: "hsCode", header: "HS Code" },
  { accessorKey: "region", header: "地区" },
  { accessorKey: "type", header: "税种" },
  { accessorKey: "rate", header: "税率" },
  { accessorKey: "updatedAt", header: "最近更新" },
  { id: "actions", header: "操作" },
]);

const filteredRules = computed(() =>
  rules.value.filter((rule) => {
    const matchesKeyword =
      !keyword.value ||
      rule.category.includes(keyword.value) ||
      rule.hsCode.includes(keyword.value);
    const matchesRegion = !regionFilter.value || rule.region === regionFilter.value;
    const matchesType = !typeFilter.value || rule.type === typeFilter.value;
    return matchesKeyword && matchesRegion && matchesType;
  }),
);

const summaryCards = computed(() => [
  {
    title: "近 30 天更新税率",
    value: rules.value.length,
    trend: 4.5,
  },
  {
    title: "跨境备案在效",
    value: "12 份",
    trend: 1.8,
  },
  {
    title: "风险提醒",
    value: "3 条",
    trend: -0.9,
  },
]);

const filings = ref([
  {
    id: "F-01",
    country: "日本跨境备案",
    code: "JP-INV-2024",
    owner: "王楠",
    expireAt: "2025-01-31",
    status: "有效",
  },
  {
    id: "F-02",
    country: "欧盟 OSS 备案",
    code: "EU-OSS-8891",
    owner: "陈晨",
    expireAt: "2024-06-30",
    status: "即将到期",
  },
]);

const risks = ref([
  {
    id: "RISK-1",
    title: "德国 VAT 发票缺失",
    desc: "有 8 单订单缺少客户 vat id，将在下月申报前补齐。",
    time: "今天 09:40",
  },
  {
    id: "RISK-2",
    title: "美国消费税调整",
    desc: "加州将于 3 月起提升电子产品税率 1.5%，请更新价格策略。",
    time: "昨天 17:20",
  },
]);
</script>
