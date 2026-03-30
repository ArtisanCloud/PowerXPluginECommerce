<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约仿真沙盘</h1>
        <p class="text-gray-500 dark:text-gray-400">在发布策略前，先比较时效、成本、异常率影响。</p>
      </div>
      <div class="flex gap-2">
        <UButton color="primary" :loading="loading" @click="refreshAll">刷新</UButton>
      </div>
    </div>

    <UAlert v-if="message" :color="messageType === 'error' ? 'error' : 'success'" variant="soft" :description="message" />

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">场景配置</h3>
      </template>
      <div class="grid gap-2 md:grid-cols-6">
        <UInput v-model="scenarioForm.name" placeholder="场景名称" />
        <UInput v-model="scenarioForm.carrierId" placeholder="承运商ID" />
        <UInput v-model="scenarioForm.warehouseId" placeholder="仓库ID" />
        <UInput v-model="scenarioForm.destinationZone" placeholder="区域" />
        <UInput v-model.number="scenarioForm.baselineTimelinessRate" type="number" placeholder="时效率%" />
        <UInput v-model.number="scenarioForm.baselineCostIndex" type="number" placeholder="成本指数" />
      </div>
      <div class="mt-2 grid gap-2 md:grid-cols-4">
        <UInput v-model.number="scenarioForm.baselineExceptionRate" type="number" placeholder="异常率%" />
        <UInput v-model.number="scenarioForm.baselineRecoveryHours" type="number" placeholder="恢复时长(h)" />
        <UInput v-model.number="scenarioForm.windowDays" type="number" placeholder="窗口天数" />
        <USelect v-model="runStrategy" :options="strategyOptions" />
      </div>
      <div class="mt-3 flex gap-2">
        <UButton color="info" variant="soft" :loading="saving" @click="saveScenario">保存场景</UButton>
        <UButton color="primary" :loading="running" @click="runScenario">执行仿真</UButton>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">场景列表</h3>
      </template>
      <UTable :columns="scenarioColumns" :data="scenarios" />
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">运行记录</h3>
      </template>
      <div class="grid gap-2 md:grid-cols-2">
        <USelect v-model="compareBaselineRunId" :options="runOptions" placeholder="基线 Run" />
        <USelect v-model="compareCandidateRunId" :options="runOptions" placeholder="候选 Run" />
      </div>
      <div class="mt-2">
        <UButton size="sm" color="warning" variant="soft" :loading="comparing" @click="compareRuns">对比结果</UButton>
      </div>
      <UTable class="mt-3" :columns="runColumns" :data="runs" />
    </UCard>

    <UCard v-if="compareResult">
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">策略对比结论</h3>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-5 text-sm">
        <div class="rounded border border-gray-200 p-3 dark:border-gray-800">时效变化：{{ compareResult.timelinessDelta.toFixed(2) }}</div>
        <div class="rounded border border-gray-200 p-3 dark:border-gray-800">成本变化：{{ compareResult.costDelta.toFixed(2) }}</div>
        <div class="rounded border border-gray-200 p-3 dark:border-gray-800">异常变化：{{ compareResult.exceptionDelta.toFixed(2) }}</div>
        <div class="rounded border border-gray-200 p-3 dark:border-gray-800">恢复变化：{{ compareResult.recoveryHourDelta.toFixed(2) }}</div>
        <div class="rounded border border-gray-200 p-3 dark:border-gray-800">评分变化：{{ compareResult.scoreDelta.toFixed(2) }}</div>
      </div>
      <p class="mt-3 text-sm text-gray-700 dark:text-gray-200">{{ compareResult.recommendation }}</p>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import {
  type LogisticsFulfillmentSandboxCompareResult,
  type LogisticsFulfillmentSandboxRun,
  type LogisticsFulfillmentSandboxScenario,
  useLogisticsApi,
} from "~/composables/api";

definePageMeta({
  name: "shipping-sandbox",
});

const logisticsApi = useLogisticsApi();
const loading = ref(false);
const saving = ref(false);
const running = ref(false);
const comparing = ref(false);
const message = ref("");
const messageType = ref<"success" | "error">("success");
const scenarios = ref<LogisticsFulfillmentSandboxScenario[]>([]);
const runs = ref<LogisticsFulfillmentSandboxRun[]>([]);
const compareResult = ref<LogisticsFulfillmentSandboxCompareResult | null>(null);
const compareBaselineRunId = ref("");
const compareCandidateRunId = ref("");
const runStrategy = ref("balanced");

const strategyOptions = [
  { label: "平衡策略", value: "balanced" },
  { label: "时效优先", value: "timeliness_first" },
  { label: "成本优先", value: "cost_first" },
  { label: "韧性优先", value: "resilience_first" },
];

const scenarioForm = reactive({
  name: "",
  carrierId: "",
  warehouseId: "",
  destinationZone: "",
  baselineTimelinessRate: 93,
  baselineCostIndex: 1,
  baselineExceptionRate: 3.5,
  baselineRecoveryHours: 12,
  windowDays: 7,
});

