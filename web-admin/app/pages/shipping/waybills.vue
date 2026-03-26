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
        <template #orderNo-cell="{ row }">
          <div class="space-y-1">
            <p class="font-medium text-gray-900 dark:text-white">{{ row.original.orderNo }}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ orderSummary(row.original).packageCount }} 包裹 ·
              {{ orderStatusMeta(orderSummary(row.original).aggregateStatus).label }}
            </p>
          </div>
        </template>
        <template #packageNo-cell="{ row }">
          <div class="flex items-center gap-2">
            <UBadge color="neutral" variant="soft">包裹 #{{ row.original.packageNo }}</UBadge>
            <span class="text-xs text-gray-500">{{ row.original.packageKey || "-" }}</span>
          </div>
        </template>
        <template #orderFulfillmentStatus-cell="{ getValue }">
          <UBadge :color="orderStatusMeta(getValue()).color" variant="subtle">
            {{ orderStatusMeta(getValue()).label }}
          </UBadge>
        </template>
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
        <template #eta-cell="{ row }">
          <div class="space-y-0.5 text-xs">
            <p class="text-gray-900 dark:text-white">承诺达：{{ formatEta(row.original.promisedAt) }}</p>
            <p class="text-gray-500 dark:text-gray-400">预计达：{{ formatEta(row.original.estimatedAt) }}</p>
          </div>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" color="neutral" @click="openRoutingPreview(row.original)">
              路由预览
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="openRedelivery(row.original)">
              失败重派
            </UButton>
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

    <UModal v-model:open="routingPreviewOpen" title="仓配路由预览">
      <template #body>
        <div class="space-y-3">
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="routingForm.warehouseId" placeholder="仓库ID（可选）" />
            <UInput v-model="routingForm.destinationZone" placeholder="目的区域（如 CN-EAST）" />
            <UInput v-model="routingForm.serviceCode" placeholder="服务编码（默认 std）" />
            <UInput v-model.number="routingForm.weight" type="number" step="0.1" placeholder="重量(kg)" />
          </div>
          <div class="text-xs text-gray-500">
            运单：{{ routingForm.waybillNo || "-" }}，偏好承运商：{{ routingForm.preferredCarrierName || "未指定" }}
          </div>
          <UButton color="primary" :loading="routingPreviewLoading" @click="previewRouting">
            预览路由
          </UButton>
          <UCard v-if="routingPreviewResult">
            <p class="text-sm">命中策略：{{ routingPreviewResult.strategy }}（{{ routingPreviewResult.reason }}）</p>
            <p class="text-sm">结果承运商：{{ routingPreviewResult.carrierName }} / {{ routingPreviewResult.serviceCode }}</p>
            <p class="text-xs text-gray-500">
              命中规则：{{ routingPreviewResult.matchedRuleName || routingPreviewResult.matchedRuleId || "无（兜底）" }}
            </p>
          </UCard>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="redeliveryOpen" title="妥投失败二次派送">
      <template #body>
        <div class="space-y-3">
          <p class="text-xs text-gray-500">运单 {{ redeliveryForm.waybillNo || "-" }} 的失败重派闭环。</p>
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="redeliveryForm.reason" placeholder="原因（failed_delivery）" />
            <UInput v-model="redeliveryForm.operatorId" placeholder="操作人" />
            <UInput v-model="redeliveryForm.requestKey" placeholder="请求幂等键（可选）" />
            <UInput v-model="redeliveryForm.addressLine" placeholder="改址内容（可选）" />
          </div>
          <div class="flex flex-wrap gap-2">
            <UButton color="warning" :loading="redeliveryLoading" @click="initiateRedelivery">发起</UButton>
            <UButton color="neutral" :loading="redeliveryLoading" :disabled="!redeliveryForm.taskId" @click="updateRedeliveryAddress">
              改址
            </UButton>
            <UButton color="primary" :loading="redeliveryLoading" :disabled="!redeliveryForm.taskId" @click="redispatchRedelivery">
              重派
            </UButton>
            <UButton color="success" :loading="redeliveryLoading" :disabled="!redeliveryForm.taskId" @click="closeRedelivery">
              关闭
            </UButton>
          </div>
          <UCard>
            <template #header>
              <div class="text-sm font-medium">最近任务</div>
            </template>
            <ul class="space-y-1 text-xs">
              <li v-for="task in redeliveryTasks" :key="task.id" class="flex items-center justify-between">
                <span>{{ task.status }} · attempt={{ task.attemptNo }} · {{ task.lastReason || "-" }}</span>
                <UButton size="xs" variant="ghost" @click="selectRedeliveryTask(task)">选择</UButton>
              </li>
              <li v-if="!redeliveryTasks.length" class="text-gray-500">暂无任务</li>
            </ul>
          </UCard>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import type { LogisticsRedeliveryTask, LogisticsWaybillETA } from "~/composables/api/useLogistics";
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
  carrierId: string;
  waybillNo: string;
  packageNo: number;
  packageKey: string;
  orderFulfillmentStatus: "partial_shipped" | "fully_shipped";
  shipmentItems: string[];
  carrier: string;
  channel: string;
  status: WaybillStatus;
  progress: number;
  eta: string;
  promisedAt: string;
  estimatedAt: string;
  timezone: string;
  timeline: TimelineNode[];
};

