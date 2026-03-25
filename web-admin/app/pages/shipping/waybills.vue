<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">运单与轨迹</h1>
        <p class="text-gray-500 dark:text-gray-400">
          聚合承运商运单号，实时追踪节点并监测异常。
        </p>
      </div>
      <div class="flex gap-2">
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path">
          刷新轨迹
        </UButton>
        <UButton color="primary" icon="i-heroicons-plus">录入运单</UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">运单列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              追踪揽收、运输、签收以及异常节点。
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UInput
              v-model="keyword"
              class="w-56"
              placeholder="订单号/运单号"
              icon="i-heroicons-magnifying-glass"
            />
            <USelect v-model="carrierFilter" class="w-44" :options="carrierOptions" />
            <USelect v-model="statusFilter" class="w-40" :options="statusOptions" />
          </div>
        </div>
      </template>

      <UTable :columns="columns" :data="filteredWaybills">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #progress-cell="{ getValue }">
          <div class="flex items-center gap-2">
            <UProgress :value="getValue()" size="xs" class="flex-1" />
            <span class="text-xs text-gray-500">{{ getValue() }}%</span>
          </div>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" @click="selectWaybill(row.original)">
              查看轨迹
            </UButton>
            <UButton size="xs" variant="ghost" color="primary">通知客户</UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard v-if="selectedWaybill">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              轨迹 - {{ selectedWaybill.waybillNo }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              订单 {{ selectedWaybill.orderNo }} · {{ selectedWaybill.carrier }}
            </p>
          </div>
          <UBadge :color="statusMeta(selectedWaybill.status).color" variant="subtle">
            {{ statusMeta(selectedWaybill.status).label }}
          </UBadge>
        </div>
      </template>

      <ol class="space-y-4">
        <li
          v-for="node in selectedWaybill.timeline"
          :key="node.time"
          class="flex gap-4 rounded-xl border border-gray-100 p-4 dark:border-gray-800"
        >
          <div class="flex-none text-sm text-gray-500 dark:text-gray-400">
            <p class="font-semibold text-gray-900 dark:text-white">{{ node.time }}</p>
            <p>{{ node.city }}</p>
          </div>
          <div class="flex-1">
            <p class="font-medium text-gray-900 dark:text-white">{{ node.status }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ node.message }}</p>
          </div>
          <UBadge v-if="node.exception" color="warning" variant="subtle">
            {{ node.exception }}
          </UBadge>
        </li>
      </ol>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi } from "~/composables/api";

definePageMeta({
  name: "shipping-waybills",
});

type WaybillStatus = "created" | "in-transit" | "delay" | "delivered";

type TimelineNode = {
  time: string;
  city: string;
  status: string;
  message: string;
  exception?: string;
};

type Waybill = {
  id: string;
  orderNo: string;
  waybillNo: string;
  carrier: string;
  channel: string;
  status: WaybillStatus;
  progress: number;
  eta: string;
  timeline: TimelineNode[];
};

const logisticsApi = useLogisticsApi();
const waybills = ref<Waybill[]>([]);
const carrierMap = ref<Record<string, string>>({});

const normalizeStatus = (status: string): WaybillStatus => {
  const v = String(status || "").toLowerCase();
  if (v === "delivered" || v === "signed") return "delivered";
  if (v === "delay" || v === "exception") return "delay";
  if (v === "in_transit" || v === "in-transit" || v === "shipping") return "in-transit";
  return "created";
};

const progressByStatus = (status: WaybillStatus): number => {
  switch (status) {
    case "delivered":
      return 100;
    case "delay":
      return 45;
    case "in-transit":
      return 60;
    default:
      return 15;
  }
};

const loadWaybills = async () => {
  const [carriers, rows] = await Promise.all([
    logisticsApi.listCarriers(),
    logisticsApi.listWaybills(),
  ]);
  carrierMap.value = Object.fromEntries(carriers.map((item) => [item.id, item.name]));
  waybills.value = rows.map((row) => {
    const status = normalizeStatus(row.status);
    return {
      id: row.id,
      orderNo: row.orderId,
      waybillNo: row.waybillNo,
      carrier: carrierMap.value[row.carrierId] || row.carrierId,
      channel: row.serviceCode || "标准",
      status,
      progress: progressByStatus(status),
      eta: status === "delivered" ? "已签收" : "待更新",
      timeline: [],
    };
  });
};

const loadWaybillDetail = async (waybill: Waybill) => {
  const detail = await logisticsApi.getWaybillDetail(waybill.id);
  const timeline: TimelineNode[] = detail.tracking.map((node) => ({
    time: node.occurredAt ? node.occurredAt.replace("T", " ").slice(0, 16) : "-",
    city: String(node.payload?.city || "-"),
    status: node.status,
    message: node.description || "-",
    exception: node.status === "exception" ? "异常" : undefined,
  }));
  waybill.timeline = timeline;
};

const keyword = ref("");
const carrierFilter = ref("");
const statusFilter = ref<WaybillStatus | "">("");
const selectedWaybill = ref<Waybill | null>(null);

const carrierOptions = computed(() =>
  [{ label: "全部承运商", value: "" }].concat(
    Array.from(new Set(waybills.value.map((waybill) => waybill.carrier))).map((carrier) => ({
      label: carrier,
      value: carrier,
    })),
  ),
);

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "待揽收", value: "created" },
  { label: "运输中", value: "in-transit" },
  { label: "延误预警", value: "delay" },
  { label: "已签收", value: "delivered" },
];

const columns = computed<TableColumn<Waybill>[]>(() => [
  { accessorKey: "orderNo", header: "订单号" },
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "carrier", header: "承运商" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "progress", header: "进度" },
  { accessorKey: "eta", header: "预计到达" },
  { id: "actions", header: "操作" },
]);

const filteredWaybills = computed(() =>
  waybills.value.filter((waybill) => {
    const matchesKeyword =
      !keyword.value ||
      waybill.orderNo.includes(keyword.value) ||
      waybill.waybillNo.includes(keyword.value);
    const matchesCarrier = !carrierFilter.value || waybill.carrier === carrierFilter.value;
    const matchesStatus = !statusFilter.value || waybill.status === statusFilter.value;
    return matchesKeyword && matchesCarrier && matchesStatus;
  }),
);

watch(
  filteredWaybills,
  async (list) => {
    if (!selectedWaybill.value && list.length) {
      selectedWaybill.value = list[0];
      await loadWaybillDetail(list[0]);
      return;
    }
    if (
      selectedWaybill.value &&
      !list.some((waybill) => waybill.waybillNo === selectedWaybill.value?.waybillNo)
    ) {
      selectedWaybill.value = list[0] ?? null;
      if (list[0]) {
        await loadWaybillDetail(list[0]);
      }
    }
  },
  { immediate: true },
);

const selectWaybill = async (waybill: Waybill) => {
  selectedWaybill.value = waybill;
  await loadWaybillDetail(waybill);
};

onMounted(async () => {
  await loadWaybills();
});

const statusMeta = (status: WaybillStatus | "") => {
  switch (status) {
    case "created":
      return { label: "待揽收", color: "neutral" as const };
    case "in-transit":
      return { label: "运输中", color: "info" as const };
    case "delay":
      return { label: "延误", color: "warning" as const };
    case "delivered":
      return { label: "已签收", color: "success" as const };
    default:
      return { label: "未知", color: "neutral" as const };
  }
};
</script>

