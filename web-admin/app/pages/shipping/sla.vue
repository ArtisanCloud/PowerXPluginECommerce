<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">承运商 SLA 看板</h1>
        <p class="text-gray-500 dark:text-gray-400">监控揽收时效、签收时效与异常率。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="carrierId" class="w-44" placeholder="承运商ID（可选）" />
        <UInput v-model.number="pickupSlaHours" class="w-36" type="number" placeholder="揽收 SLA(h)" />
        <UInput v-model.number="deliverySlaHours" class="w-36" type="number" placeholder="签收 SLA(h)" />
        <UButton color="info" variant="soft" icon="i-heroicons-bolt" :loading="syncJobLoading" @click="createSyncJob">
          创建批量同步任务
        </UButton>
        <UButton color="primary" icon="i-heroicons-arrow-path" @click="loadSnapshot">刷新</UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">网关健康</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">请求总数</p>
          <p class="mt-1 text-xl font-semibold">{{ gateway.summary.totalRequests }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">成功率</p>
          <p class="mt-1 text-xl font-semibold">{{ gateway.summary.successRate.toFixed(2) }}%</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">P95 延迟</p>
          <p class="mt-1 text-xl font-semibold">{{ gateway.summary.p95LatencyMS }} ms</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">失败请求</p>
          <p class="mt-1 text-xl font-semibold text-warning-600">{{ gateway.summary.failedRequests }}</p>
        </div>
      </div>
      <div class="mt-3 space-y-2">
        <p class="text-sm font-medium text-gray-900 dark:text-white">告警</p>
        <div v-if="!gateway.alerts.length" class="text-xs text-gray-500">暂无告警</div>
        <div v-for="item in gateway.alerts" :key="item.code" class="rounded border border-gray-200 p-2 text-xs dark:border-gray-800">
          <span class="font-semibold">{{ item.code }}</span> · {{ item.message }}
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">成本与配额</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">调用量</p>
          <p class="mt-1 text-xl font-semibold">{{ gatewayCost.summary.totalRequests }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">累计成本</p>
          <p class="mt-1 text-xl font-semibold">¥{{ gatewayCost.summary.totalCost.toFixed(2) }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">配额使用</p>
          <p class="mt-1 text-xl font-semibold">{{ gatewayCost.summary.quotaUsed }} / {{ gatewayCost.summary.quotaLimit }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">配额使用率</p>
          <p class="mt-1 text-xl font-semibold" :class="gatewayCost.summary.quotaUsageRate >= 100 ? 'text-error-600' : ''">
            {{ gatewayCost.summary.quotaUsageRate.toFixed(2) }}%
          </p>
        </div>
      </div>
      <div class="mt-3 space-y-2">
        <p class="text-sm font-medium text-gray-900 dark:text-white">配额告警</p>
        <div v-if="!gatewayCost.alerts.length" class="text-xs text-gray-500">暂无告警</div>
        <div
          v-for="item in gatewayCost.alerts"
          :key="`${item.code}-${item.carrierId}-${item.provider}`"
          class="rounded border border-gray-200 p-2 text-xs dark:border-gray-800"
        >
          <span class="font-semibold">{{ item.code }}</span> · {{ item.message }}
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">同步计划管理</h3>
      </template>
      <div class="grid gap-2 md:grid-cols-6">
        <UInput v-model="scheduleForm.name" placeholder="计划名称" />
        <UInput v-model="scheduleForm.cronExpr" placeholder="cron（如 */15 * * * *）" />
        <UInput v-model="scheduleForm.carrierId" placeholder="承运商ID（可选）" />
        <USelect v-model="scheduleForm.waybillStatus" :options="scheduleStatusOptions" />
        <UInput v-model.number="scheduleForm.batchLimit" type="number" placeholder="批次" />
        <UInput v-model.number="scheduleForm.eventLimit" type="number" placeholder="单运单条数" />
      </div>
      <div class="mt-2">
        <UButton color="primary" size="sm" :loading="scheduleLoading" @click="createSchedule">
          创建计划
        </UButton>
      </div>
      <ul class="mt-3 space-y-2 text-xs">
        <li
          v-for="row in schedules"
          :key="row.id"
          class="rounded border border-gray-200 p-2 dark:border-gray-800"
        >
          <div class="flex items-center justify-between gap-2">
            <span>{{ row.name }} · {{ row.cronExpr }} · {{ row.enabled ? "启用" : "停用" }}</span>
            <div class="flex gap-2">
              <UButton size="xs" variant="ghost" :loading="scheduleLoading" @click="triggerSchedule(row.id)">立即执行</UButton>
              <UButton size="xs" color="warning" variant="ghost" :loading="scheduleLoading" @click="toggleSchedule(row.id, !row.enabled)">
                {{ row.enabled ? "停用" : "启用" }}
              </UButton>
            </div>
          </div>
        </li>
        <li v-if="!schedules.length" class="text-gray-500">暂无计划</li>
      </ul>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">SLO 守卫与自动限流</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">策略总数</p>
          <p class="mt-1 text-xl font-semibold">{{ sloStatus.totalPolicies }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">启用策略</p>
          <p class="mt-1 text-xl font-semibold">{{ sloStatus.enabledPolicies }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">触发限流</p>
          <p class="mt-1 text-xl font-semibold" :class="sloStatus.throttledPolicies > 0 ? 'text-error-600' : ''">
            {{ sloStatus.throttledPolicies }}
          </p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">守卫状态</p>
          <p class="mt-1 text-xl font-semibold" :class="sloStatus.throttleRequired ? 'text-error-600' : 'text-success-600'">
            {{ sloStatus.throttleRequired ? "限流中" : "正常" }}
          </p>
        </div>
      </div>
      <div class="mt-3 grid gap-2 md:grid-cols-8">
        <UInput v-model="sloForm.name" placeholder="策略名称" />
        <UInput v-model="sloForm.carrierId" placeholder="承运商ID（可选）" />
        <UInput v-model.number="sloForm.windowHours" type="number" placeholder="窗口(h)" />
        <UInput v-model.number="sloForm.minSuccessRate" type="number" placeholder="最小成功率%" />
        <UInput v-model.number="sloForm.maxP95LatencyMS" type="number" placeholder="最大P95(ms)" />
        <UInput v-model.number="sloForm.maxFailedRequests" type="number" placeholder="最大失败数" />
        <UInput v-model.number="sloForm.throttleRatio" type="number" placeholder="限流比例%" />
        <USelect v-model="sloForm.action" :options="sloActionOptions" />
      </div>
      <div class="mt-2 flex flex-wrap gap-2">
        <UButton size="sm" color="primary" :loading="sloLoading" @click="saveSLOPolicy">保存策略</UButton>
        <UButton size="sm" color="warning" variant="soft" :loading="sloLoading" @click="evaluateSLOGuard">
          立即评估
        </UButton>
      </div>
      <ul class="mt-3 space-y-2 text-xs">
        <li v-for="item in sloStatus.triggered" :key="item.policyId" class="rounded border border-error-200 p-2 dark:border-error-800">
          <div class="flex items-center justify-between gap-2">
            <span>
              <b>{{ item.policyName }}</b> · {{ item.reasonMessage }} · 限流 {{ item.throttleRatio }}%
            </span>
            <UButton size="xs" variant="ghost" :loading="sloLoading" @click="releaseSLOPolicy(item.policyId)">
              手动解除
            </UButton>
          </div>
        </li>
        <li v-if="!sloStatus.triggered.length" class="text-gray-500">当前无触发限流策略</li>
      </ul>
      <UTable class="mt-3" :columns="sloColumns" :data="sloPolicies" />
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">根因分析</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">异常总数</p>
          <p class="mt-1 text-xl font-semibold">{{ rootCauseSummary.totalCases }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">待处理</p>
          <p class="mt-1 text-xl font-semibold text-warning-600">{{ rootCauseSummary.openCases }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">已闭环</p>
          <p class="mt-1 text-xl font-semibold text-success-600">{{ rootCauseSummary.resolvedCases }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">分析窗口</p>
          <p class="mt-1 text-xl font-semibold">{{ rootCauseSummary.windowHours }}h</p>
        </div>
      </div>
      <div class="mt-3 flex flex-wrap gap-2">
        <UInput v-model.number="rootCauseWindowHours" class="w-36" type="number" placeholder="窗口(h)" />
        <UButton size="sm" color="primary" :loading="rootCauseLoading" @click="analyzeRootCauses">
          执行分析
        </UButton>
      </div>
      <div class="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2">
        <div class="rounded border border-gray-200 p-3 text-xs dark:border-gray-800">
          <p class="mb-2 text-sm font-medium text-gray-900 dark:text-white">异常分布</p>
          <div v-if="!rootCauseSummary.anomalyDistribution.length" class="text-gray-500">暂无数据</div>
          <div
            v-for="item in rootCauseSummary.anomalyDistribution"
            :key="`anomaly-${item.key}`"
            class="mb-1 flex items-center justify-between"
          >
            <span>{{ item.key }}</span>
            <span>{{ item.count }}</span>
          </div>
        </div>
        <div class="rounded border border-gray-200 p-3 text-xs dark:border-gray-800">
          <p class="mb-2 text-sm font-medium text-gray-900 dark:text-white">责任占比</p>
          <div v-if="!rootCauseSummary.ownerDistribution.length" class="text-gray-500">暂无数据</div>
          <div
            v-for="item in rootCauseSummary.ownerDistribution"
            :key="`owner-${item.key}`"
            class="mb-1 flex items-center justify-between"
          >
            <span>{{ item.key }}</span>
            <span>{{ item.count }}</span>
          </div>
        </div>
      </div>
      <div class="mt-3 rounded border border-gray-200 p-3 text-xs dark:border-gray-800">
        <p class="mb-2 text-sm font-medium text-gray-900 dark:text-white">建议动作</p>
        <div class="flex flex-wrap gap-2">
          <UBadge v-for="item in rootCauseSummary.suggestedActions" :key="item" color="info" variant="soft">{{ item }}</UBadge>
          <span v-if="!rootCauseSummary.suggestedActions.length" class="text-gray-500">暂无建议动作</span>
        </div>
      </div>
      <UTable class="mt-3" :columns="rootCauseColumns" :data="rootCauseItems">
        <template #status-cell="{ row }">
          <UBadge :color="row.original.status === 'resolved' ? 'success' : 'warning'" variant="soft">
            {{ row.original.status }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <UButton
            size="xs"
            color="primary"
            :loading="rootCauseLoading"
            :disabled="row.original.status === 'resolved'"
            @click="resolveRootCause(row.original.id)"
          >
            标记闭环
          </UButton>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">履约自动化运维中心</h3>
          <div class="flex flex-wrap gap-2">
            <UButton size="sm" color="warning" variant="soft" :loading="opsLoading" @click="evaluateOpsCenter">立即评估</UButton>
            <UButton size="sm" color="neutral" variant="ghost" :loading="opsLoading" @click="loadSnapshot">刷新</UButton>
          </div>
        </div>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">告警抑制命中</p>
          <p class="mt-1 text-xl font-semibold">{{ opsSnapshot.suppressionHits }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">自动恢复成功</p>
          <p class="mt-1 text-xl font-semibold text-success-600">{{ opsSnapshot.autoRecovered }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">升级链路触发</p>
          <p class="mt-1 text-xl font-semibold text-warning-600">{{ opsSnapshot.escalated }}</p>
        </div>
      </div>
      <div class="mt-3 grid grid-cols-1 gap-2 md:grid-cols-5">
        <UInput v-model="opsForm.name" placeholder="策略名称" />
        <UInput v-model="opsForm.carrierId" placeholder="承运商ID（可选）" />
        <UInput v-model.number="opsForm.minFailedEvents" type="number" placeholder="抑制阈值" />
        <UInput v-model.number="opsForm.maxAutoRecover" type="number" placeholder="自动恢复次数" />
        <UInput v-model.number="opsForm.escalationThreshold" type="number" placeholder="升级阈值" />
      </div>
      <div class="mt-2">
        <UButton size="sm" color="primary" :loading="opsLoading" @click="saveOpsPolicy">保存策略</UButton>
      </div>
      <UTable class="mt-3" :columns="opsPolicyColumns" :data="opsPolicies" />
      <UTable class="mt-3" :columns="opsRunColumns" :data="opsRuns">
        <template #suppressed-cell="{ getValue }">
          <UBadge :color="getValue() ? 'info' : 'neutral'" variant="subtle">{{ getValue() ? "是" : "否" }}</UBadge>
        </template>
        <template #autoRecovered-cell="{ getValue }">
          <UBadge :color="getValue() ? 'success' : 'neutral'" variant="subtle">{{ getValue() ? "是" : "否" }}</UBadge>
        </template>
        <template #escalated-cell="{ getValue }">
          <UBadge :color="getValue() ? 'warning' : 'neutral'" variant="subtle">{{ getValue() ? "是" : "否" }}</UBadge>
        </template>
        <template #actions-cell="{ row }">
          <UButton size="xs" variant="ghost" color="warning" :loading="opsLoading" @click="takeoverOpsRun(row.original.id)">
            手动接管
          </UButton>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">总体指标</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">运单总数</p>
          <p class="mt-1 text-xl font-semibold">{{ total.waybillCount }}</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">揽收准时率</p>
          <p class="mt-1 text-xl font-semibold">{{ total.pickupOnTimeRate.toFixed(2) }}%</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">签收准时率</p>
          <p class="mt-1 text-xl font-semibold">{{ total.signOnTimeRate.toFixed(2) }}%</p>
        </div>
        <div class="rounded-lg border border-gray-200 p-4 dark:border-gray-800">
          <p class="text-xs text-gray-500">异常率</p>
          <p class="mt-1 text-xl font-semibold text-warning-600">{{ total.exceptionRate.toFixed(2) }}%</p>
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">承运商分布</h3>
      </template>
      <UTable :columns="columns" :data="rows">
        <template #pickupOnTimeRate-cell="{ getValue }">
          <span>{{ Number(getValue()).toFixed(2) }}%</span>
        </template>
        <template #signOnTimeRate-cell="{ getValue }">
          <span>{{ Number(getValue()).toFixed(2) }}%</span>
        </template>
        <template #exceptionRate-cell="{ getValue }">
          <span class="text-warning-600">{{ Number(getValue()).toFixed(2) }}%</span>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import {
  useLogisticsApi,
  type LogisticsGatewayCostSnapshot,
  type LogisticsGatewayHealthSnapshot,
  type LogisticsOpsAutomationPolicy,
  type LogisticsOpsAutomationRun,
  type LogisticsOpsAutomationSnapshot,
  type LogisticsTrackingRootCause,
  type LogisticsTrackingRootCauseSummary,
  type LogisticsSLOGuardPolicy,
  type LogisticsSLOGuardStatusSnapshot,
  type LogisticsSLASummaryItem,
  type LogisticsTrackingSyncSchedule,
} from "~/composables/api";

definePageMeta({
  name: "shipping-sla",
});

const logisticsApi = useLogisticsApi();
const carrierId = ref("");
const pickupSlaHours = ref(24);
const deliverySlaHours = ref(72);
const rows = ref<LogisticsSLASummaryItem[]>([]);
const syncJobLoading = ref(false);
const scheduleLoading = ref(false);
const sloLoading = ref(false);
const opsLoading = ref(false);
const rootCauseLoading = ref(false);
const rootCauseWindowHours = ref(24);
const schedules = ref<LogisticsTrackingSyncSchedule[]>([]);
const sloPolicies = ref<LogisticsSLOGuardPolicy[]>([]);
const opsPolicies = ref<LogisticsOpsAutomationPolicy[]>([]);
const opsRuns = ref<LogisticsOpsAutomationRun[]>([]);
const opsSnapshot = ref<LogisticsOpsAutomationSnapshot>({
  suppressionHits: 0,
  autoRecovered: 0,
  escalated: 0,
  runs: [],
});
const rootCauseItems = ref<LogisticsTrackingRootCause[]>([]);
const rootCauseSummary = ref<LogisticsTrackingRootCauseSummary>({
  windowHours: 24,
  totalCases: 0,
  openCases: 0,
  resolvedCases: 0,
  anomalyDistribution: [],
  ownerDistribution: [],
  suggestedActions: [],
});
const sloStatus = ref<LogisticsSLOGuardStatusSnapshot>({
  windowHours: 24,
  carrierId: "",
  totalPolicies: 0,
  enabledPolicies: 0,
  throttledPolicies: 0,
  throttleRequired: false,
  gateway: {
    windowHours: 24,
    totalRequests: 0,
    successRequests: 0,
    failedRequests: 0,
    successRate: 0,
    p95LatencyMS: 0,
  },
  triggered: [],
});
const gateway = ref<LogisticsGatewayHealthSnapshot>({
  summary: {
    windowHours: 24,
    totalRequests: 0,
    successRequests: 0,
    failedRequests: 0,
    successRate: 0,
    p95LatencyMS: 0,
  },
  carriers: [],
  alerts: [],
});
const gatewayCost = ref<LogisticsGatewayCostSnapshot>({
  summary: {
    windowHours: 24,
    totalRequests: 0,
    successCount: 0,
    failedCount: 0,
    totalCost: 0,
    quotaLimit: 5000,
    quotaUsed: 0,
    quotaUsageRate: 0,
  },
  carriers: [],
  alerts: [],
});
const scheduleForm = reactive({
  name: "",
  cronExpr: "*/15 * * * *",
  carrierId: "",
  waybillStatus: "in_transit",
  batchLimit: 20,
  eventLimit: 20,
});
const scheduleStatusOptions = [
  { label: "运输中", value: "in_transit" },
  { label: "待揽收", value: "created" },
  { label: "延误", value: "delay" },
  { label: "全部状态", value: "" },
];
const sloActionOptions = [
  { label: "限流", value: "throttle" },
  { label: "拒绝", value: "reject" },
];
const sloForm = reactive({
  name: "",
  carrierId: "",
  windowHours: 24,
  minSuccessRate: 95,
  maxP95LatencyMS: 2000,
  maxFailedRequests: 10,
  throttleRatio: 50,
  action: "throttle",
});
const opsForm = reactive({
  name: "",
  carrierId: "",
  minFailedEvents: 2,
  maxAutoRecover: 1,
  escalationThreshold: 3,
});
const total = ref<LogisticsSLASummaryItem>({
  carrierId: "all",
  carrierName: "全部承运商",
  waybillCount: 0,
  pickupOnTimeCount: 0,
  pickupOnTimeRate: 0,
  signOnTimeCount: 0,
  signOnTimeRate: 0,
  exceptionCount: 0,
  exceptionRate: 0,
  pickupSLAHours: 24,
  deliverySLAHours: 72,
});

const columns = computed<TableColumn<LogisticsSLASummaryItem>[]>(() => [
  { accessorKey: "carrierName", header: "承运商" },
  { accessorKey: "waybillCount", header: "运单数" },
  { accessorKey: "pickupOnTimeCount", header: "揽收准时单" },
  { accessorKey: "pickupOnTimeRate", header: "揽收准时率" },
  { accessorKey: "signOnTimeCount", header: "签收准时单" },
  { accessorKey: "signOnTimeRate", header: "签收准时率" },
  { accessorKey: "exceptionCount", header: "异常单数" },
  { accessorKey: "exceptionRate", header: "异常率" },
]);

const sloColumns = computed<TableColumn<LogisticsSLOGuardPolicy>[]>(() => [
  { accessorKey: "name", header: "策略" },
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "windowHours", header: "窗口(h)" },
  { accessorKey: "minSuccessRate", header: "成功率阈值%" },
  { accessorKey: "maxP95LatencyMS", header: "P95阈值(ms)" },
  { accessorKey: "maxFailedRequests", header: "失败阈值" },
  { accessorKey: "throttleRatio", header: "限流比例%" },
  { accessorKey: "enabled", header: "启用" },
]);

const opsPolicyColumns = computed<TableColumn<LogisticsOpsAutomationPolicy>[]>(() => [
  { accessorKey: "name", header: "策略" },
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "lastEvaluatedAt", header: "最近评估" },
  { accessorKey: "enabled", header: "启用" },
]);

const opsRunColumns = computed<TableColumn<LogisticsOpsAutomationRun>[]>(() => [
  { accessorKey: "triggerSource", header: "触发源" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "suppressed", header: "抑制" },
  { accessorKey: "autoRecovered", header: "自动恢复" },
  { accessorKey: "escalated", header: "升级" },
  { accessorKey: "takeoverBy", header: "接管人" },
  { accessorKey: "createdAt", header: "时间" },
  { id: "actions", header: "操作" },
]);

const rootCauseColumns = computed<TableColumn<LogisticsTrackingRootCause>[]>(() => [
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "anomalyType", header: "异常类型" },
  { accessorKey: "ownerType", header: "责任归因" },
  { accessorKey: "severity", header: "严重级别" },
  { accessorKey: "suggestedAction", header: "建议动作" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "处理" },
]);

const loadSnapshot = async () => {
  const snapshot = await logisticsApi.getSLADashboard({
    carrier_id: carrierId.value || undefined,
    pickup_sla_hours: pickupSlaHours.value || undefined,
    delivery_sla_hours: deliverySlaHours.value || undefined,
  });
  rows.value = snapshot.summary || [];
  total.value = snapshot.total || total.value;
  gateway.value = await logisticsApi.getGatewayHealth({ window_hours: 24 });
  gatewayCost.value = await logisticsApi.getGatewayCosts({ window_hours: 24, quota_limit: 5000 });
  schedules.value = await logisticsApi.listTrackingSyncSchedules({ limit: 20 });
  sloPolicies.value = await logisticsApi.listSLOGuardPolicies({ carrier_id: carrierId.value || undefined, limit: 20 });
  sloStatus.value = await logisticsApi.getSLOGuardStatus({ carrier_id: carrierId.value || undefined, window_hours: 24 });
  opsPolicies.value = await logisticsApi.listOpsAutomationPolicies({ carrier_id: carrierId.value || undefined, limit: 20 });
  opsRuns.value = await logisticsApi.listOpsAutomationRuns({ carrier_id: carrierId.value || undefined, limit: 20 });
  opsSnapshot.value = {
    suppressionHits: opsRuns.value.filter((item) => item.suppressed).length,
    autoRecovered: opsRuns.value.filter((item) => item.autoRecovered).length,
    escalated: opsRuns.value.filter((item) => item.escalated).length,
    runs: opsRuns.value,
  };
  rootCauseSummary.value = await logisticsApi.getTrackingRootCauseSummary({
    carrier_id: carrierId.value || undefined,
    window_hours: rootCauseWindowHours.value || 24,
    limit: 50,
  });
  rootCauseItems.value = await logisticsApi.listTrackingRootCauses({
    carrier_id: carrierId.value || undefined,
    window_hours: rootCauseWindowHours.value || 24,
    limit: 50,
  });
};

const analyzeRootCauses = async () => {
  rootCauseLoading.value = true;
  try {
    rootCauseItems.value = await logisticsApi.analyzeTrackingRootCauses({
      carrier_id: carrierId.value || undefined,
      window_hours: rootCauseWindowHours.value || 24,
      limit: 50,
    });
    rootCauseSummary.value = await logisticsApi.getTrackingRootCauseSummary({
      carrier_id: carrierId.value || undefined,
      window_hours: rootCauseWindowHours.value || 24,
      limit: 50,
    });
  } finally {
    rootCauseLoading.value = false;
  }
};

const resolveRootCause = async (id: string) => {
  rootCauseLoading.value = true;
  try {
    await logisticsApi.handleTrackingRootCause(id, {
      action: "manual_review",
      status: "resolved",
      operator_id: "admin",
      result_note: "resolved from sla dashboard",
    });
    await loadSnapshot();
  } finally {
    rootCauseLoading.value = false;
  }
};

const createSyncJob = async () => {
  syncJobLoading.value = true;
  try {
    await logisticsApi.createTrackingSyncJob({
      carrier_id: carrierId.value || undefined,
      waybill_status: "in_transit",
      batch_limit: 20,
      event_limit: 20,
    });
    await loadSnapshot();
  } finally {
    syncJobLoading.value = false;
  }
};

const createSchedule = async () => {
  scheduleLoading.value = true;
  try {
    await logisticsApi.upsertTrackingSyncSchedule({
      name: scheduleForm.name || "默认同步计划",
      cron_expr: scheduleForm.cronExpr || "*/15 * * * *",
      carrier_id: scheduleForm.carrierId || undefined,
      waybill_status: scheduleForm.waybillStatus || undefined,
      enabled: true,
      batch_limit: Number(scheduleForm.batchLimit || 20),
      event_limit: Number(scheduleForm.eventLimit || 20),
    });
    await loadSnapshot();
  } finally {
    scheduleLoading.value = false;
  }
};

const toggleSchedule = async (id: string, enabled: boolean) => {
  scheduleLoading.value = true;
  try {
    await logisticsApi.toggleTrackingSyncSchedule(id, enabled);
    await loadSnapshot();
  } finally {
    scheduleLoading.value = false;
  }
};

const triggerSchedule = async (id: string) => {
  scheduleLoading.value = true;
  try {
    await logisticsApi.triggerTrackingSyncSchedule(id);
    await loadSnapshot();
  } finally {
    scheduleLoading.value = false;
  }
};

const saveSLOPolicy = async () => {
  sloLoading.value = true;
  try {
    await logisticsApi.upsertSLOGuardPolicy({
      name: sloForm.name || "默认SLO守卫",
      carrier_id: sloForm.carrierId || undefined,
      window_hours: Number(sloForm.windowHours || 24),
      min_success_rate: Number(sloForm.minSuccessRate || 95),
      max_p95_latency_ms: Number(sloForm.maxP95LatencyMS || 2000),
      max_failed_requests: Number(sloForm.maxFailedRequests || 10),
      throttle_ratio: Number(sloForm.throttleRatio || 50),
      action: sloForm.action || "throttle",
      enabled: true,
    });
    await loadSnapshot();
  } finally {
    sloLoading.value = false;
  }
};

const evaluateSLOGuard = async () => {
  sloLoading.value = true;
  try {
    sloStatus.value = await logisticsApi.evaluateSLOGuard({
      carrier_id: carrierId.value || undefined,
      window_hours: 24,
    });
    await loadSnapshot();
  } finally {
    sloLoading.value = false;
  }
};

const releaseSLOPolicy = async (policyId: string) => {
  sloLoading.value = true;
  try {
    await logisticsApi.releaseSLOGuardPolicy(policyId, {
      operator_id: "admin",
      reason: "manual release from sla dashboard",
    });
    await loadSnapshot();
  } finally {
    sloLoading.value = false;
  }
};

const saveOpsPolicy = async () => {
  opsLoading.value = true;
  try {
    await logisticsApi.upsertOpsAutomationPolicy({
      name: opsForm.name || "默认运维策略",
      carrier_id: opsForm.carrierId || undefined,
      retry_strategy: {
        enabled: true,
        max_attempts: Number(opsForm.maxAutoRecover || 1),
      },
      suppression_rule: {
        min_failed_events: Number(opsForm.minFailedEvents || 2),
      },
      escalation_chain: {
        failed_event_threshold: Number(opsForm.escalationThreshold || 3),
        slo_throttled_required: true,
      },
      enabled: true,
    });
    await loadSnapshot();
  } finally {
    opsLoading.value = false;
  }
};

const evaluateOpsCenter = async () => {
  opsLoading.value = true;
  try {
    opsSnapshot.value = await logisticsApi.evaluateOpsAutomation({
      carrier_id: carrierId.value || undefined,
      trigger_source: "manual_evaluate",
      limit: 20,
    });
    opsRuns.value = opsSnapshot.value.runs;
    await loadSnapshot();
  } finally {
    opsLoading.value = false;
  }
};

const takeoverOpsRun = async (id: string) => {
  opsLoading.value = true;
  try {
    await logisticsApi.takeoverOpsAutomationRun(id, {
      action: "manual_takeover",
      operator_id: "admin",
      reason: "manual takeover from sla dashboard",
    });
    await loadSnapshot();
  } finally {
    opsLoading.value = false;
  }
};

onMounted(() => {
  loadSnapshot();
});
</script>