const logisticsApi = useLogisticsApi();
const waybills = ref<Waybill[]>([]);
const carrierMap = ref<Record<string, string>>({});
const routingPreviewOpen = ref(false);
const routingPreviewLoading = ref(false);
const routingPreviewResult = ref<any>(null);
const redeliveryOpen = ref(false);
const redeliveryLoading = ref(false);
const redeliveryTasks = ref<LogisticsRedeliveryTask[]>([]);
const routingForm = reactive({
  waybillNo: "",
  warehouseId: "",
  destinationZone: "GLOBAL",
  serviceCode: "std",
  weight: 1,
  preferredCarrierId: "",
  preferredCarrierName: "",
});
const redeliveryForm = reactive({
  taskId: "",
  waybillId: "",
  waybillNo: "",
  reason: "failed_delivery",
  operatorId: "admin",
  requestKey: "",
  addressLine: "",
});

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
      carrierId: row.carrierId,
      waybillNo: row.waybillNo,
      packageNo: row.packageNo || 1,
      packageKey: row.packageKey || "",
      orderFulfillmentStatus:
        row.orderFulfillmentStatus === "fully_shipped" ? "fully_shipped" : "partial_shipped",
      shipmentItems: row.shipmentItems || [],
      carrier: carrierMap.value[row.carrierId] || row.carrierId,
      channel: row.serviceCode || "标准",
      status,
      progress: progressByStatus(status),
      eta: status === "delivered" ? "已签收" : "待更新",
      promisedAt: "",
      estimatedAt: "",
      timezone: "UTC",
      timeline: [],
    };
  });
};

