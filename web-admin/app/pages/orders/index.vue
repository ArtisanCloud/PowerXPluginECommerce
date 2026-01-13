<template>
  <div class="space-y-4">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold">订单管理</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">后台订单列表（按租户）</p>
      </div>

      <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
        <USelect v-model="status" :items="statusItems" class="w-48" placeholder="全部状态" />
        <UInput v-model.trim="customerId" class="w-56" placeholder="客户ID（可选）" />
        <UButton color="primary" :loading="loading" @click="refresh">查询</UButton>
      </div>
    </div>

    <UCard>
      <UTable :columns="columns" :data="rows" :loading="loading">
        <template #totalMinor-cell="{ row }">
          <div class="text-right tabular-nums">
            {{ formatMoney(row.original.currency, row.original.totalMinor) }}
          </div>
        </template>

        <template #status-cell="{ row }">
          <UBadge :color="statusColor(row.original.status)" variant="subtle">
            {{ statusLabel(row.original.status) }}
          </UBadge>
        </template>

        <template #createdAt-cell="{ row }">
          <span class="text-sm">{{ formatTime(row.original.createdAt) }}</span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton size="xs" variant="ghost" @click="goDetail(row.original.orderId)">详情</UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div
          class="flex flex-col gap-3 text-sm text-gray-500 dark:text-gray-400 lg:flex-row lg:items-center lg:justify-between"
        >
          <span>共 {{ total }} 条订单</span>
          <div class="flex items-center gap-3">
            <USelect v-model="pageSize" :items="pageSizeItems" class="w-24" />
            <UPagination v-model="page" :total="total" :page-count="pageSize" show-first show-last />
          </div>
        </div>
      </template>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useOrderApi } from "~/composables/api/useOrder";
import type { OrderSummary } from "~/types/order";

definePageMeta({
  name: "orders",
});

type OrderRow = {
  orderId: string;
  orderNo: string;
  status: string;
  currency: string;
  totalMinor: number;
  createdAt: string;
};

const api = useOrderApi();
const toast = useToastAlert();
const router = useRouter();

const loading = ref(false);
const items = ref<OrderSummary[]>([]);
const total = ref(0);

const page = ref(1);
const pageSize = ref(20);
const status = ref<string>("");
const customerId = ref<string>("");

const statusItems = [
  { label: "全部状态", value: "" },
  { label: "待支付", value: "pending_payment" },
  { label: "已支付", value: "paid" },
  { label: "已取消", value: "cancelled" },
  { label: "草稿", value: "draft" },
];

const pageSizeItems = [
  { label: "10", value: 10 },
  { label: "20", value: 20 },
  { label: "50", value: 50 },
  { label: "100", value: 100 },
];

const rows = computed<OrderRow[]>(() =>
  (items.value || []).map((it) => ({
    orderId: it.orderId,
    orderNo: it.orderNo,
    status: it.status,
    currency: it.amounts?.currency || "CNY",
    totalMinor: Number(it.amounts?.total || 0),
    createdAt: it.createdAt,
  })),
);

const columns = computed<TableColumn<OrderRow>[]>(() => [
  { accessorKey: "orderNo", header: "订单号" },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "totalMinor", header: "金额", meta: { class: { td: "text-right" } } },
  { accessorKey: "createdAt", header: "创建时间" },
  { id: "actions", header: "操作" },
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

const refresh = async () => {
  loading.value = true;
  try {
    const resp = await api.listOrders({
      page: page.value,
      pageSize: pageSize.value,
      status: status.value || undefined,
      customerId: customerId.value || undefined,
    });
    items.value = resp.items || [];
    total.value = Number(resp.total || 0);
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

watch([page, pageSize], () => {
  refresh();
});

const goDetail = (id: string) => {
  router.push(`/orders/${id}`);
};

await refresh();
</script>
