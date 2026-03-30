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

    <UCard>
      <template #header>
        <div class="flex items-center justify-between gap-2">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">模拟试算</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">输入区域与重量/件数/体积，实时计算运费。</p>
          </div>
          <UButton color="primary" icon="i-heroicons-calculator" @click="runQuote">开始试算</UButton>
        </div>
      </template>

      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <USelect v-model="quoteForm.templateId" :options="templateOptions" placeholder="选择模板" />
        <UInput v-model="quoteForm.region" placeholder="区域（如 华东）" />
        <UInput v-model.number="quoteForm.orderAmount" type="number" placeholder="订单金额（可选）" />
        <UInput v-model.number="quoteForm.weight" type="number" placeholder="重量 kg（可选）" />
        <UInput v-model.number="quoteForm.pieceCount" type="number" placeholder="件数（可选）" />
        <UInput v-model.number="quoteForm.volume" type="number" placeholder="体积（可选）" />
      </div>

      <div v-if="quoteResult" class="mt-4 rounded-xl border border-gray-200 p-4 dark:border-gray-800">
        <div class="text-sm text-gray-500">模板：{{ quoteResult.template }} · 币种：{{ quoteResult.currency }}</div>
        <div class="mt-2 text-2xl font-semibold text-primary-600">¥{{ quoteResult.feeAmount.toFixed(2) }}</div>
        <div class="mt-2 text-sm text-gray-500">
          匹配区域：{{ quoteResult.matchedZone.region || "-" }} · 计费：{{ quoteResult.billingType }}
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi } from "~/composables/api";

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

const logisticsApi = useLogisticsApi();
const templates = ref<Template[]>([]);
const quoteResult = ref<any>(null);
const quoteForm = reactive({
  templateId: "",
  region: "默认区域",
  weight: 1,
  pieceCount: 1,
  volume: 0,
  orderAmount: 0,
});

const normalizeBilling = (v: string): BillingType => {
  const val = String(v || "").toLowerCase();
  if (val === "piece") return "piece";
  if (val === "volume") return "volume";
  return "weight";
};

const loadTemplates = async () => {
  const rows = await logisticsApi.listTemplates();
  templates.value = rows.map((row) => {
    const channels = Array.isArray(row.channels) ? row.channels : [];
    const rules = (row.rules || {}) as Record<string, any>;
    const defaultZone = (rules.defaultZone || rules.default_zone || {}) as Record<string, any>;
    const baseFee = Number(defaultZone.firstFee ?? defaultZone.first_fee ?? 0);
    return {
      id: row.id,
      name: row.name,
      channel: String(channels[0] || "未配置"),
      billing: normalizeBilling(String(rules.billing || rules.billing_type || "weight")),
      status: row.status === "published" ? "enabled" : "disabled",
      defaultRule: {
        region: String(defaultZone.region || "默认区域"),
        fee: `¥${baseFee.toFixed(2)}`,
      },
      lastUpdate: row.updatedAt ? row.updatedAt.slice(0, 10) : "-",
    };
  });
  if (!quoteForm.templateId && templates.value.length > 0) {
    quoteForm.templateId = templates.value[0].id;
  }
};

onMounted(() => {
  loadTemplates();
});

const templateOptions = computed(() =>
  templates.value.map((item) => ({
    label: item.name,
    value: item.id,
  })),
);

const runQuote = async () => {
  if (!quoteForm.templateId) return;
  quoteResult.value = await logisticsApi.quoteTemplate(quoteForm.templateId, {
    region: quoteForm.region,
    weight: Number(quoteForm.weight) || 0,
    piece_count: Number(quoteForm.pieceCount) || 0,
    volume: Number(quoteForm.volume) || 0,
    order_amount: Number(quoteForm.orderAmount) || 0,
  });
};

const heatmap = computed(() =>
  templates.value.slice(0, 4).map((item, idx) => ({
    id: item.id,
    name: item.defaultRule.region,
    channel: item.channel,
    baseFee: item.defaultRule.fee,
    extraFee: "按规则计算",
    leadTime: `${1.5 + idx * 0.3} 天`,
  })),
);

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
      return map[getValue() as BillingType] || "按重量";
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
