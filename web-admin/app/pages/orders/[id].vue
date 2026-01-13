<template>
  <div class="space-y-4">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-2">
        <UButton icon="i-heroicons-arrow-left" variant="ghost" color="neutral" @click="back">
          返回
        </UButton>
        <div>
          <h1 class="text-2xl font-semibold">订单详情</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ detail?.summary?.orderNo || route.params.id }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <UBadge v-if="detail?.summary" :color="statusColor(detail.summary.status)" variant="subtle">
          {{ statusLabel(detail.summary.status) }}
        </UBadge>
        <UButton
          v-if="canCancel"
          color="warning"
          variant="soft"
          :loading="cancelling"
          @click="openCancel"
        >
          取消订单
        </UButton>
      </div>
    </div>

    <UCard v-if="detail?.summary" :ui="{ body: { base: 'space-y-3' } }">
      <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">订单ID</div>
          <div class="text-sm break-all">{{ detail.summary.orderId }}</div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">订单号</div>
          <div class="text-sm">{{ detail.summary.orderNo }}</div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">创建时间</div>
          <div class="text-sm">{{ formatTime(detail.summary.createdAt) }}</div>
        </div>
        <div class="text-right">
          <div class="text-xs text-gray-500 dark:text-gray-400">订单金额</div>
          <div class="text-lg font-semibold tabular-nums">
            {{ formatMoney(detail.summary.amounts.currency, detail.summary.amounts.total) }}
          </div>
        </div>
      </div>
    </UCard>

    <UCard title="订单明细">
      <UTable :columns="itemColumns" :data="detail?.items || []" :loading="loading">
        <template #unitPrice-cell="{ row }">
          <div class="text-right tabular-nums">
            {{ formatMoney(detail?.summary?.amounts?.currency || "CNY", row.original.unitPrice) }}
          </div>
        </template>
        <template #lineAmount-cell="{ row }">
          <div class="text-right tabular-nums">
            {{ formatMoney(detail?.summary?.amounts?.currency || "CNY", row.original.lineAmount) }}
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard title="事件日志">
      <UTable :columns="eventColumns" :data="detail?.events || []" :loading="loading">
        <template #createdAt-cell="{ row }">
          <span class="text-sm">{{ formatTime(row.original.createdAt) }}</span>
        </template>
      </UTable>
    </UCard>

    <UModal
      v-model:open="cancelOpen"
      title="确认取消订单？"
      description="仅 pending_payment 状态允许取消。取消后会释放库存锁定。"
      :prevent-close="cancelling"
    >
      <template #body>
        <div class="p-4 sm:p-5">
          <UFormField label="取消原因（可选）">
            <UTextarea v-model.trim="cancelReason" :rows="3" placeholder="例如：客户要求取消" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="cancelling" @click="cancelOpen = false">
            返回
          </UButton>
          <UButton color="warning" type="button" :loading="cancelling" @click="confirmCancel">
            确认取消
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useOrderApi } from "~/composables/api/useOrder";
import type { OrderDetail, OrderEvent, OrderItem } from "~/types/order";

definePageMeta({
  name: "order-detail",
});

const route = useRoute();
const router = useRouter();
const toast = useToastAlert();
const api = useOrderApi();

const loading = ref(false);
const cancelling = ref(false);
const detail = ref<OrderDetail | null>(null);

const cancelOpen = ref(false);
const cancelReason = ref("");

const id = computed(() => String(route.params.id || ""));

const itemColumns = computed<TableColumn<OrderItem>[]>(() => [
  { accessorKey: "skuId", header: "SKU" },
  { accessorKey: "qty", header: "数量" },
  { accessorKey: "unitPrice", header: "单价", meta: { class: { td: "text-right" } } },
  { accessorKey: "lineAmount", header: "行金额", meta: { class: { td: "text-right" } } },
]);

const eventColumns = computed<TableColumn<OrderEvent>[]>(() => [
  { accessorKey: "eventType", header: "事件" },
  { accessorKey: "operatorType", header: "操作者类型" },
  { accessorKey: "operator", header: "操作者" },
  { accessorKey: "createdAt", header: "时间" },
]);

const formatMoney = (currency: string, minor: number) => {
  const amount = Number(minor || 0) / 100;
  try {
    return new Intl.NumberFormat("zh-CN", { style: "currency", currency }).format(amount);
  } catch {
    return `${amount.toFixed(2)} ${currency}`;
  }
};

const formatTime = (raw: string) => {
  const d = raw ? new Date(raw) : null;
  if (!d || Number.isNaN(d.getTime())) return "-";
  return d.toLocaleString("zh-CN");
};

const statusLabel = (st: string) => {
  const map: Record<string, string> = {
    pending_payment: "待支付",
    paid: "已支付",
    cancelled: "已取消",
    draft: "草稿",
  };
  return map[st] || st || "-";
};

const statusColor = (st: string) => {
  const map: Record<string, "warning" | "success" | "neutral" | "info" | "error" | "primary"> = {
    pending_payment: "warning",
    paid: "success",
    cancelled: "neutral",
    draft: "info",
  };
  return map[st] || "neutral";
};

const canCancel = computed(() => detail.value?.summary?.status === "pending_payment");

const fetchDetail = async () => {
  if (!id.value) return;
  loading.value = true;
  try {
    detail.value = await api.getOrder(id.value);
  } catch (e: any) {
    toast.add({
      title: "加载订单失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    loading.value = false;
  }
};

const openCancel = () => {
  cancelReason.value = "";
  cancelOpen.value = true;
};

const confirmCancel = async () => {
  if (!id.value) return;
  cancelling.value = true;
  try {
    await api.cancelOrder(id.value, { reason: cancelReason.value || undefined });
    toast.add({ title: "取消成功", color: "success" });
    cancelOpen.value = false;
    await fetchDetail();
  } catch (e: any) {
    toast.add({
      title: "取消失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    cancelling.value = false;
  }
};

const back = () => router.push("/orders");

watch(id, () => fetchDetail(), { immediate: true });
</script>

