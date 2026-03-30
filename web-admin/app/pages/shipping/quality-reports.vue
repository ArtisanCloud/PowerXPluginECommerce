<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约质量审计与复盘报告</h1>
        <p class="text-gray-500 dark:text-gray-400">按周期聚合 KPI、根因与跨仓协同数据，输出复盘结论与行动项。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="filters.carrierId" class="w-40" placeholder="承运商 ID" />
        <UInput v-model="filters.warehouseId" class="w-36" placeholder="仓库 ID" />
        <UInput v-model="filters.destinationZone" class="w-36" placeholder="区域" />
        <USelect v-model="filters.status" class="w-36" :options="statusOptions" />
        <UInput v-model.number="filters.windowHours" class="w-32" type="number" placeholder="窗口(小时)" />
        <UButton color="primary" :loading="loading" @click="refreshReports">刷新</UButton>
        <UButton color="info" variant="soft" :loading="generating" @click="generateReport">生成报告</UButton>
        <UButton color="neutral" variant="soft" :loading="exporting" @click="exportReports">导出 CSV</UButton>
      </div>
    </div>

    <UAlert
      v-if="message"
      :color="messageType === 'error' ? 'error' : 'success'"
      variant="soft"
      :title="messageType === 'error' ? '操作失败' : '操作成功'"
      :description="message"
    />

    <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
      <UCard>
        <p class="text-xs text-gray-500">报告总数</p>
        <p class="mt-1 text-xl font-semibold">{{ reports.length }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">平均准时率</p>
        <p class="mt-1 text-xl font-semibold">{{ avgOnTimeRate.toFixed(2) }}%</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">异常总量</p>
        <p class="mt-1 text-xl font-semibold text-warning-600">{{ totalExceptions }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">跨仓协同次数</p>
        <p class="mt-1 text-xl font-semibold">{{ totalInterwarehouse }}</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">复盘报告列表</h3>
      </template>
      <UTable :columns="columns" :data="reports">
        <template #period-cell="{ row }">
          <div class="text-xs">
            <div>{{ formatTime(row.original.reportPeriodFrom) }}</div>
            <div class="text-gray-500">~ {{ formatTime(row.original.reportPeriodTo) }}</div>
          </div>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="statusColor(row.original.status)" variant="soft">{{ row.original.status }}</UBadge>
        </template>
        <template #actions-cell="{ row }">
          <UButton size="xs" variant="ghost" color="primary" @click="viewDetail(row.original.id)">
            查看详情
          </UButton>
        </template>
      </UTable>
    </UCard>

    <UCard v-if="selectedReport">
      <template #header>
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">报告详情</h3>
          <UBadge :color="statusColor(selectedReport.status)" variant="soft">{{ selectedReport.status }}</UBadge>
        </div>
      </template>
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div class="rounded border border-gray-200 p-3 text-xs dark:border-gray-800">
          <p>样本：{{ selectedReport.totalWaybills }}</p>
          <p>签收：{{ selectedReport.deliveredCount }}</p>
          <p>异常：{{ selectedReport.exceptionCount }}</p>
          <p>超时：{{ selectedReport.timeoutCount }}</p>
        </div>
        <div class="rounded border border-gray-200 p-3 text-xs dark:border-gray-800">
          <p>准时率：{{ selectedReport.onTimeRate.toFixed(2) }}%</p>
          <p>成功率：{{ selectedReport.deliverySuccessRate.toFixed(2) }}%</p>
          <p>总成本：{{ selectedReport.totalCost.toFixed(2) }}</p>
          <p>单均成本：{{ selectedReport.avgCost.toFixed(2) }}</p>
        </div>
        <div class="rounded border border-gray-200 p-3 text-xs dark:border-gray-800">
          <p>容量预测：{{ selectedReport.forecastCount }}</p>
          <p>根因案例：{{ selectedReport.rootCauseCount }}</p>
          <p>跨仓协同：{{ selectedReport.interwarehouseCount }}</p>
        </div>
      </div>
      <p class="mt-3 text-sm font-medium text-gray-900 dark:text-white">结论：{{ selectedReport.conclusion || "-" }}</p>
      <div class="mt-2 space-y-2">
        <p class="text-sm font-medium text-gray-900 dark:text-white">行动项</p>
        <ul class="space-y-1 text-xs">
          <li
            v-for="(item, idx) in selectedReport.actionItems"
            :key="`${item.area}-${idx}`"
            class="rounded border border-gray-200 p-2 dark:border-gray-800"
          >
            [{{ item.priority }}] {{ item.area }}：{{ item.action }}
          </li>
          <li v-if="!selectedReport.actionItems.length" class="text-gray-500">暂无行动项</li>
        </ul>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { type LogisticsQualityAuditReport, useLogisticsApi } from "~/composables/api";

definePageMeta({
  name: "shipping-quality-reports",
});

const logisticsApi = useLogisticsApi();
const loading = ref(false);
const generating = ref(false);
const exporting = ref(false);
const message = ref("");
const messageType = ref<"success" | "error">("success");
const reports = ref<LogisticsQualityAuditReport[]>([]);
const selectedReport = ref<LogisticsQualityAuditReport | null>(null);

const filters = reactive({
  carrierId: "",
  warehouseId: "",
  destinationZone: "",
  status: "" as "" | "generated" | "published" | "archived",
  windowHours: 24 * 7,
});

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "已生成", value: "generated" },
  { label: "已发布", value: "published" },
  { label: "已归档", value: "archived" },
];

