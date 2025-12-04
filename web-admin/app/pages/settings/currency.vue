<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">货币与小数精度</h1>
        <p class="text-gray-500 dark:text-gray-400">
          配置主币种、展示精度以及各渠道的汇率策略。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton variant="ghost" color="neutral" icon="i-heroicons-arrow-path">
          同步汇率
        </UButton>
        <UButton color="primary" icon="i-heroicons-check">
          保存
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">基础配置</h3>
      </template>
      <div class="grid gap-4 md:grid-cols-2">
        <UFormGroup label="默认币种">
          <USelect v-model="baseCurrency" :options="currencyOptions" />
        </UFormGroup>
        <UFormGroup label="显示精度">
          <USelect v-model="decimalPrecision" :options="precisionOptions" />
        </UFormGroup>
        <UFormGroup label="四舍五入规则">
          <USelect v-model="roundingMode" :options="roundingOptions" />
        </UFormGroup>
        <UFormGroup label="税前展示">
          <UToggle v-model="displayWithoutTax" />
          <p class="mt-1 text-xs text-gray-500">
            开启后列表默认为税前价格，可在详情页切换。
          </p>
        </UFormGroup>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">汇率表</h3>
          <UButton size="sm" variant="soft" icon="i-heroicons-plus">
            新增币种
          </UButton>
        </div>
      </template>
      <UTable :columns="rateColumns" :data="currencyRates">
        <template #rate-cell="{ getValue }">
          {{ getValue() }}
        </template>
        <template #autoUpdate-cell="{ row }">
          <UToggle v-model="row.original.autoUpdate" />
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">编辑</UButton>
            <UButton size="xs" variant="ghost" color="error">删除</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道覆盖</h3>
          <UButton size="sm" variant="ghost">导出映射</UButton>
        </div>
      </template>
      <div class="grid gap-4 md:grid-cols-2">
        <UCard v-for="channel in currencyChannels" :key="channel.name">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">
                  {{ channel.name }}
                </p>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  默认币种：{{ channel.currency }}
                </p>
              </div>
              <UBadge
                :color="channel.mode === 'auto' ? 'success' : 'neutral'"
                variant="subtle"
              >
                {{ channel.mode === "auto" ? "自动" : "手动" }}
              </UBadge>
            </div>
          </template>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ channel.desc }}
          </p>
        </UCard>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "settings-currency",
});

const baseCurrency = ref("CNY");
const decimalPrecision = ref("2");
const roundingMode = ref("half-up");
const displayWithoutTax = ref(false);

const currencyOptions = [
  { label: "人民币 (CNY)", value: "CNY" },
  { label: "美元 (USD)", value: "USD" },
  { label: "港币 (HKD)", value: "HKD" },
  { label: "欧元 (EUR)", value: "EUR" },
];
const precisionOptions = [
  { label: "0 位", value: "0" },
  { label: "1 位", value: "1" },
  { label: "2 位", value: "2" },
  { label: "3 位", value: "3" },
];
const roundingOptions = [
  { label: "四舍五入", value: "half-up" },
  { label: "向上取整", value: "ceil" },
  { label: "向下取整", value: "floor" },
];

type CurrencyRate = {
  code: string;
  name: string;
  rate: number;
  autoUpdate: boolean;
  updatedAt: string;
};

const currencyRates = ref<CurrencyRate[]>([
  { code: "USD", name: "美元", rate: 0.139, autoUpdate: true, updatedAt: "09:30" },
  { code: "HKD", name: "港币", rate: 1.09, autoUpdate: true, updatedAt: "09:30" },
  { code: "EUR", name: "欧元", rate: 0.128, autoUpdate: false, updatedAt: "昨天" },
]);

const rateColumns = computed<TableColumn<CurrencyRate>[]>(() => [
  { accessorKey: "code", header: "币种" },
  { accessorKey: "name", header: "名称" },
  { accessorKey: "rate", header: "对 CNY 汇率" },
  { accessorKey: "updatedAt", header: "更新于" },
  { accessorKey: "autoUpdate", header: "自动更新" },
  { id: "actions", header: "操作" },
]);

const currencyChannels = ref([
  {
    name: "天猫旗舰店",
    currency: "CNY",
    mode: "auto",
    desc: "默认人民币结算，可根据用户 IP 提示换算。",
  },
  {
    name: "京东自营",
    currency: "CNY",
    mode: "auto",
    desc: "同步京东平台自动转换策略。",
  },
  {
    name: "跨境独立站",
    currency: "USD",
    mode: "manual",
    desc: "统一按 USD 结算，后台手动维护汇率。",
  },
]);
</script>
