<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约容量预测与动态配额</h1>
        <p class="text-gray-500 dark:text-gray-400">查看预测趋势，确认建议，并一键下发配额。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="filters.carrierId" class="w-40" placeholder="承运商 ID" />
        <UInput v-model="filters.warehouseId" class="w-36" placeholder="仓库 ID" />
        <UInput v-model="filters.destinationZone" class="w-36" placeholder="区域" />
        <USelect v-model="filters.status" class="w-36" :options="statusOptions" />
        <UInput v-model.number="windowDays" class="w-32" type="number" placeholder="窗口(天)" />
        <UButton color="primary" :loading="loading" @click="refreshRows">刷新</UButton>
        <UButton color="info" variant="soft" :loading="generating" @click="generateForecasts">生成预测</UButton>
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
        <p class="text-xs text-gray-500">建议总数</p>
        <p class="mt-1 text-xl font-semibold">{{ rows.length }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">高风险建议</p>
        <p class="mt-1 text-xl font-semibold text-error-600">{{ highRiskCount }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">建议总配额</p>
        <p class="mt-1 text-xl font-semibold">{{ totalRecommendedQuota }}</p>
      </UCard>
      <UCard>
        <p class="text-xs text-gray-500">平均置信度</p>
        <p class="mt-1 text-xl font-semibold">{{ avgConfidence.toFixed(2) }}%</p>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">趋势图（预测量 / 当前容量）</h3>
      </template>
      <div class="space-y-2">
        <div
          v-for="item in trendRows"
          :key="item.id"
          class="rounded border border-gray-200 p-2 text-xs dark:border-gray-800"
        >
          <div class="mb-1 flex items-center justify-between gap-2">
            <span class="font-medium">{{ item.label }}</span>
            <span>{{ item.predictedDailyVolume }} / {{ item.currentDailyCapacity }}</span>
          </div>
          <div class="h-2 rounded bg-gray-100 dark:bg-gray-800">
            <div class="h-2 rounded bg-primary-500" :style="{ width: `${item.ratio}%` }" />
          </div>
        </div>
        <div v-if="!trendRows.length" class="text-xs text-gray-500">暂无趋势数据</div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">建议面板</h3>
      </template>
      <UTable :columns="columns" :data="rows">
        <template #riskLevel-cell="{ row }">
          <UBadge :color="riskColor(row.original.riskLevel)" variant="soft">{{ row.original.riskLevel }}</UBadge>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="statusColor(row.original.status)" variant="soft">{{ row.original.status }}</UBadge>
        </template>
        <template #actions-cell="{ row }">
          <UButton
            size="xs"
            color="primary"
            :disabled="row.original.status === 'applied'"
            :loading="applyingId === row.original.id"
            @click="applyRecommendation(row.original.id)"
          >
            一键下发
          </UButton>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { type LogisticsCapacityForecast, useLogisticsApi } from "~/composables/api";

definePageMeta({
  name: "shipping-capacity-forecast",
});

const logisticsApi = useLogisticsApi();
const loading = ref(false);
const generating = ref(false);
const applyingId = ref("");
const message = ref("");
const messageType = ref<"success" | "error">("success");
const rows = ref<LogisticsCapacityForecast[]>([]);
const windowDays = ref(7);

const filters = reactive({
  carrierId: "",
  warehouseId: "",
  destinationZone: "",
  status: "" as "" | "suggested" | "applied" | "dismissed",
});

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "待确认", value: "suggested" },
  { label: "已下发", value: "applied" },
  { label: "已忽略", value: "dismissed" },
];

const columns = computed<TableColumn<LogisticsCapacityForecast>[]>(() => [
  { accessorKey: "carrierId", header: "承运商" },
  { accessorKey: "warehouseId", header: "仓库" },
  { accessorKey: "destinationZone", header: "区域" },
  { accessorKey: "currentDailyCapacity", header: "当前容量" },
  { accessorKey: "predictedDailyVolume", header: "预测日量" },
  { accessorKey: "targetCapacity", header: "目标容量" },
  { accessorKey: "recommendedQuota", header: "建议配额" },
  { accessorKey: "confidence", header: "置信度(%)" },
  { accessorKey: "riskLevel", header: "风险" },
  { accessorKey: "strategy", header: "策略" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "操作" },
]);

const query = computed(() => ({
  carrier_id: filters.carrierId || undefined,
  warehouse_id: filters.warehouseId || undefined,
  destination_zone: filters.destinationZone || undefined,
  status: filters.status || undefined,
  limit: 100,
}));

const highRiskCount = computed(() => rows.value.filter((item) => item.riskLevel === "high").length);
const totalRecommendedQuota = computed(() => rows.value.reduce((sum, item) => sum + item.recommendedQuota, 0));
const avgConfidence = computed(() => {
  if (!rows.value.length) return 0;
  return rows.value.reduce((sum, item) => sum + item.confidence, 0) / rows.value.length;
});

const trendRows = computed(() =>
  rows.value.map((item) => {
    const current = Math.max(item.currentDailyCapacity, 1);
    const ratio = Math.min((item.predictedDailyVolume / current) * 100, 100);
    return {
      ...item,
      ratio,
      label: `${item.carrierId || "-"} / ${item.warehouseId || "-"} / ${item.destinationZone || "-"}`,
    };
  }),
);

const resetMessage = () => {
  message.value = "";
};

const showError = (err: unknown) => {
  messageType.value = "error";
  message.value = err instanceof Error ? err.message : "请求失败";
};

const riskColor = (risk: string) => {
  if (risk === "high") return "error";
  if (risk === "medium") return "warning";
  return "success";
};

const statusColor = (status: string) => {
  if (status === "applied") return "success";
  if (status === "dismissed") return "neutral";
  return "warning";
};

const refreshRows = async () => {
  resetMessage();
  loading.value = true;
  try {
    rows.value = await logisticsApi.listCapacityForecasts(query.value);
  } catch (err) {
    showError(err);
  } finally {
    loading.value = false;
  }
};

const generateForecasts = async () => {
  resetMessage();
  generating.value = true;
  try {
    const generated = await logisticsApi.generateCapacityForecast({
      carrier_id: filters.carrierId || undefined,
      warehouse_id: filters.warehouseId || undefined,
      destination_zone: filters.destinationZone || undefined,
      window_days: Number(windowDays.value || 7),
    });
    rows.value = generated;
    messageType.value = "success";
    message.value = `已生成 ${generated.length} 条建议`;
  } catch (err) {
    showError(err);
  } finally {
    generating.value = false;
  }
};

const applyRecommendation = async (id: string) => {
  resetMessage();
  applyingId.value = id;
  try {
    const updated = await logisticsApi.applyCapacityForecast(id, {});
    rows.value = rows.value.map((item) => (item.id === updated.id ? updated : item));
    messageType.value = "success";
    message.value = "建议已下发";
  } catch (err) {
    showError(err);
  } finally {
    applyingId.value = "";
  }
};

onMounted(() => {
  refreshRows();
});
</script>
