<template>
  <div>
    <!-- 页面标题和筛选 -->
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
        {{ $t("orders.title") }}
      </h1>
      <USelect
        v-model="selectedStatus"
        :options="statusOptions"
        option-attribute="label"
        value-attribute="value"
        :placeholder="$t('orders.allStatus')"
        class="w-48"
      />
    </div>

    <!-- 订单表格 -->
    <UCard>
      <UTable
        :data="rows"
        :columns="columns"
        :loading="loading"
        class="w-full"
      >
        <!-- v3: 用 -cell，而不是 -data -->
        <template #amountMinor-cell="{ row }">
          <div class="text-right font-medium">
            {{
              new Intl.NumberFormat("zh-CN", {
                style: "currency",
                currency: row.original.currency || "CNY",
              }).format(Number(row.original.amountMinor || 0) / 100)
            }}
          </div>
        </template>

        <template #customerName-cell="{ row }">
          <NuxtLink
            v-if="row.original.customerId"
            class="text-primary-500 hover:underline"
            :to="customerDetailTo(row.original.customerId)"
          >
            {{ row.original.customerName || row.original.customerId }}
          </NuxtLink>
          <span v-else>-</span>
        </template>

        <template #source-cell="{ row }">
          <UBadge
            color="neutral"
            variant="subtle"
            :title="row.original.channel ? `channel: ${row.original.channel}` : ''"
          >
            {{ sourceText(row.original.createdByType) }}
          </UBadge>
          <span v-if="row.original.channel" class="ml-2 text-xs text-gray-500 dark:text-gray-400">
            {{ row.original.channel }}
          </span>
        </template>

        <template #createdAt-cell="{ getValue }">
          <span>
            {{ new Date(getValue() as string).toLocaleString("zh-CN") }}
          </span>
        </template>

        <template #status-cell="{ getValue }">
          <UBadge :color="getStatusColor(String(getValue()))" variant="subtle">
            {{ statusText(String(getValue())) }}
          </UBadge>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
              :disabled="!row.original.orderId"
              @click="goView(row.original.orderId)"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-cog-6-tooth"
              :disabled="!row.original.orderId"
              @click="goEdit(row.original.orderId)"
            >
              {{ $t("common.edit") }}
            </UButton>
            <UButton
              v-if="row.original.status === 'pending_payment'"
              color="error"
              variant="ghost"
              size="sm"
              icon="i-heroicons-trash"
              :disabled="!row.original.orderId || deleting"
              @click="openDelete(row.original.orderId)"
            >
              {{ $t("orders.delete") }}
            </UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div
          class="flex flex-col gap-3 text-sm text-gray-500 dark:text-gray-400 lg:flex-row lg:items-center lg:justify-between"
        >
          <span>{{ $t("common.total", { count: total }) }}</span>
          <div class="flex items-center gap-3">
            <USelect v-model="pageSize" :options="pageSizeItems" option-attribute="label" value-attribute="value" class="w-24" />
            <UPagination v-model="page" :total="total" :page-count="pageSize" show-first show-last />
          </div>
        </div>
      </template>
    </UCard>

    <UModal
      v-model:open="deleteOpen"
      :title="$t('orders.deleteConfirmTitle')"
      :description="$t('orders.deleteConfirmDesc')"
      :prevent-close="deleting"
    >
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="deleting" @click="deleteOpen = false">
            {{ $t("common.cancel") }}
          </UButton>
          <UButton color="error" type="button" :loading="deleting" @click="confirmDelete">
            {{ $t("orders.delete") }}
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useOrderApi, type OrderListQuery } from "~/composables/api/useOrder";
import type { OrderSummary } from "~/types/order";
import { useCustomerService } from "~/composables/api/services/customerService";

const { t, te } = useI18n();
const toast = useToastAlert();
const api = useOrderApi();
const customerService = useCustomerService();
const router = useRouter();

// 状态筛选：Nuxt UI Select 的 item value 不能是空字符串；用 undefined 表示“未选择/全部”
const selectedStatus = ref<string | undefined>(undefined);

const statusItems = [
  { value: "pending_payment", labelKey: "orders.statuses.pending_payment" },
  { value: "paid", labelKey: "orders.statuses.paid" },
  { value: "cancelled", labelKey: "orders.statuses.cancelled" },
  { value: "draft", labelKey: "orders.statuses.draft" },
] as const;

const statusOptions = computed(() =>
  statusItems.map((it) => ({
    value: it.value,
    label: t(it.labelKey),
  })),
);

type OrderRow = {
  orderId: string;
  orderNo: string;
  customerId?: string;
  customerName?: string;
  channel?: string;
  createdByType?: string;
  source: string;
  currency: string;
  amountMinor: number;
  status: string;
  createdAt: string;
};

