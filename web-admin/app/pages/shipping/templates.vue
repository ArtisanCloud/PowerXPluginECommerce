<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">运费模板</h1>
        <p class="text-gray-500 dark:text-gray-400">
          定义不同区域、重量和履约方式的运费规则。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-inbox-arrow-down">
          导入模板
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          新建模板
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">模板列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">支持按渠道与计费方式筛选。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-60"
              placeholder="搜索模板/渠道"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="channelFilter"
              class="w-44"
              placeholder="全部渠道"
              :options="channelOptions"
            />
            <USelect
              v-model="billingFilter"
              class="w-44"
              placeholder="计费方式"
              :options="billingOptions"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredTemplates">
        <template #status-cell="{ getValue }">
          <UBadge :color="getValue() === 'enabled' ? 'success' : 'neutral'" variant="subtle">
            {{ getValue() === "enabled" ? "启用" : "停用" }}
          </UBadge>
        </template>
        <template #defaultRule-cell="{ row }">
          <span>{{ row.original.defaultRule.region }}</span>
          <span class="ml-2 text-sm text-gray-500 dark:text-gray-400">
            {{ row.original.defaultRule.fee }}
          </span>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">查看</UButton>
            <UButton size="xs" variant="ghost" color="primary">配置</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">热门区域组合</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">基于最近 30 天订单生成。</p>
          </div>
          <UBadge color="info" variant="subtle">{{ heatmap.length }} 组</UBadge>
        </div>
      </template>

      <div class="grid gap-4 md:grid-cols-2">
        <UCard
          v-for="item in heatmap"
          :key="item.id"
          class="border border-gray-100 dark:border-gray-800"
        >
          <template #header>
            <div class="flex items-center justify-between">
              <div class="font-semibold text-gray-900 dark:text-white">{{ item.name }}</div>
              <UBadge :color="item.channel === '京东自营' ? 'primary' : 'neutral'" variant="subtle">
                {{ item.channel }}
              </UBadge>
            </div>
          </template>
          <div class="space-y-2 text-sm text-gray-500 dark:text-gray-400">
            <div class="flex justify-between">
              <span>基础运费</span>
              <span class="text-gray-900 dark:text-white">{{ item.baseFee }}</span>
            </div>
            <div class="flex justify-between">
              <span>续重</span>
              <span class="text-gray-900 dark:text-white">{{ item.extraFee }}</span>
            </div>
            <div class="flex justify-between">
              <span>平均履约时长</span>
              <span class="text-gray-900 dark:text-white">{{ item.leadTime }}</span>
            </div>
          </div>
        </UCard>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "shipping-templates",
});

type BillingType = "weight" | "piece" | "volume";

type Template = {
  id: string;
  name: string;
  channel: string;
  billing: BillingType;
  status: "enabled" | "disabled";
  defaultRule: {
    region: string;
    fee: string;
  };
  lastUpdate: string;
};

const templates = ref<Template[]>([
  {
    id: "TMP-001",
    name: "全国标准模板",
    channel: "自营商城",
    billing: "weight",
    status: "enabled",
    defaultRule: { region: "大陆 1kg 内", fee: "¥12" },
    lastUpdate: "2024-02-10",
  },
  {
    id: "TMP-002",
    name: "华南极速达",
    channel: "京东自营",
    billing: "piece",
    status: "enabled",
    defaultRule: { region: "珠三角", fee: "¥15" },
    lastUpdate: "2024-02-08",
  },
  {
    id: "TMP-003",
    name: "跨境保税仓",
    channel: "天猫国际",
    billing: "weight",
    status: "disabled",
    defaultRule: { region: "保税区入仓", fee: "¥28" },
    lastUpdate: "2024-01-30",
  },
]);

const heatmap = ref([
  {
    id: "RT-1",
    name: "华北 → 华东",
    channel: "自营商城",
    baseFee: "¥13",
    extraFee: "¥3 /kg",
    leadTime: "2.2 天",
  },
  {
    id: "RT-2",
    name: "华南 → 华中",
    channel: "京东自营",
    baseFee: "¥15",
    extraFee: "¥2.5 /kg",
    leadTime: "1.8 天",
  },
]);

const keyword = ref("");
const channelFilter = ref("");
const billingFilter = ref<BillingType | "">("");

const channelOptions = computed(() =>
  [{ label: "全部渠道", value: "" }].concat(
    Array.from(new Set(templates.value.map((tpl) => tpl.channel))).map((channel) => ({
      label: channel,
      value: channel,
    })),
  ),
);

const billingOptions = [
  { label: "全部计费方式", value: "" },
  { label: "按重量", value: "weight" },
  { label: "按件数", value: "piece" },
  { label: "体积计费", value: "volume" },
];

const columns = computed<TableColumn<Template>[]>(() => [
  { accessorKey: "name", header: "模板名称" },
  { accessorKey: "channel", header: "适用渠道" },
  {
    accessorKey: "billing",
    header: "计费方式",
    cell: ({ getValue }) => {
      const map: Record<BillingType, string> = {
        weight: "按重量",
        piece: "按件数",
        volume: "按体积",
      };
      return map[getValue()];
    },
  },
  { accessorKey: "defaultRule", header: "基础规则" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "lastUpdate", header: "最近更新" },
  { id: "actions", header: "操作" },
]);

const filteredTemplates = computed(() =>
  templates.value.filter((tpl) => {
    const matchesKeyword =
      !keyword.value ||
      tpl.name.includes(keyword.value) ||
      tpl.channel.includes(keyword.value);
    const matchesChannel = !channelFilter.value || tpl.channel === channelFilter.value;
    const matchesBilling = !billingFilter.value || tpl.billing === billingFilter.value;
    return matchesKeyword && matchesChannel && matchesBilling;
  }),
);
</script>
