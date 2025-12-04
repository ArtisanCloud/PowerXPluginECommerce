<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">店铺 / 渠道</h1>
        <p class="text-gray-500 dark:text-gray-400">
          管理各平台店铺、授权与接入状态，监控渠道健康度。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path">
          同步授权
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">
          接入新渠道
        </UButton>
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
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              支持按平台、授权状态、运营负责人筛选。
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-56"
              placeholder="搜索店铺/负责人"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect
              v-model="platformFilter"
              class="w-44"
              :options="platformOptions"
              placeholder="全部平台"
            />
            <USelect
              v-model="statusFilter"
              class="w-40"
              :options="statusOptions"
              placeholder="授权状态"
            />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredChannels">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #syncStatus-cell="{ getValue }">
          <div class="flex items-center gap-2">
            <UIcon
              :name="
                getValue() === 'success'
                  ? 'i-heroicons-check-circle'
                  : getValue() === 'warning'
                    ? 'i-heroicons-exclamation-triangle'
                    : 'i-heroicons-arrow-path'
              "
              :class="[
                'h-4 w-4',
                getValue() === 'success'
                  ? 'text-emerald-500'
                  : getValue() === 'warning'
                    ? 'text-amber-500'
                    : 'text-gray-400',
              ]"
            />
            <span class="text-sm text-gray-600 dark:text-gray-300">
              {{
                getValue() === "success"
                  ? "同步正常"
                  : getValue() === "warning"
                    ? "待处理"
                    : "同步中"
              }}
            </span>
          </div>
        </template>
        <template #actions-cell>
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost">设置</UButton>
            <UButton size="xs" variant="ghost" color="primary">检查授权</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <div class="grid gap-6 lg:grid-cols-2">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">渠道上线 Checklist</h3>
            <UBadge color="info" variant="subtle">{{ checklist.length }} 项</UBadge>
          </div>
        </template>
        <ul class="space-y-4">
          <li
            v-for="item in checklist"
            :key="item.id"
            class="flex items-start gap-3 rounded-xl border border-gray-100 p-4 dark:border-gray-800"
          >
            <UIcon
              :name="item.done ? 'i-heroicons-check-circle-solid' : 'i-heroicons-clock'"
              :class="[
                'h-5 w-5',
                item.done ? 'text-emerald-500' : 'text-gray-400 dark:text-gray-500',
              ]"
            />
            <div>
              <p class="font-medium text-gray-900 dark:text-white">{{ item.title }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ item.desc }}</p>
            </div>
          </li>
        </ul>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                渠道运营提醒
              </h3>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                预警低库存/活动节点，提醒提前准备。
              </p>
            </div>
            <UBadge color="warning" variant="subtle">重要</UBadge>
          </div>
        </template>

        <ul class="space-y-4">
          <li
            v-for="alert in alerts"
            :key="alert.id"
            class="rounded-xl border border-amber-100 p-4 dark:border-amber-900/40"
          >
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ alert.title }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ alert.date }}</span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ alert.desc }}</p>
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
  name: "channels",
});

type ChannelStatus = "active" | "pending" | "disabled";
type SyncStatus = "success" | "warning" | "syncing";

type Channel = {
  id: string;
  name: string;
  platform: string;
  region: string;
  status: ChannelStatus;
  owner: string;
  syncStatus: SyncStatus;
  updatedAt: string;
};

const channels = ref<Channel[]>([
  {
    id: "TMALL-01",
    name: "天猫旗舰店",
    platform: "天猫",
    region: "全国",
    status: "active",
    owner: "陈曦",
    syncStatus: "success",
    updatedAt: "2024-02-10 12:30",
  },
  {
    id: "JD-POP-02",
    name: "京东自营旗舰",
    platform: "京东",
    region: "全国",
    status: "active",
    owner: "王帆",
    syncStatus: "syncing",
    updatedAt: "2024-02-11 09:10",
  },
  {
    id: "TM-OVERSEA",
    name: "天猫国际店",
    platform: "天猫国际",
    region: "跨境",
    status: "pending",
    owner: "李倩",
    syncStatus: "warning",
    updatedAt: "2024-02-09 21:40",
  },
  {
    id: "DOUYIN-01",
    name: "抖音旗舰店",
    platform: "抖音电商",
    region: "全国",
    status: "disabled",
    owner: "周杨",
    syncStatus: "warning",
    updatedAt: "2024-02-06 08:00",
  },
]);

const keyword = ref("");
const platformFilter = ref("");
const statusFilter = ref<ChannelStatus | "">("");

const platformOptions = computed(() =>
  [{ label: "全部平台", value: "" }].concat(
    Array.from(new Set(channels.value.map((ch) => ch.platform))).map((platform) => ({
      label: platform,
      value: platform,
    })),
  ),
);

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "已上线", value: "active" },
  { label: "待上线", value: "pending" },
  { label: "暂停", value: "disabled" },
];

const columns = computed<TableColumn<Channel>[]>(() => [
  { accessorKey: "name", header: "店铺" },
  { accessorKey: "platform", header: "平台" },
  { accessorKey: "region", header: "区域" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "syncStatus", header: "同步状态" },
  { accessorKey: "owner", header: "负责人" },
  { accessorKey: "updatedAt", header: "最近同步" },
  { id: "actions", header: "操作" },
]);

const filteredChannels = computed(() =>
  channels.value.filter((channel) => {
    const matchesKeyword =
      !keyword.value ||
      channel.name.includes(keyword.value) ||
      channel.owner.includes(keyword.value);
    const matchesPlatform = !platformFilter.value || channel.platform === platformFilter.value;
    const matchesStatus = !statusFilter.value || channel.status === statusFilter.value;
    return matchesKeyword && matchesPlatform && matchesStatus;
  }),
);

const statusMeta = (status: ChannelStatus | "") => {
  switch (status) {
    case "active":
      return { label: "已上线", color: "success" as const };
    case "pending":
      return { label: "待上线", color: "info" as const };
    case "disabled":
      return { label: "暂停", color: "neutral" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};

const summaryCards = computed(() => [
  {
    title: "已上线渠道",
    value: channels.value.filter((item) => item.status === "active").length,
    trend: 3.1,
  },
  {
    title: "待上线渠道",
    value: channels.value.filter((item) => item.status === "pending").length,
    trend: -1.4,
  },
  {
    title: "同步健康度",
    value:
      Math.round(
        (channels.value.filter((item) => item.syncStatus === "success").length /
          channels.value.length) *
          100,
      ) + "%",
    trend: 0.8,
  },
]);

const checklist = ref([
  { id: 1, title: "完成平台授权", desc: "天猫旗舰店授权将于 2.20 过期，请提前续约", done: true },
  { id: 2, title: "上传商品素材", desc: "京东旗舰店需补齐 25 个 SKU 的详情图", done: false },
  { id: 3, title: "配置营销活动", desc: "3 月新品联动渠道待配置活动页", done: false },
]);

const alerts = ref([
  {
    id: "AL-1",
    title: "京东仓低库存提醒",
    desc: "SKU JD123 距安全库存仅剩 3 天，请提前补货。",
    date: "今天 10:00",
  },
  {
    id: "AL-2",
    title: "抖音直播大促",
    desc: "本周五 20:00 直播活动需要提前完成素材审核与价格锁定。",
    date: "周三 14:30",
  },
]);
</script>
