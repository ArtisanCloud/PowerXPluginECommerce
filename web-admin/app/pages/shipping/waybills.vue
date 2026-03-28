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
        <UButton color="info" variant="soft" icon="i-heroicons-bolt" :loading="syncJobCreating" @click="openSyncJobModal">
          批量同步任务
        </UButton>
        <UButton
          color="neutral"
          variant="ghost"
          icon="i-heroicons-arrow-path"
          :loading="syncingAll"
          @click="refreshTracking"
        >
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
            <UButton size="xs" variant="ghost" color="primary" @click="openAllocation(row.original)">
              智能分单
            </UButton>
            <UButton size="xs" variant="ghost" color="neutral" @click="openRoutingPreview(row.original)">
              联合路由仿真
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="openRedelivery(row.original)">
              失败重派
            </UButton>
            <UButton size="xs" variant="ghost" color="error" @click="openFailureCompensation(row.original)">
              失败补偿
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="openOrchestration(row.original)">
              异常编排
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="openLastmileRecovery(row.original)">
              末端自愈
            </UButton>
            <UButton size="xs" variant="ghost" color="info" @click="openAddressValidation(row.original)">
              地址校验
            </UButton>
            <UButton
              size="xs"
              variant="ghost"
              color="info"
              :loading="Boolean(syncingWaybillIDs[row.original.id])"
              @click="syncTracking(row.original)"
            >
              同步轨迹
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

    <UModal v-model:open="routingPreviewOpen" title="联合路由仿真">
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
            执行仿真
          </UButton>
          <UCard v-if="routingPreviewResult">
            <p class="text-sm">策略：{{ routingPreviewResult.strategy }}（{{ routingPreviewResult.reason }}）</p>
            <p class="text-sm">结果承运商：{{ routingPreviewResult.carrierName }}</p>
            <p class="text-xs text-gray-500">{{ routingPreviewResult.explain || "-" }}</p>
            <ul class="mt-2 space-y-1 text-xs">
              <li v-for="item in routingPreviewResult.candidates || []" :key="item.carrierId">
                {{ item.carrierName }} · score={{ Number(item.finalScore || 0).toFixed(2) }}
              </li>
            </ul>
          </UCard>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="allocationOpen" title="智能分单与手工改派">
      <template #body>
        <div class="space-y-3">
          <p class="text-xs text-gray-500">
            运单：{{ allocationForm.waybillNo || "-" }} / 订单：{{ allocationForm.orderId || "-" }}
          </p>
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="allocationForm.warehouseId" placeholder="仓库ID（可选）" />
            <UInput v-model="allocationForm.destinationZone" placeholder="目的区域（可选）" />
            <UInput v-model="allocationForm.preferredCarrier" placeholder="偏好承运商（可选）" />
            <UInput v-model="allocationForm.overrideCarrierId" placeholder="手工改派承运商ID" />
          </div>
          <div class="flex gap-2">
            <UButton color="primary" :loading="allocationLoading" @click="runAllocation">自动分单</UButton>
            <UButton color="warning" :loading="allocationLoading" :disabled="!allocationResult" @click="runOverrideAllocation">
              手工改派
            </UButton>
          </div>
          <UCard v-if="allocationResult">
            <p class="text-sm">策略：{{ allocationResult.strategy }}</p>
            <p class="text-sm">结果：{{ allocationResult.carrierName || allocationResult.carrierId }}</p>
            <p class="text-xs text-gray-500">{{ allocationResult.reason }}</p>
            <ul class="mt-2 space-y-1 text-xs">
              <li v-for="item in allocationResult.candidates || []" :key="item.carrierId">
                {{ item.carrierName }} · available={{ item.available }} · score={{ item.finalScore.toFixed(2) }}
              </li>
            </ul>
          </UCard>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="syncJobOpen" title="创建批量轨迹同步任务">
      <template #body>
        <div class="space-y-3">
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="syncJobForm.carrierId" placeholder="承运商ID（可选）" />
            <USelect v-model="syncJobForm.waybillStatus" :options="syncJobStatusOptions" />
            <UInput v-model.number="syncJobForm.batchLimit" type="number" placeholder="运单批次大小（默认20）" />
            <UInput v-model.number="syncJobForm.eventLimit" type="number" placeholder="单运单拉取条数（默认20）" />
          </div>
          <UButton color="primary" :loading="syncJobCreating" @click="createSyncJob">
            创建并执行
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="failureCompensationOpen" title="网关失败补偿">
      <template #body>
        <div class="space-y-3">
          <p class="text-xs text-gray-500">
            运单：{{ failureWaybillNo || "-" }}
          </p>
          <div class="flex gap-2">
            <UButton size="xs" color="neutral" :loading="failureLoading" @click="loadGatewayFailures">刷新失败列表</UButton>
            <UButton size="xs" color="warning" :loading="failureLoading" @click="ingestGatewayFailures">拉取失败样本</UButton>
          </div>
          <ul class="space-y-2 text-xs">
            <li
              v-for="item in gatewayFailures"
              :key="item.id"
              class="rounded border border-gray-200 p-2 dark:border-gray-800"
            >
              <div class="flex items-center justify-between gap-2">
                <span>{{ item.errorClass }} · {{ item.errorCode }} · retry={{ item.retryCount }}</span>
                <UButton size="xs" color="primary" :loading="failureLoading" @click="compensateFailure(item.id)">
                  补偿重试
                </UButton>
              </div>
              <p class="mt-1 text-gray-500">{{ item.errorMessage || "-" }}</p>
            </li>
            <li v-if="!gatewayFailures.length" class="text-gray-500">暂无失败记录</li>
          </ul>
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

    <UModal v-model:open="orchestrationOpen" title="异常自动编排">
      <template #body>
        <div class="space-y-3">
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="orchestrationForm.ruleName" placeholder="规则名称" />
            <USelect v-model="orchestrationForm.triggerEvent" :options="orchestrationTriggerOptions" />
            <USelect v-model="orchestrationForm.action" :options="orchestrationActionOptions" />
            <UInput v-model.number="orchestrationForm.priority" type="number" placeholder="优先级" />
          </div>
          <div class="flex gap-2">
            <UButton size="xs" color="primary" :loading="orchestrationLoading" @click="createOrchestrationRule">创建规则</UButton>
            <UButton size="xs" color="warning" :loading="orchestrationLoading" @click="executeOrchestration">执行编排</UButton>
          </div>
          <ul class="space-y-2 text-xs">
            <li v-for="item in orchestrationRuns" :key="item.id" class="rounded border border-gray-200 p-2 dark:border-gray-800">
              {{ item.trigger }} · {{ item.result }} · {{ item.message }}
            </li>
            <li v-if="!orchestrationRuns.length" class="text-gray-500">暂无执行记录</li>
          </ul>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="addressValidationOpen" title="地址智能校验">
      <template #body>
        <div class="space-y-3">
          <UInput v-model="addressForm.address" placeholder="输入收货地址" />
          <UButton size="xs" color="primary" :loading="addressLoading" @click="checkAddressValidation">立即校验</UButton>
          <UCard v-if="latestAddressValidation">
            <p class="text-xs">标准化：{{ latestAddressValidation.normalized || "-" }}</p>
            <p class="text-xs">可达性：{{ latestAddressValidation.reachable ? "可达" : "不可达" }}</p>
            <p class="text-xs">风险：{{ latestAddressValidation.riskLevel }}</p>
            <p class="text-xs">建议：{{ latestAddressValidation.suggestion || "-" }}</p>
          </UCard>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="lastmileRecoveryOpen" title="末端异常自愈中心">
      <template #body>
        <div class="space-y-3">
          <p class="text-xs text-gray-500">运单：{{ lastmileForm.waybillNo || "-" }}</p>
          <div class="grid gap-2 sm:grid-cols-2">
            <UInput v-model="lastmileForm.ruleName" placeholder="规则名称" />
            <USelect v-model="lastmileForm.triggerEvent" :options="lastmileTriggerOptions" />
            <USelect v-model="lastmileForm.action" :options="lastmileActionOptions" />
            <UInput v-model.number="lastmileForm.maxRetries" type="number" placeholder="最大重试次数" />
          </div>
          <div class="flex gap-2">
            <UButton size="xs" color="primary" :loading="lastmileLoading" @click="createLastmileRule">创建规则</UButton>
            <UButton size="xs" color="warning" :loading="lastmileLoading" @click="executeLastmile">执行自愈</UButton>
          </div>
          <ul class="space-y-2 text-xs">
            <li v-for="item in lastmileRuns" :key="item.id" class="rounded border border-gray-200 p-2 dark:border-gray-800">
              <div class="flex items-center justify-between gap-2">
                <span>{{ item.triggerEvent }} · {{ item.action }} · {{ item.status }} · retry {{ item.retryCount }}/{{ item.maxRetries }}</span>
                <UButton size="xs" variant="ghost" :loading="lastmileLoading" @click="takeoverLastmile(item.id)">
                  人工接管
                </UButton>
              </div>
              <p class="mt-1 text-gray-500">{{ item.message || "-" }}</p>
            </li>
            <li v-if="!lastmileRuns.length" class="text-gray-500">暂无自愈记录</li>
          </ul>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import type {
  LogisticsAddressValidation,
  LogisticsAllocationResult,
  LogisticsExceptionOrchestrationRun,
  LogisticsExceptionOrchestrationRule,
  LogisticsGatewayFailureEvent,
  LogisticsLastmileRecoveryRule,
  LogisticsLastmileRecoveryRun,
  LogisticsRedeliveryTask,
  LogisticsWaybillETA,
} from "~/composables/api/useLogistics";
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
const toast = useToast();
const waybills = ref<Waybill[]>([]);
const carrierMap = ref<Record<string, string>>({});
const syncingAll = ref(false);
const syncingWaybillIDs = ref<Record<string, boolean>>({});
const syncJobOpen = ref(false);
const syncJobCreating = ref(false);
const failureCompensationOpen = ref(false);
const failureLoading = ref(false);
const failureWaybillNo = ref("");
const gatewayFailures = ref<LogisticsGatewayFailureEvent[]>([]);
const routingPreviewOpen = ref(false);
const routingPreviewLoading = ref(false);
const routingPreviewResult = ref<any>(null);
const allocationOpen = ref(false);
const allocationLoading = ref(false);
const allocationResult = ref<LogisticsAllocationResult | null>(null);
const redeliveryOpen = ref(false);
const redeliveryLoading = ref(false);
const redeliveryTasks = ref<LogisticsRedeliveryTask[]>([]);
const orchestrationOpen = ref(false);
const orchestrationLoading = ref(false);
const orchestrationRules = ref<LogisticsExceptionOrchestrationRule[]>([]);
const orchestrationRuns = ref<LogisticsExceptionOrchestrationRun[]>([]);
const addressValidationOpen = ref(false);
const addressLoading = ref(false);
const latestAddressValidation = ref<LogisticsAddressValidation | null>(null);
const lastmileRecoveryOpen = ref(false);
const lastmileLoading = ref(false);
const lastmileRules = ref<LogisticsLastmileRecoveryRule[]>([]);
const lastmileRuns = ref<LogisticsLastmileRecoveryRun[]>([]);
const routingForm = reactive({
  waybillNo: "",
  warehouseId: "",
  destinationZone: "GLOBAL",
  serviceCode: "std",
  weight: 1,
  preferredCarrierId: "",
  preferredCarrierName: "",
});
const allocationForm = reactive({
  waybillId: "",
  waybillNo: "",
  orderId: "",
  warehouseId: "",
  destinationZone: "",
  preferredCarrier: "",
  overrideCarrierId: "",
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
const orchestrationForm = reactive({
  waybillId: "",
  waybillNo: "",
  ruleName: "延误自动补偿",
  triggerEvent: "delay",
  action: "auto_compensate",
  priority: 100,
});
const addressForm = reactive({
  waybillId: "",
  waybillNo: "",
  address: "",
});
const lastmileForm = reactive({
  waybillId: "",
  waybillNo: "",
  ruleName: "末端异常自愈规则",
  triggerEvent: "delay",
  action: "redispatch",
  maxRetries: 3,
});
const orchestrationTriggerOptions = [
  { label: "延误", value: "delay" },
  { label: "拒收", value: "rejected" },
  { label: "丢件", value: "lost" },
];
const orchestrationActionOptions = [
  { label: "自动补偿", value: "auto_compensate" },
  { label: "创建工单", value: "create_ticket" },
  { label: "SLA升级", value: "escalate" },
];
const lastmileTriggerOptions = [
  { label: "延误", value: "delay" },
  { label: "拒收", value: "rejected" },
  { label: "丢件", value: "lost" },
  { label: "超时", value: "timeout" },
];
const lastmileActionOptions = [
  { label: "改派", value: "redispatch" },
  { label: "补发", value: "reship" },
  { label: "退款", value: "refund" },
  { label: "人工复核", value: "manual_review" },
];
const syncJobForm = reactive({
  carrierId: "",
  waybillStatus: "",
  batchLimit: 20,
  eventLimit: 20,
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

const resolveErrorMessage = (error: any): string => {
  const dataMessage =
    error?.data?.error?.message || error?.data?.message || error?.response?._data?.error?.message;
  if (typeof dataMessage === "string" && dataMessage.trim()) return dataMessage;
  if (typeof error?.message === "string" && error.message.trim()) return error.message;
  return "请求失败，请稍后重试";
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

const syncJobStatusOptions = [
  { label: "全部状态", value: "" },
  { label: "待揽收", value: "created" },
  { label: "运输中", value: "in_transit" },
  { label: "延误", value: "delay" },
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

const updateWaybillStatus = (waybillID: string, nextStatus: string) => {
  const normalized = normalizeStatus(nextStatus);
  waybills.value = waybills.value.map((item) =>
    item.id === waybillID
      ? {
          ...item,
          status: normalized,
          progress: progressByStatus(normalized),
        }
      : item,
  );
  if (selectedWaybill.value?.id === waybillID) {
    selectedWaybill.value = {
      ...selectedWaybill.value,
      status: normalized,
      progress: progressByStatus(normalized),
    };
  }
};

const syncTracking = async (waybill: Waybill) => {
  syncingWaybillIDs.value = {
    ...syncingWaybillIDs.value,
    [waybill.id]: true,
  };
  try {
    const result = await logisticsApi.syncWaybillTracking(waybill.id, { limit: 20 });
    updateWaybillStatus(waybill.id, result.currentStatus || waybill.status);
    if (selectedWaybill.value?.id === waybill.id) {
      await loadWaybillDetail(selectedWaybill.value);
    }
    toast.add({
      title: `轨迹同步完成 · ${waybill.waybillNo}`,
      description: `新增 ${result.appended} 条，重放 ${result.replayed} 条，当前状态 ${result.currentStatus || "-"}`,
      color: "success",
    });
  } catch (error: any) {
    toast.add({
      title: `轨迹同步失败 · ${waybill.waybillNo}`,
      description: resolveErrorMessage(error),
      color: "error",
    });
  } finally {
    syncingWaybillIDs.value = {
      ...syncingWaybillIDs.value,
      [waybill.id]: false,
    };
  }
};

const refreshTracking = async () => {
  const target = selectedWaybill.value || filteredWaybills.value[0];
  if (!target) return;
  syncingAll.value = true;
  try {
    await syncTracking(target);
  } finally {
    syncingAll.value = false;
  }
};

const openSyncJobModal = () => {
  syncJobOpen.value = true;
};

const createSyncJob = async () => {
  syncJobCreating.value = true;
  try {
    const job = await logisticsApi.createTrackingSyncJob({
      carrier_id: syncJobForm.carrierId || undefined,
      waybill_status: syncJobForm.waybillStatus || undefined,
      batch_limit: Number(syncJobForm.batchLimit || 20),
      event_limit: Number(syncJobForm.eventLimit || 20),
    });
    syncJobOpen.value = false;
    await loadWaybills();
    toast.add({
      title: "批量同步任务已执行",
      description: `状态 ${job.status}，成功 ${job.successCount}，失败 ${job.failedCount}`,
      color: job.failedCount > 0 ? "warning" : "success",
    });
  } catch (error: any) {
    toast.add({
      title: "批量同步任务失败",
      description: resolveErrorMessage(error),
      color: "error",
    });
  } finally {
    syncJobCreating.value = false;
  }
};

const openFailureCompensation = async (waybill: Waybill) => {
  failureWaybillNo.value = waybill.waybillNo;
  failureCompensationOpen.value = true;
  await loadGatewayFailures();
};

const loadGatewayFailures = async () => {
  failureLoading.value = true;
  try {
    gatewayFailures.value = await logisticsApi.listGatewayFailures({
      waybill_no: failureWaybillNo.value || undefined,
      limit: 20,
    });
  } finally {
    failureLoading.value = false;
  }
};

const ingestGatewayFailures = async () => {
  failureLoading.value = true;
  try {
    await logisticsApi.ingestGatewayFailures({ hours: 24 });
    await loadGatewayFailures();
  } finally {
    failureLoading.value = false;
  }
};

const compensateFailure = async (id: string) => {
  failureLoading.value = true;
  try {
    await logisticsApi.compensateGatewayFailure(id);
    await loadGatewayFailures();
    await loadWaybills();
    if (selectedWaybill.value) {
      await loadWaybillDetail(selectedWaybill.value);
    }
  } finally {
    failureLoading.value = false;
  }
};

const openOrchestration = async (waybill: Waybill) => {
  orchestrationForm.waybillId = waybill.id;
  orchestrationForm.waybillNo = waybill.waybillNo;
  orchestrationOpen.value = true;
  orchestrationLoading.value = true;
  try {
    orchestrationRules.value = await logisticsApi.listExceptionOrchestrationRules({ enabled: true });
    orchestrationRuns.value = await logisticsApi.listExceptionOrchestrationRuns({
      waybill_no: waybill.waybillNo,
      limit: 20,
    });
  } finally {
    orchestrationLoading.value = false;
  }
};

const createOrchestrationRule = async () => {
  orchestrationLoading.value = true;
  try {
    const row = await logisticsApi.upsertExceptionOrchestrationRule({
      name: orchestrationForm.ruleName || "自动编排规则",
      trigger_event: orchestrationForm.triggerEvent || "delay",
      action: orchestrationForm.action || "auto_compensate",
      priority: Number(orchestrationForm.priority || 100),
      enabled: true,
    });
    orchestrationRules.value = [row, ...orchestrationRules.value.filter((item) => item.id !== row.id)];
  } finally {
    orchestrationLoading.value = false;
  }
};

const executeOrchestration = async () => {
  const rule = orchestrationRules.value[0];
  if (!rule) return;
  orchestrationLoading.value = true;
  try {
    await logisticsApi.executeExceptionOrchestration({
      rule_id: rule.id,
      waybill_id: orchestrationForm.waybillId || undefined,
      waybill_no: orchestrationForm.waybillNo || undefined,
      trigger: orchestrationForm.triggerEvent || undefined,
    });
    orchestrationRuns.value = await logisticsApi.listExceptionOrchestrationRuns({
      waybill_no: orchestrationForm.waybillNo,
      limit: 20,
    });
  } finally {
    orchestrationLoading.value = false;
  }
};

const openLastmileRecovery = async (waybill: Waybill) => {
  lastmileForm.waybillId = waybill.id;
  lastmileForm.waybillNo = waybill.waybillNo;
  lastmileRecoveryOpen.value = true;
  lastmileLoading.value = true;
  try {
    lastmileRules.value = await logisticsApi.listLastmileRecoveryRules({ enabled: true });
    lastmileRuns.value = await logisticsApi.listLastmileRecoveryRuns({
      waybill_no: waybill.waybillNo,
      limit: 20,
    });
  } finally {
    lastmileLoading.value = false;
  }
};

const createLastmileRule = async () => {
  lastmileLoading.value = true;
  try {
    const row = await logisticsApi.upsertLastmileRecoveryRule({
      name: lastmileForm.ruleName || "末端异常自愈规则",
      trigger_event: lastmileForm.triggerEvent || "delay",
      action: lastmileForm.action || "redispatch",
      priority: 100,
      max_retries: Number(lastmileForm.maxRetries || 3),
      enabled: true,
    });
    lastmileRules.value = [row, ...lastmileRules.value.filter((item) => item.id !== row.id)];
  } finally {
    lastmileLoading.value = false;
  }
};

const executeLastmile = async () => {
  const rule = lastmileRules.value[0];
  lastmileLoading.value = true;
  try {
    await logisticsApi.executeLastmileRecovery({
      request_key: `lastmile#${lastmileForm.waybillId}#${Date.now()}`,
      rule_id: rule?.id,
      waybill_id: lastmileForm.waybillId || undefined,
      waybill_no: lastmileForm.waybillNo || undefined,
      trigger_event: lastmileForm.triggerEvent || undefined,
    });
    lastmileRuns.value = await logisticsApi.listLastmileRecoveryRuns({
      waybill_no: lastmileForm.waybillNo,
      limit: 20,
    });
  } finally {
    lastmileLoading.value = false;
  }
};

const takeoverLastmile = async (id: string) => {
  lastmileLoading.value = true;
  try {
    await logisticsApi.takeoverLastmileRecovery(id, {
      action: "takeover",
      operator_id: "admin",
      reason: "manual intervention",
    });
    lastmileRuns.value = await logisticsApi.listLastmileRecoveryRuns({
      waybill_no: lastmileForm.waybillNo,
      limit: 20,
    });
  } finally {
    lastmileLoading.value = false;
  }
};

const openAddressValidation = (waybill: Waybill) => {
  addressForm.waybillId = waybill.id;
  addressForm.waybillNo = waybill.waybillNo;
  addressForm.address = "";
  latestAddressValidation.value = null;
  addressValidationOpen.value = true;
};

const checkAddressValidation = async () => {
  if (!addressForm.address) return;
  addressLoading.value = true;
  try {
    latestAddressValidation.value = await logisticsApi.checkAddressValidation({
      request_key: `${addressForm.waybillNo || "manual"}#${Date.now()}`,
      waybill_id: addressForm.waybillId || undefined,
      waybill_no: addressForm.waybillNo || undefined,
      address: addressForm.address,
    });
  } finally {
    addressLoading.value = false;
  }
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

const openAllocation = (waybill: Waybill) => {
  allocationForm.waybillId = waybill.id;
  allocationForm.waybillNo = waybill.waybillNo;
  allocationForm.orderId = waybill.orderNo;
  allocationForm.preferredCarrier = waybill.carrierId || "";
  allocationForm.overrideCarrierId = "";
  allocationResult.value = null;
  allocationOpen.value = true;
};

const runAllocation = async () => {
  allocationLoading.value = true;
  try {
    allocationResult.value = await logisticsApi.allocateCarrier({
      request_key: `waybill-alloc#${allocationForm.waybillId}#${Date.now()}`,
      waybill_id: allocationForm.waybillId || undefined,
      order_id: allocationForm.orderId || undefined,
      warehouse_id: allocationForm.warehouseId || undefined,
      destination_zone: allocationForm.destinationZone || undefined,
      preferred_carrier: allocationForm.preferredCarrier || undefined,
      strategy: "capacity_first",
      operator_id: "admin",
    });
  } finally {
    allocationLoading.value = false;
  }
};

const runOverrideAllocation = async () => {
  if (!allocationResult.value) return;
  allocationLoading.value = true;
  try {
    allocationResult.value = await logisticsApi.overrideAllocation({
      decision_id: allocationResult.value.decisionID,
      carrier_id: allocationForm.overrideCarrierId || allocationForm.preferredCarrier,
      reason: "manual reassignment",
      operator_id: "admin",
    });
  } finally {
    allocationLoading.value = false;
  }
};

const previewRouting = async () => {
  routingPreviewLoading.value = true;
  try {
    routingPreviewResult.value = await logisticsApi.simulateRoutingOptimizer({
      request_key: `waybill-sim#${Date.now()}`,
      warehouse_id: routingForm.warehouseId || undefined,
      destination_zone: routingForm.destinationZone || undefined,
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