// 表格列定义（v3 TanStack 风格；用 computed 以便 i18n 切换时表头刷新）
const columns = computed<TableColumn<OrderRow>[]>(() => [
  { accessorKey: "orderNo", header: t("orders.orderNo") },
  { accessorKey: "customerName", header: t("orders.customer") },
  { accessorKey: "source", header: t("orders.source") },
  {
    accessorKey: "amountMinor",
    header: t("orders.amount"),
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "status", header: t("orders.status") },
  { accessorKey: "createdAt", header: t("orders.createdAt") },
  { id: "actions", header: t("orders.actions") },
]);

const loading = ref(false);
const items = ref<OrderSummary[]>([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const deleteOpen = ref(false);
const deleting = ref(false);
const deleteTargetId = ref<string>("");

const pageSizeItems = [
  { label: "10", value: 10 },
  { label: "20", value: 20 },
  { label: "50", value: 50 },
  { label: "100", value: 100 },
];

const customerNameById = ref<Record<string, string>>({});
const inflightCustomerIds = new Set<string>();

const rows = computed<OrderRow[]>(() =>
  (items.value || []).map((it) => ({
    orderId: it.orderId,
    orderNo: it.orderNo,
    customerId: it.customerId || (it as any).customer_id,
    customerName:
      (it.customerId || (it as any).customer_id)
        ? customerNameById.value[String(it.customerId || (it as any).customer_id)]
        : undefined,
    channel: it.channel || (it as any).channel,
    createdByType: it.createdByType || (it as any).created_by_type,
    source: sourceText(it.createdByType || (it as any).created_by_type),
    currency: it.amounts?.currency || "CNY",
    amountMinor: Number(it.amounts?.total || 0),
    status: it.status,
    createdAt: it.createdAt,
  })),
);

const statusText = (status: string) => {
  const key = `orders.statuses.${status}`;
  return te(key) ? t(key) : status || "-";
};

const sourceText = (createdByType?: string) => {
  const st = String(createdByType || "").trim();
  const key = st ? `orders.sources.${st}` : "";
  if (key && te(key)) return t(key);
  if (!st) return "-";
  return st;
};

// 获取状态颜色（语义色）
const getStatusColor = (status: string) => {
  const colorMap: Record<
    string,
    "warning" | "info" | "primary" | "success" | "neutral"
  > = {
    pending_payment: "warning",
    paid: "success",
    shipped: "primary",
    cancelled: "neutral",
    draft: "info",
  };
  return colorMap[status] || "neutral";
};

const customerDetailTo = (customerId: string) => ({
  path: "/customer",
  query: { customerId },
});

const hydrateCustomerNames = async (next: OrderSummary[]) => {
  const ids = Array.from(
    new Set(
      (next || [])
        .map((it) => String(it.customerId || (it as any).customer_id || "").trim())
        .filter(Boolean),
    ),
  );

  const missing = ids.filter(
    (id) => !customerNameById.value[id] && !inflightCustomerIds.has(id),
  );
  if (!missing.length) return;

  await Promise.allSettled(
    missing.map(async (id) => {
      inflightCustomerIds.add(id);
      try {
        const customer = await customerService.getCustomer(id);
        const name = String(customer?.name || "").trim();
        customerNameById.value[id] = name || id;
      } catch {
        customerNameById.value[id] = id;
      } finally {
        inflightCustomerIds.delete(id);
      }
    }),
  );
};

const goView = (orderId: string) => {
  if (!orderId) return;
  router.push(`/orders/${orderId}`);
};

const goEdit = (orderId: string) => {
  if (!orderId) return;
  router.push({ path: `/orders/${orderId}`, query: { mode: "edit" } });
};

const openDelete = (orderId: string) => {
  if (!orderId) return;
  deleteTargetId.value = orderId;
  deleteOpen.value = true;
};

const confirmDelete = async () => {
  const orderId = String(deleteTargetId.value || "").trim();
  if (!orderId) return;
  deleting.value = true;
  try {
    await api.cancelOrder(orderId, { reason: t("orders.deleteReason") });
    toast.add({ title: t("orders.deleteSuccess"), color: "success" });
    deleteOpen.value = false;
    deleteTargetId.value = "";
    await refresh();
  } catch (e: any) {
    toast.add({
      title: t("orders.deleteFailedTitle"),
      description: e?.message || t("orders.deleteFailedDesc"),
      color: "error",
    });
  } finally {
    deleting.value = false;
  }
};

const refresh = async () => {
  loading.value = true;
  try {
    const query: OrderListQuery = {
      page: page.value,
      pageSize: pageSize.value,
      status: selectedStatus.value || undefined,
    };
    const resp = await api.listOrders(query);
    items.value = resp.items || [];
    total.value = Number(resp.total || 0);
    await hydrateCustomerNames(items.value);
  } catch (e: any) {
    toast.add({
      title: t("orders.loadFailedTitle"),
      description: e?.message || t("orders.loadFailedDesc"),
      color: "error",
    });
  } finally {
    loading.value = false;
  }
};

watch([page, pageSize], () => refresh());
watch(selectedStatus, () => {
  page.value = 1;
  refresh();
});

await refresh();
</script>