const formatEta = (value?: string) => {
  const raw = String(value || "").trim();
  if (!raw) return "-";
  const d = new Date(raw);
  if (Number.isNaN(d.getTime())) return raw;
  return d.toLocaleString("zh-CN", {
    hour12: false,
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const loadWaybillETA = async () => {
  if (!waybills.value.length) return;
  const rows = await logisticsApi.listWaybillETA({
    waybill_ids: waybills.value.map((item) => item.id),
  });
  const etaMap: Record<string, LogisticsWaybillETA> = Object.fromEntries(
    rows.map((item) => [item.waybillId, item]),
  );
  waybills.value = waybills.value.map((item) => {
    const eta = etaMap[item.id];
    if (!eta) return item;
    const next: Waybill = {
      ...item,
      promisedAt: eta.promisedAt || "",
      estimatedAt: eta.estimatedAt || "",
      timezone: eta.timezone || "UTC",
      eta: eta.estimatedAt ? formatEta(eta.estimatedAt) : item.eta,
    };
    if (eta.delayed && next.status !== "delivered") {
      next.status = "delay";
      next.progress = progressByStatus("delay");
    }
    return next;
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
  { accessorKey: "packageNo", header: "包裹" },
  { accessorKey: "carrier", header: "承运商" },
  { accessorKey: "channel", header: "渠道" },
  { accessorKey: "orderFulfillmentStatus", header: "订单履约" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "progress", header: "进度" },
  { accessorKey: "eta", header: "承诺达 / 预计达" },
  { id: "actions", header: "操作" },
]);

const groupedWaybills = computed(() =>
  [...waybills.value].sort((a, b) => {
    if (a.orderNo === b.orderNo) return a.packageNo - b.packageNo;
    return a.orderNo.localeCompare(b.orderNo);
  }),
);

const filteredWaybills = computed(() =>
  groupedWaybills.value.filter((waybill) => {
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
  await loadWaybillETA();
});

const openRoutingPreview = (waybill: Waybill) => {
  routingForm.waybillNo = waybill.waybillNo;
  routingForm.serviceCode = waybill.channel || "std";
  routingForm.preferredCarrierId = waybill.carrierId;
  routingForm.preferredCarrierName = waybill.carrier;
  routingPreviewResult.value = null;
  routingPreviewOpen.value = true;
};

const previewRouting = async () => {
  routingPreviewLoading.value = true;
  try {
    routingPreviewResult.value = await logisticsApi.previewRouting({
      warehouse_id: routingForm.warehouseId || undefined,
      destination_zone: routingForm.destinationZone || undefined,
      service_code: routingForm.serviceCode || undefined,
      weight: Number(routingForm.weight || 0),
      preferred_carrier_id: routingForm.preferredCarrierId || undefined,
    });
  } finally {
    routingPreviewLoading.value = false;
  }
};

const loadRedeliveryTasks = async (waybillId: string) => {
  redeliveryTasks.value = await logisticsApi.listRedeliveryTasks({ waybill_id: waybillId });
  if (redeliveryTasks.value.length && !redeliveryForm.taskId) {
    redeliveryForm.taskId = redeliveryTasks.value[0].id;
  }
};

const openRedelivery = async (waybill: Waybill) => {
  redeliveryForm.taskId = "";
  redeliveryForm.waybillId = waybill.id;
  redeliveryForm.waybillNo = waybill.waybillNo;
  redeliveryForm.reason = "failed_delivery";
  redeliveryForm.operatorId = "admin";
  redeliveryForm.requestKey = "";
  redeliveryForm.addressLine = "";
  redeliveryOpen.value = true;
  await loadRedeliveryTasks(waybill.id);
};

const selectRedeliveryTask = (task: LogisticsRedeliveryTask) => {
  redeliveryForm.taskId = task.id;
  redeliveryForm.reason = task.lastReason || redeliveryForm.reason;
};

const initiateRedelivery = async () => {
  if (!redeliveryForm.waybillId) return;
  redeliveryLoading.value = true;
  try {
    const resp = await logisticsApi.initiateRedeliveryTask({
      waybill_id: redeliveryForm.waybillId,
      request_key: redeliveryForm.requestKey || undefined,
      reason: redeliveryForm.reason || undefined,
      operator_id: redeliveryForm.operatorId || undefined,
      address: redeliveryForm.addressLine ? { line1: redeliveryForm.addressLine } : undefined,
    });
    redeliveryForm.taskId = resp.task.id;
    await loadRedeliveryTasks(redeliveryForm.waybillId);
  } finally {
    redeliveryLoading.value = false;
  }
};

const updateRedeliveryAddress = async () => {
  if (!redeliveryForm.taskId) return;
  redeliveryLoading.value = true;
  try {
    await logisticsApi.updateRedeliveryAddress(redeliveryForm.taskId, {
      address: { line1: redeliveryForm.addressLine || "updated" },
      operator_id: redeliveryForm.operatorId || undefined,
      reason: redeliveryForm.reason || undefined,
    });
    await loadRedeliveryTasks(redeliveryForm.waybillId);
  } finally {
    redeliveryLoading.value = false;
  }
};

const redispatchRedelivery = async () => {
  if (!redeliveryForm.taskId) return;
  redeliveryLoading.value = true;
  try {
    await logisticsApi.redispatchRedeliveryTask(redeliveryForm.taskId, {
      request_key: redeliveryForm.requestKey || undefined,
      operator_id: redeliveryForm.operatorId || undefined,
      reason: redeliveryForm.reason || undefined,
    });
    await loadRedeliveryTasks(redeliveryForm.waybillId);
  } finally {
    redeliveryLoading.value = false;
  }
};

const closeRedelivery = async () => {
  if (!redeliveryForm.taskId) return;
  redeliveryLoading.value = true;
  try {
    await logisticsApi.closeRedeliveryTask(redeliveryForm.taskId, {
      operator_id: redeliveryForm.operatorId || undefined,
      reason: redeliveryForm.reason || undefined,
    });
    await loadRedeliveryTasks(redeliveryForm.waybillId);
  } finally {
    redeliveryLoading.value = false;
  }
};

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

const orderSummary = (waybill: Waybill) => {
  const siblings = waybills.value.filter((row) => row.orderNo === waybill.orderNo);
  const aggregateStatus = siblings.some((row) => row.orderFulfillmentStatus === "fully_shipped")
    ? "fully_shipped"
    : "partial_shipped";
  return {
    packageCount: siblings.length,
    aggregateStatus,
  };
};

const orderStatusMeta = (status: string) => {
  if (status === "fully_shipped") {
    return { label: "全部发货", color: "success" as const };
  }
  return { label: "部分发货", color: "warning" as const };
};
</script>
