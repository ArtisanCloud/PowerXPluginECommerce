<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">履约成本对账</h1>
        <p class="text-gray-500 dark:text-gray-400">按承运商聚合预估/实际费用并查看差异明细。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="carrierId" class="w-44" placeholder="承运商ID（可选）" />
        <UButton color="neutral" variant="ghost" icon="i-heroicons-arrow-path" @click="loadSnapshot">
          刷新
        </UButton>
        <UButton color="primary" icon="i-heroicons-arrow-down-tray" @click="exportCsv">
          导出 CSV
        </UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">承运商汇总</h3>
      </template>
      <UTable :columns="summaryColumns" :data="summaryRows" />
    </UCard>

    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">运单差异明细</h3>
      </template>
      <UTable :columns="itemColumns" :data="itemRows">
        <template #feeDiffAmount-cell="{ getValue }">
          <span :class="Number(getValue()) === 0 ? 'text-gray-600' : 'text-warning-600 font-semibold'">
            {{ Number(getValue()).toFixed(2) }}
          </span>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">对账异常工单</h3>
          <div class="flex gap-2">
            <UInput v-model="newCase.waybillId" class="w-56" placeholder="运单ID" />
            <UInput v-model="newCase.reason" class="w-64" placeholder="异常原因（可选）" />
            <UButton color="primary" variant="soft" @click="createCase">创建工单</UButton>
          </div>
        </div>
      </template>
      <UTable :columns="caseColumns" :data="caseRows">
        <template #status-cell="{ getValue }">
          <UBadge :color="caseStatusMeta(getValue()).color" variant="subtle">
            {{ caseStatusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" color="info" @click="transitionCase(row.original.id, 'confirm')">
              确认
            </UButton>
            <UButton size="xs" variant="ghost" color="warning" @click="transitionCase(row.original.id, 'appeal')">
              申诉
            </UButton>
            <UButton size="xs" variant="ghost" color="success" @click="transitionCase(row.original.id, 'writeoff')">
              核销
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard>
      <template #header>
        <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">结算批次与归因建议</h3>
          <div class="flex gap-2">
            <UButton color="warning" variant="soft" @click="createSettlementBatch">创建结算批次</UButton>
            <UButton color="neutral" variant="ghost" @click="loadSettlementData">刷新批次</UButton>
          </div>
        </div>
      </template>
      <UTable :columns="settlementBatchColumns" :data="settlementBatches">
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" color="info" @click="selectSettlementBatch(row.original.id)">查看差异</UButton>
            <UButton size="xs" variant="ghost" color="success" @click="confirmSettlementBatch(row.original.id)">确认批次</UButton>
          </div>
        </template>
      </UTable>
      <UCard class="mt-3" v-if="settlementDiffs.length">
        <template #header>
          <h4 class="text-sm font-semibold">归因建议（批次差异）</h4>
        </template>
        <UTable :columns="settlementDiffColumns" :data="settlementDiffs">
          <template #actions-cell="{ row }">
            <div class="flex gap-2">
              <UButton size="xs" variant="ghost" color="success" @click="handleSettlementDiff(row.original.id, 'accept')">接受</UButton>
              <UButton size="xs" variant="ghost" color="warning" @click="handleSettlementDiff(row.original.id, 'dispute')">发起争议</UButton>
            </div>
          </template>
        </UTable>
      </UCard>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useLogisticsApi } from "~/composables/api";

definePageMeta({
  name: "shipping-billing",
});

const logisticsApi = useLogisticsApi();
const carrierId = ref("");
const summaryRows = ref<any[]>([]);
const itemRows = ref<any[]>([]);
const caseRows = ref<any[]>([]);
const settlementBatches = ref<any[]>([]);
const settlementDiffs = ref<any[]>([]);
const selectedSettlementBatchId = ref("");
const newCase = reactive({
  waybillId: "",
  reason: "",
});

const summaryColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "carrierName", header: "承运商" },
  { accessorKey: "waybillCount", header: "运单数" },
  { accessorKey: "estimatedFee", header: "预估运费" },
  { accessorKey: "actualFee", header: "实际运费" },
  { accessorKey: "diffFee", header: "差异" },
  { accessorKey: "abnormalCount", header: "异常单数" },
]);

const itemColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "orderId", header: "订单ID" },
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "carrierId", header: "承运商ID" },
  { accessorKey: "feeAmount", header: "预估运费" },
  { accessorKey: "actualFeeAmount", header: "实际运费" },
  { accessorKey: "feeDiffAmount", header: "差异" },
  { accessorKey: "billingStatus", header: "对账状态" },
]);

const caseColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "caseNo", header: "工单号" },
  { accessorKey: "waybillId", header: "运单ID" },
  { accessorKey: "carrierId", header: "承运商ID" },
  { accessorKey: "diffAmount", header: "差异金额" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "reason", header: "原因" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const settlementBatchColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "batchNo", header: "批次号" },
  { accessorKey: "carrierId", header: "承运商ID" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "waybillCount", header: "运单数" },
  { accessorKey: "diffCount", header: "待处理差异" },
  { accessorKey: "totalDiffAmount", header: "总差异金额" },
  { id: "actions", header: "操作" },
]);

const settlementDiffColumns = computed<TableColumn<any>[]>(() => [
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "attribution", header: "归因" },
  { accessorKey: "suggestion", header: "建议动作" },
  { accessorKey: "diffAmount", header: "差异金额" },
  { accessorKey: "status", header: "状态" },
  { id: "actions", header: "处理" },
]);

const caseStatusMeta = (status: string) => {
  switch (status) {
    case "confirmed":
      return { label: "已确认", color: "info" as const };
    case "appealed":
      return { label: "申诉中", color: "warning" as const };
    case "written_off":
      return { label: "已核销", color: "success" as const };
    default:
      return { label: "待处理", color: "neutral" as const };
  }
};

const loadSnapshot = async () => {
  const snapshot = await logisticsApi.getBillingSummary({
    carrier_id: carrierId.value || undefined,
  });
  summaryRows.value = snapshot.summary;
  itemRows.value = snapshot.items;
  caseRows.value = await logisticsApi.listBillingCases({
    carrier_id: carrierId.value || undefined,
  });
  await loadSettlementData();
};

const loadSettlementData = async () => {
  settlementBatches.value = await logisticsApi.listSettlementBatches({
    carrier_id: carrierId.value || undefined,
    limit: 20,
  });
  if (selectedSettlementBatchId.value) {
    settlementDiffs.value = await logisticsApi.listSettlementDiffs({
      batch_id: selectedSettlementBatchId.value,
      limit: 50,
    });
  }
};

const createSettlementBatch = async () => {
  const row = await logisticsApi.createSettlementBatch({
    carrier_id: carrierId.value || undefined,
  });
  selectedSettlementBatchId.value = row.id;
  await loadSettlementData();
};

const selectSettlementBatch = async (id: string) => {
  selectedSettlementBatchId.value = id;
  settlementDiffs.value = await logisticsApi.listSettlementDiffs({ batch_id: id, limit: 50 });
};

const handleSettlementDiff = async (id: string, action: "accept" | "dispute") => {
  await logisticsApi.handleSettlementDiff(id, {
    action,
    operator_id: "admin",
    note: `manual-${action}`,
  });
  if (selectedSettlementBatchId.value) {
    await selectSettlementBatch(selectedSettlementBatchId.value);
  }
  await loadSettlementData();
};

const confirmSettlementBatch = async (id: string) => {
  await logisticsApi.confirmSettlementBatch(id);
  await loadSettlementData();
};

const exportCsv = async () => {
  const payload = await logisticsApi.exportBilling({
    carrier_id: carrierId.value || undefined,
    format: "csv",
  });
  const text = String(payload?.content || "");
  if (!text) return;
  const blob = new Blob([text], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "billing-export.csv";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
};

const createCase = async () => {
  if (!newCase.waybillId) return;
  await logisticsApi.createBillingCase({
    waybill_id: newCase.waybillId,
    reason: newCase.reason || undefined,
  });
  newCase.waybillId = "";
  newCase.reason = "";
  await loadSnapshot();
};

const transitionCase = async (id: string, action: "confirm" | "appeal" | "writeoff") => {
  await logisticsApi.transitionBillingCase(id, {
    action,
    operator_id: "admin",
    note: `manual ${action}`,
  });
  await loadSnapshot();
};

onMounted(() => {
  loadSnapshot();
});
</script>
