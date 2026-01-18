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
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ card.subtitle }}</p>
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
              :items="typeItems"
              placeholder="支付类型"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :items="statusItems"
              placeholder="服务状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredProviders" :loading="loading">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #feeRate-cell="{ getValue }">{{ formatFeeRate(getValue()) }}</template>
        <template #updatedAt-cell="{ getValue }">{{ fmtDT(getValue()) }}</template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">配置</UButton>
            <UButton size="xs" variant="ghost" color="primary">查看账单</UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { usePaymentsApi } from "~/composables/api";
import { useToastAlert } from "~/composables/useToastAlert";
import type { PaymentProvider } from "~/types/payments";

definePageMeta({
  name: "payments-providers",
});

const { listProviders } = usePaymentsApi();
const toast = useToastAlert();
const loading = ref(false);
const providers = ref<PaymentProvider[]>([]);

const keyword = ref("");
const ALL_FILTER = "all";
const typeFilter = ref(ALL_FILTER);
const statusFilter = ref(ALL_FILTER);

const typeItems = computed(() => {
  const types = Array.from(
    new Set(
      providers.value
        .map((provider) => String(provider.type || "").trim())
        .filter((type) => type.length > 0),
    ),
  );
  return [{ label: "全部类型", value: ALL_FILTER }].concat(
    types.map((type) => ({ label: type, value: type })),
  );
});

const statusItems = [
  { label: "全部状态", value: ALL_FILTER },
  { label: "合作中", value: "active" },
  { label: "观察中", value: "monitor" },
  { label: "暂停", value: "paused" },
  { label: "停用", value: "inactive" },
];

const columns = computed<TableColumn<PaymentProvider>[]>(() => [
  { accessorKey: "name", header: "渠道" },
  { accessorKey: "type", header: "类型" },
  { accessorKey: "feeRate", header: "费率" },
  { accessorKey: "currency", header: "币种" },
  { accessorKey: "settlementCycle", header: "结算周期" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const filteredProviders = computed(() =>
  providers.value.filter((provider) => {
    const matchesKeyword =
      !keyword.value ||
      provider.name.includes(keyword.value);
    const matchesType =
      typeFilter.value === ALL_FILTER || provider.type === typeFilter.value;
    const matchesStatus =
      statusFilter.value === ALL_FILTER || provider.status === statusFilter.value;
    return matchesKeyword && matchesType && matchesStatus;
  }),
);

const statusMeta = (status: string | "") => {
  switch (status) {
    case "active":
      return { label: "合作中", color: "success" as const };
    case "monitor":
      return { label: "观察中", color: "warning" as const };
    case "paused":
      return { label: "暂停", color: "neutral" as const };
    case "inactive":
      return { label: "停用", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => [
  {
    title: "支付渠道总数",
    value: providers.value.length,
    subtitle: "全部已接入渠道",
  },
  {
    title: "合作中渠道",
    value: providers.value.filter((p) => p.status === "active").length,
    subtitle: "已上线可用",
  },
  {
    title: "平均费率",
    value: formatFeeRate(
      providers.value.length
        ? providers.value.reduce((sum, provider) => sum + provider.feeRate, 0) /
            providers.value.length
        : 0,
    ),
    subtitle: "根据当前配置",
  },
]);

const fmtDT = (s: string) =>
  s ? new Date(s).toLocaleString("zh-CN", { hour12: false }) : "-";

const formatFeeRate = (value: number) => {
  const rate = Number(value) || 0;
  if (rate <= 0) return "0%";
  const percent = rate <= 1 ? rate * 100 : rate;
  return `${percent.toFixed(2)}%`;
};

const loadProviders = async () => {
  loading.value = true;
  try {
    providers.value = await listProviders();
  } catch (error: any) {
    toast.add({
      title: "获取支付渠道失败",
      description: error?.message || "请稍后重试",
      color: "red",
    });
  } finally {
    loading.value = false;
  }
};

onMounted(loadProviders);
</script>
