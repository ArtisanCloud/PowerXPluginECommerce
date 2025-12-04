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
        :data="filteredOrders"
        :columns="columns"
        :loading="false"
        class="w-full"
      >
        <!-- v3: 用 -cell，而不是 -data -->
        <template #amount-cell="{ getValue }">
          <div class="text-right font-medium">
            {{
              new Intl.NumberFormat("zh-CN", {
                style: "currency",
                currency: "CNY",
              }).format(Number(getValue() || 0))
            }}
          </div>
        </template>

        <template #orderTime-cell="{ getValue }">
          <span>
            {{ new Date(getValue() as string).toLocaleString("zh-CN") }}
          </span>
        </template>

        <template #status-cell="{ getValue }">
          <UBadge :color="getStatusColor(String(getValue()))" variant="subtle">
            {{ $t(`orders.${String(getValue())}`) }}
          </UBadge>
        </template>

        <template #actions-cell>
          <div class="flex gap-2">
            <UButton
              color="neutral"
              variant="ghost"
              size="sm"
              icon="i-heroicons-eye"
            >
              {{ $t("common.view") }}
            </UButton>
            <UButton
              color="primary"
              variant="ghost"
              size="sm"
              icon="i-heroicons-cog-6-tooth"
            >
              {{ $t("common.edit") }}
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
const { t } = useI18n();

// 状态筛选
const selectedStatus = ref<string>("");

// 状态选项
const statusOptions = [
  { value: "", label: t("orders.allStatus") },
  { value: "pending", label: t("orders.pending") },
  { value: "processing", label: t("orders.processing") },
  { value: "shipped", label: t("orders.shipped") },
  { value: "delivered", label: t("orders.delivered") },
];

// 数据类型
type Order = {
  id: string;
  customer: string;
  product: string;
  amount: number; // 用 number 存储
  status: "pending" | "processing" | "shipped" | "delivered";
  orderTime: string; // ISO/可被 Date 解析的字符串
};

// 表格列定义（v3 TanStack 风格；用 computed 以便 i18n 切换时表头刷新）
const columns = computed<TableColumn<Order>[]>(() => [
  { accessorKey: "id", header: t("orders.orderId") },
  { accessorKey: "customer", header: t("orders.customer") },
  { accessorKey: "product", header: t("orders.product") },
  {
    accessorKey: "amount",
    header: t("orders.amount"),
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "status", header: t("orders.status") },
  { accessorKey: "orderTime", header: t("orders.orderTime") },
  { id: "actions", header: t("orders.actions") },
]);

// 示例数据（amount 改为 number；orderTime 改成可解析格式）
const orders = ref<Order[]>([
  {
    id: "ORD001",
    customer: "张三",
    product: "iPhone 15 Pro",
    amount: 8999,
    status: "pending",
    orderTime: "2024-01-15T10:30:00",
  },
  {
    id: "ORD002",
    customer: "李四",
    product: "MacBook Pro",
    amount: 12999,
    status: "processing",
    orderTime: "2024-01-15T09:15:00",
  },
  {
    id: "ORD003",
    customer: "王五",
    product: "AirPods Pro",
    amount: 1899,
    status: "shipped",
    orderTime: "2024-01-14T16:45:00",
  },
  {
    id: "ORD004",
    customer: "赵六",
    product: "iPad Air",
    amount: 4599,
    status: "delivered",
    orderTime: "2024-01-13T14:20:00",
  },
]);

// 筛选订单
const filteredOrders = computed(() => {
  if (!selectedStatus.value) return orders.value;
  return orders.value.filter((o) => o.status === selectedStatus.value);
});

// 获取状态颜色（语义色）
const getStatusColor = (status: string) => {
  const colorMap: Record<
    string,
    "warning" | "info" | "primary" | "success" | "neutral"
  > = {
    pending: "warning",
    processing: "info",
    shipped: "primary",
    delivered: "success",
  };
  return colorMap[status] || "neutral";
};
</script>
