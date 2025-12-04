<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">SKU 映射</h1>
        <p class="text-gray-500 dark:text-gray-400">
          维护平台 SKU 与内部 SPU/SKU 的对应关系，保障库存与价格同步。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-document-arrow-down">
          导出映射
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          新建映射
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">映射列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              支持按渠道、商品、同步状态过滤。
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-60"
              placeholder="内部 SKU / 渠道 SKU"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="channelFilter"
              class="w-44"
              :options="channelOptions"
              placeholder="全部渠道"
            />
            <USelect
              v-model="syncFilter"
              class="w-40"
              :options="syncOptions"
              placeholder="同步状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredMappings">
        <template #syncStatus-cell="{ getValue }">
          <UBadge :color="syncMeta(getValue()).color" variant="subtle">
            {{ syncMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">查看</UButton>
            <UButton size="xs" variant="ghost" color="primary">重新同步</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">映射建议</h3>
            <UBadge color="info" variant="subtle">{{ suggestions.length }} 条推荐</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="item in suggestions"
            :key="item.id"
            class="rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ item.product }}</p>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ item.channel }}</p>
              </div>
              <UBadge color="primary" variant="subtle">相似度 {{ item.score }}%</UBadge>
            </div>
            <dl class="mt-3 text-sm text-gray-500 dark:text-gray-400">
              <div class="flex justify-between">
                <dt>内部 SKU</dt>
                <dd class="text-gray-900 dark:text-white">{{ item.internalSku }}</dd>
              </div>
              <div class="flex justify-between">
                <dt>渠道 SKU</dt>
                <dd class="text-gray-900 dark:text-white">{{ item.channelSku }}</dd>
              </div>
            </dl>
            <div class="mt-3 flex gap-2">
              <UButton size="xs" color="primary" variant="soft">接受建议</UButton>
              <UButton size="xs" variant="ghost">忽略</UButton>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">同步健康度</h3>
            <UBadge color="success" variant="subtle">{{ health.score }}%</UBadge>
          </div>
        </template>
        <div class="space-y-4">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">库存同步</p>
            <UProgress :value="health.stock" size="sm" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">价格同步</p>
            <UProgress :value="health.price" size="sm" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">属性映射</p>
            <UProgress :value="health.attr" size="sm" />
          </div>
          <ul class="space-y-2 text-sm text-gray-500 dark:text-gray-400">
            <li>• 2 个 SKU 属性缺失需补齐</li>
            <li>• 5 个渠道 SKU 库存同步延迟超过 15 分钟</li>
          </ul>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

definePageMeta({
  name: "channels-sku-mapping",
});

type SyncState = "ok" | "delay" | "missing";

type Mapping = {
  id: string;
  product: string;
  internalSku: string;
  channelSku: string;
  channel: string;
  syncStatus: SyncState;
  lastSync: string;
};

const mappings = ref<Mapping[]>([
  {
    id: "MAP-1001",
    product: "iPhone 15 Pro 256G",
    internalSku: "IP15P-256-BLK",
    channelSku: "TM-IPH15P-256",
    channel: "天猫旗舰店",
    syncStatus: "ok",
    lastSync: "5 分钟前",
  },
  {
    id: "MAP-1002",
    product: "MacBook Air M3 13\"",
    internalSku: "MBA-M3-8G",
    channelSku: "JD-MBA-2024",
    channel: "京东自营",
    syncStatus: "delay",
    lastSync: "32 分钟前",
  },
  {
    id: "MAP-1003",
    product: "Nintendo Switch OLED",
    internalSku: "NS-OLED-WHT",
    channelSku: "DY-NS-OLED",
    channel: "抖音旗舰店",
    syncStatus: "missing",
    lastSync: "待匹配",
  },
  {
    id: "MAP-1004",
    product: "AirPods Pro 2",
    internalSku: "APP2-WHT",
    channelSku: "TM-APPRO2",
    channel: "天猫旗舰店",
    syncStatus: "ok",
    lastSync: "12 分钟前",
  },
]);

const keyword = ref("");
const channelFilter = ref("");
const syncFilter = ref<SyncState | "">("");

const channelOptions = computed(() =>
  [{ label: "全部渠道", value: "" }].concat(
    Array.from(new Set(mappings.value.map((mapping) => mapping.channel))).map((channel) => ({
      label: channel,
      value: channel,
    })),
  ),
);

const syncOptions = [
  { label: "全部状态", value: "" },
  { label: "同步正常", value: "ok" },
  { label: "延迟", value: "delay" },
  { label: "待匹配", value: "missing" },
];

const columns = computed<TableColumn<Mapping>[]>(() => [
  { accessorKey: "product", header: "商品" },
  { accessorKey: "internalSku", header: "内部 SKU" },
  { accessorKey: "channelSku", header: "渠道 SKU" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "syncStatus", header: "同步状态" },
  { accessorKey: "lastSync", header: "最近同步" },
  { id: "actions", header: "操作" },
]);

const filteredMappings = computed(() =>
  mappings.value.filter((mapping) => {
    const matchesKeyword =
      !keyword.value ||
      mapping.internalSku.toLowerCase().includes(keyword.value.toLowerCase()) ||
      mapping.channelSku.toLowerCase().includes(keyword.value.toLowerCase());
    const matchesChannel = !channelFilter.value || mapping.channel === channelFilter.value;
    const matchesSync = !syncFilter.value || mapping.syncStatus === syncFilter.value;
    return matchesKeyword && matchesChannel && matchesSync;
  }),
);

const syncMeta = (status: SyncState | "") => {
  switch (status) {
    case "ok":
      return { label: "同步正常", color: "success" as const };
    case "delay":
      return { label: "延迟", color: "warning" as const };
    case "missing":
      return { label: "待匹配", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const suggestions = ref([
  {
    id: "SG-101",
    product: "Microsoft Surface Laptop Studio 2",
    internalSku: "MS-SLS2-GRAY",
    channelSku: "JD-SLS2-16G",
    channel: "京东自营",
    score: 93,
  },
  {
    id: "SG-102",
    product: "ROG 幻 16",
    internalSku: "ROG16-RTX4070",
    channelSku: "TM-ROG16-4070",
    channel: "天猫旗舰店",
    score: 88,
  },
]);

const health = reactive({
  score: 86,
  stock: 90,
  price: 82,
  attr: 78,
});
</script>