const scenarioColumns = computed<TableColumn<LogisticsFulfillmentSandboxScenario>[]>(() => [
  { accessorKey: "name", header: "场景" },
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "warehouseId", header: "仓库" },
  { accessorKey: "destinationZone", header: "区域" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "updatedAt", header: "更新时间" },
]);

const runColumns = computed<TableColumn<LogisticsFulfillmentSandboxRun>[]>(() => [
  { accessorKey: "strategy", header: "策略" },
  { accessorKey: "windowDays", header: "窗口(天)" },
  { accessorKey: "timelinessRate", header: "时效率%" },
  { accessorKey: "costIndex", header: "成本指数" },
  { accessorKey: "exceptionRate", header: "异常率%" },
  { accessorKey: "recoveryHours", header: "恢复时长(h)" },
  { accessorKey: "score", header: "评分" },
  { accessorKey: "createdAt", header: "执行时间" },
]);

const runOptions = computed(() =>
  runs.value.map((item) => ({
    label: `${item.strategy} | ${item.score.toFixed(2)} | ${item.createdAt || item.id}`,
    value: item.id,
  })),
);

const resetMessage = () => {
  message.value = "";
};

const showError = (err: unknown) => {
  messageType.value = "error";
  message.value = err instanceof Error ? err.message : "请求失败";
};

const refreshScenarios = async () => {
  scenarios.value = await logisticsApi.listFulfillmentSandboxScenarios({ limit: 50 });
};

const refreshRuns = async () => {
  const scenarioId = scenarios.value[0]?.id;
  runs.value = await logisticsApi.listFulfillmentSandboxRuns({ scenario_id: scenarioId, limit: 50 });
};

const refreshAll = async () => {
  resetMessage();
  loading.value = true;
  try {
    await refreshScenarios();
    await refreshRuns();
  } catch (err) {
    showError(err);
  } finally {
    loading.value = false;
  }
};

const saveScenario = async () => {
  resetMessage();
  saving.value = true;
  try {
    const scenario = await logisticsApi.upsertFulfillmentSandboxScenario({
      name: scenarioForm.name || "默认仿真场景",
      carrier_id: scenarioForm.carrierId || undefined,
      warehouse_id: scenarioForm.warehouseId || undefined,
      destination_zone: scenarioForm.destinationZone || undefined,
      status: "active",
      baseline_config: {
        timeliness_rate: Number(scenarioForm.baselineTimelinessRate || 93),
        cost_index: Number(scenarioForm.baselineCostIndex || 1),
        exception_rate: Number(scenarioForm.baselineExceptionRate || 3.5),
        recovery_hours: Number(scenarioForm.baselineRecoveryHours || 12),
      },
    });
    messageType.value = "success";
    message.value = "场景已保存";
    await refreshScenarios();
    if (!compareBaselineRunId.value && scenario?.id) {
      compareBaselineRunId.value = "";
    }
  } catch (err) {
    showError(err);
  } finally {
    saving.value = false;
  }
};

const runScenario = async () => {
  resetMessage();
  running.value = true;
  try {
    let scenarioId = scenarios.value[0]?.id;
    if (!scenarioId) {
      const scenario = await logisticsApi.upsertFulfillmentSandboxScenario({
        name: scenarioForm.name || "默认仿真场景",
        carrier_id: scenarioForm.carrierId || undefined,
        warehouse_id: scenarioForm.warehouseId || undefined,
        destination_zone: scenarioForm.destinationZone || undefined,
        status: "active",
        baseline_config: {
          timeliness_rate: Number(scenarioForm.baselineTimelinessRate || 93),
          cost_index: Number(scenarioForm.baselineCostIndex || 1),
          exception_rate: Number(scenarioForm.baselineExceptionRate || 3.5),
          recovery_hours: Number(scenarioForm.baselineRecoveryHours || 12),
        },
      });
      scenarioId = scenario.id;
      await refreshScenarios();
    }
    const run = await logisticsApi.runFulfillmentSandbox({
      scenario_id: scenarioId,
      strategy: runStrategy.value || "balanced",
      window_days: Number(scenarioForm.windowDays || 7),
    });
    await refreshRuns();
    if (!compareBaselineRunId.value) {
      compareBaselineRunId.value = run.id;
    } else if (!compareCandidateRunId.value) {
      compareCandidateRunId.value = run.id;
    }
    messageType.value = "success";
    message.value = "仿真已执行";
  } catch (err) {
    showError(err);
  } finally {
    running.value = false;
  }
};

const compareRuns = async () => {
  resetMessage();
  if (!compareBaselineRunId.value || !compareCandidateRunId.value) {
    messageType.value = "error";
    message.value = "请选择基线和候选运行记录";
    return;
  }
  comparing.value = true;
  try {
    compareResult.value = await logisticsApi.compareFulfillmentSandboxRuns({
      baseline_run_id: compareBaselineRunId.value,
      candidate_run_id: compareCandidateRunId.value,
    });
    messageType.value = "success";
    message.value = "对比完成";
  } catch (err) {
    showError(err);
  } finally {
    comparing.value = false;
  }
};

onMounted(() => {
  refreshAll();
});
</script>
