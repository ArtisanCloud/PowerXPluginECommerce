<template>
  <div class="p-6 space-y-6">
    <div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">逆向运单</h1>
        <p class="text-gray-500 dark:text-gray-400">
          售后单逆向物流闭环，跟踪回仓并输出补偿建议。
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <UInput v-model="form.orderId" class="w-40" placeholder="订单ID" />
        <UInput v-model="form.afterSaleId" class="w-40" placeholder="售后ID" />
        <UButton color="primary" icon="i-heroicons-plus" @click="createWaybill">创建逆向运单</UButton>
      </div>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">运单列表</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">点击“详情”查看轨迹与补偿建议。</p>
          </div>
          <UBadge color="info" variant="subtle">{{ waybills.length }} 条</UBadge>
        </div>
      </template>

      <UTable :columns="columns" :data="waybills">
        <template #status-cell="{ getValue }">
          <UBadge :color="statusMeta(getValue()).color" variant="subtle">
            {{ statusMeta(getValue()).label }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" @click="selectWaybill(row.original)">详情</UButton>
            <UButton size="xs" variant="ghost" color="primary" @click="markReceived(row.original.id)">
              标记回仓
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard v-if="detail">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              逆向详情 - {{ detail.waybill.waybillNo }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              售后 {{ detail.waybill.afterSaleId }} · 订单 {{ detail.waybill.orderId }}
            </p>
          </div>
          <UBadge color="warning" variant="subtle">
            补偿建议：{{ detail.compensation.recommendation || "none" }}
          </UBadge>
        </div>
      </template>

      <div class="grid gap-4 md:grid-cols-2">
        <UCard>
          <template #header><div class="font-semibold">轨迹</div></template>
          <ol class="space-y-2">
            <li v-for="item in detail.tracking" :key="item.id" class="rounded border border-gray-100 p-3 dark:border-gray-800">
              <p class="text-sm font-medium text-gray-900 dark:text-white">{{ item.status }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ item.occurredAt || "-" }} · {{ item.description || "-" }}</p>
            </li>
          </ol>
        </UCard>

        <UCard>
          <template #header><div class="font-semibold">回仓结论</div></template>
          <div class="space-y-3">
            <div v-for="result in detail.warehouseResults" :key="result.id" class="rounded border border-gray-100 p-3 dark:border-gray-800">
              <p class="text-sm font-medium text-gray-900 dark:text-white">{{ result.result }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ result.notes || "-" }}</p>
            </div>
            <UButton color="primary" variant="soft" @click="recordResult">
              记录回仓结论（resellable/restock）
            </UButton>
          </div>
        </UCard>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useReverseApi } from "~/composables/api";

definePageMeta({
  name: "shipping-reverse-waybills",
});

type WaybillRow = {
  id: string;
  orderId: string;
  afterSaleId: string;
  waybillNo: string;
  status: string;
  updatedAt: string;
};

const reverseApi = useReverseApi();

const form = reactive({
  orderId: "",
  afterSaleId: "",
});

const waybills = ref<WaybillRow[]>([]);
const detail = ref<any>(null);

const columns = computed<TableColumn<WaybillRow>[]>(() => [
  { accessorKey: "orderId", header: "订单ID" },
  { accessorKey: "afterSaleId", header: "售后ID" },
  { accessorKey: "waybillNo", header: "运单号" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "updatedAt", header: "更新时间" },
  { id: "actions", header: "操作" },
]);

const statusMeta = (status: string) => {
  switch (status) {
    case "in_transit":
      return { label: "逆向运输中", color: "info" as const };
    case "received":
      return { label: "已回仓", color: "warning" as const };
    case "closed":
      return { label: "已关闭", color: "success" as const };
    default:
      return { label: "已创建", color: "neutral" as const };
  }
};

const loadWaybills = async () => {
  const rows = await reverseApi.listWaybills();
  waybills.value = rows.map((item) => ({
    id: item.id,
    orderId: item.orderId,
    afterSaleId: item.afterSaleId,
    waybillNo: item.waybillNo,
    status: item.status,
    updatedAt: item.updatedAt ? item.updatedAt.replace("T", " ").slice(0, 16) : "-",
  }));
};

const createWaybill = async () => {
  if (!form.orderId || !form.afterSaleId) return;
  await reverseApi.createWaybill({
    order_id: form.orderId,
    after_sale_id: form.afterSaleId,
  });
  form.orderId = "";
  form.afterSaleId = "";
  await loadWaybills();
};

const selectWaybill = async (row: WaybillRow) => {
  detail.value = await reverseApi.getWaybillDetail(row.id);
};

const markReceived = async (id: string) => {
  await reverseApi.appendTracking(id, {
    status: "in_transit",
    description: "逆向包裹运输中",
  });
  await reverseApi.appendTracking(id, {
    status: "received",
    description: "仓库已签收逆向包裹",
  });
  await loadWaybills();
  if (detail.value?.waybill?.id === id) {
    detail.value = await reverseApi.getWaybillDetail(id);
  }
};

const recordResult = async () => {
  const id = detail.value?.waybill?.id;
  if (!id) return;
  await reverseApi.recordWarehouseResult(id, {
    result: "resellable",
    disposition: "restock",
    notes: "包装完整，允许返库",
  });
  detail.value = await reverseApi.getWaybillDetail(id);
  await loadWaybills();
};

onMounted(() => {
  loadWaybills();
});
</script>