const columns = computed<TableColumn<LogisticsQualityAuditReport>[]>(() => [
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "warehouseId", header: "仓库" },
  { accessorKey: "destinationZone", header: "区域" },
  { id: "period", header: "报告周期" },
  { accessorKey: "totalWaybills", header: "样本" },
  { accessorKey: "onTimeRate", header: "准时率(%)" },
  { accessorKey: "exceptionCount", header: "异常量" },
  { accessorKey: "interwarehouseCount", header: "跨仓协同" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

const avgOnTimeRate = computed(() => {
  if (!reports.value.length) return 0;
  return reports.value.reduce((sum, item) => sum + item.onTimeRate, 0) / reports.value.length;
});

const totalExceptions = computed(() => reports.value.reduce((sum, item) => sum + item.exceptionCount, 0));
const totalInterwarehouse = computed(() => reports.value.reduce((sum, item) => sum + item.interwarehouseCount, 0));

const statusColor = (status: string) => {
  if (status === "published") return "success";
  if (status === "archived") return "neutral";
  return "warning";
};

const formatTime = (value?: string) => {
  const raw = String(value || "").trim();
  if (!raw) return "-";
  const d = new Date(raw);
  if (Number.isNaN(d.getTime())) return raw;
  return d.toLocaleString("zh-CN", { hour12: false });
};

const reportQuery = computed(() => ({
  carrier_id: filters.carrierId || undefined,
  warehouse_id: filters.warehouseId || undefined,
  destination_zone: filters.destinationZone || undefined,
  status: filters.status || undefined,
  window_hours: Number(filters.windowHours || 24 * 7),
  limit: 100,
}));

const showError = (err: unknown) => {
  messageType.value = "error";
  message.value = err instanceof Error ? err.message : "请求失败";
};

const resetMessage = () => {
  message.value = "";
};

const refreshReports = async () => {
  resetMessage();
  loading.value = true;
  try {
    reports.value = await logisticsApi.listQualityAuditReports(reportQuery.value);
    if (selectedReport.value) {
      const found = reports.value.find((item) => item.id === selectedReport.value?.id);
      if (!found) {
        selectedReport.value = null;
      }
    }
  } catch (err) {
    showError(err);
  } finally {
    loading.value = false;
  }
};

const generateReport = async () => {
  resetMessage();
  generating.value = true;
  try {
    const created = await logisticsApi.generateQualityAuditReport({
      carrier_id: filters.carrierId || undefined,
      warehouse_id: filters.warehouseId || undefined,
      destination_zone: filters.destinationZone || undefined,
      window_hours: Number(filters.windowHours || 24 * 7),
      operator_id: "admin",
    });
    reports.value = [created, ...reports.value.filter((item) => item.id !== created.id)];
    selectedReport.value = created;
    messageType.value = "success";
    message.value = "复盘报告已生成";
  } catch (err) {
    showError(err);
  } finally {
    generating.value = false;
  }
};

const viewDetail = async (id: string) => {
  resetMessage();
  try {
    selectedReport.value = await logisticsApi.getQualityAuditReport(id);
  } catch (err) {
    showError(err);
  }
};

const exportReports = async () => {
  resetMessage();
  exporting.value = true;
  try {
    const result = await logisticsApi.exportQualityAuditReports(reportQuery.value);
    const blob = new Blob([result.content], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `quality-audit-reports-${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    messageType.value = "success";
    message.value = "报告已导出";
  } catch (err) {
    showError(err);
  } finally {
    exporting.value = false;
  }
};

onMounted(() => {
  refreshReports();
});
</script>
