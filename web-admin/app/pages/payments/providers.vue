<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">支付渠道</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理线上收单、分期、钱包等支付服务，监控费率与 SLA。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-document-arrow-down">
          导出配置
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">新增渠道</UButton>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <UCard v-for="card in summaryCards" :key="card.title">
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ card.title }}</p>
        <div class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
          {{ card.value }}
        </div>
        <p class="text-xs" :class="card.trend >= 0 ? 'text-emerald-600' : 'text-rose-500'">
          {{ card.trend >= 0 ? '+' : '' }}{{ card.trend }}% 较上周
        </p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">支付服务商</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">查看接入状态、费率及 SLA。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-52"
              placeholder="搜索渠道/联系人"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="typeFilter"
              class="w-40"
              :options="typeOptions"
              placeholder="支付类型"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :options="statusOptions"
              placeholder="服务状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredProviders">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #sla-cell="{ getValue }">
          <div class="flex items-center gap-2">
            <UProgress :value="getValue()" size="xs" class="flex-1" />
            <span class="text-xs text-gray-500">{{ getValue() }}%</span>
          </div>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">配置</UButton>
            <UButton size="xs" variant="ghost" color="primary">查看账单</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">费率日历</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">密切关注促销期的特殊费率。</p>
          </div>
          <UBadge color="info" variant="subtle">{{ rateEvents.length }} 条</UBadge>
        </div>
      </template>

      <ul class="space-y-4">
        <li
          v-for="event in rateEvents"
          :key="event.id"
          class="flex flex-col rounded-xl border border-gray-100 p-4 dark:border-gray-800 md:flex-row md:items-center md:justify-between"
        >
          <div>
            <p class="font-medium text-gray-900 dark:text-white">{{ event.title }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ event.range }}</p>
          </div>
          <div class="text-sm text-gray-500 dark:text-gray-400">
            当前费率
            <span class="ml-1 text-gray-900 dark:text-white">{{ event.rate }}</span>
          </div>
          <div class="flex gap-2">
            <UButton size="xs" variant="soft">查看详情</UButton>
            <UButton size="xs" variant="ghost">忽略</UButton>
          </div>
        </li>
      </ul>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "payments-providers",
});

type ProviderStatus = "active" | "monitor" | "paused";

type Provider = {
  id: string;
  name: string;
  type: string;
  contact: string;
  phone: string;
  feeRate: string;
  status: ProviderStatus;
  sla: number;
};

const providers = ref<Provider[]>([
  {
    id: "WX-PAY",
    name: "微信支付",
    type: "移动支付",
    contact: "李思思",
    phone: "400-800-1234",
    feeRate: "0.38%",
    status: "active",
    sla: 99.9,
  },
  {
    id: "ALI-PAY",
    name: "支付宝",
    type: "移动支付",
    contact: "赵伟",
    phone: "95188",
    feeRate: "0.35%",
    status: "active",
    sla: 99.7,
  },
  {
    id: "INSTALLMENT",
    name: "花呗分期",
    type: "分期",
    contact: "刘敏",
    phone: "400-666-8888",
    feeRate: "0.60%",
    status: "monitor",
    sla: 98.5,
  },
  {
    id: "BNPL",
    name: "分期乐",
    type: "先享后付",
    contact: "王浩",
    phone: "400-111-9999",
    feeRate: "0.75%",
    status: "paused",
    sla: 95.2,
  },
]);

const keyword = ref("");
const typeFilter = ref("");
const statusFilter = ref<ProviderStatus | "">("");

const typeOptions = computed(() =>
  [{ label: "全部类型", value: "" }].concat(
    Array.from(new Set(providers.value.map((provider) => provider.type))).map((type) => ({
      label: type,
      value: type,
    })),
  ),
);

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "合作中", value: "active" },
  { label: "观察中", value: "monitor" },
  { label: "暂停", value: "paused" },
];

const columns = computed<TableColumn<Provider>[]>(() => [
  { accessorKey: "name", header: "渠道" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "feeRate", header: "费率" },
  {
    accessorKey: "contact",
    header: "商务",
    cell: ({ row }) => `${row.original.contact} / ${row.original.phone}`,
  },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "sla", header: "SLA" },
  { id: "actions", header: "操作" },
]);

const filteredProviders = computed(() =>
  providers.value.filter((provider) => {
    const matchesKeyword =
      !keyword.value ||
      provider.name.includes(keyword.value) ||
      provider.contact.includes(keyword.value);
    const matchesType = !typeFilter.value || provider.type === typeFilter.value;
    const matchesStatus = !statusFilter.value || provider.status === statusFilter.value;
    return matchesKeyword && matchesType && matchesStatus;
  }),
);

const statusMeta = (status: ProviderStatus | "") => {
  switch (status) {
    case "active":
      return { label: "合作中", color: "success" as const };
    case "monitor":
      return { label: "观察中", color: "warning" as const };
    case "paused":
      return { label: "暂停", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => [
  {
    title: "合作支付渠道",
    value: providers.value.filter((p) => p.status === "active").length,
    trend: 1.5,
  },
  {
    title: "平均 SLA",
    value:
      (
        providers.value.reduce((sum, provider) => sum + provider.sla, 0) / providers.value.length
      ).toFixed(1) + "%",
    trend: 0.4,
  },
  {
    title: "观察/暂停渠道",
    value: providers.value.filter((p) => p.status !== "active").length,
    trend: -2.1,
  },
]);

const rateEvents = ref([
  {
    id: "EV-01",
    title: "女王节活动费率",
    range: "3 月 1 日 - 3 月 8 日",
    rate: "0.32%",
  },
  {
    id: "EV-02",
    title: "双 11 全链路补贴",
    range: "11 月 1 日 - 11 月 12 日",
    rate: "0.28%",
  },
]);
</script>
