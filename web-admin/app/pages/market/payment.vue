<!-- pages/payments.vue -->
<template>
  <div class="p-6 space-y-6">
    <!-- 标题 + 操作 -->
    <div class="flex justify-between items-center">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ $t("nav.payments") || "支付单" }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          管理订单支付、退款与对账
        </p>
      </div>
      <div class="flex gap-2">
        <USelect
          v-model="selectedStatus"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="全部状态"
          class="w-44"
        />
        <USelect
          v-model="selectedMethod"
          :options="methodOptions"
          option-attribute="label"
          value-attribute="value"
          placeholder="全部方式"
          class="w-44"
        />
        <UInput
          v-model="keyword"
          placeholder="搜索订单号/支付单号/客户"
          icon="i-heroicons-magnifying-glass"
          class="w-72"
        />
        <UButton
          variant="outline"
          icon="i-heroicons-funnel"
          @click="resetFilters"
        >
          重置
        </UButton>
        <UButton
          color="primary"
          icon="i-heroicons-arrow-down-tray"
          @click="exportCsv"
        >
          导出CSV
        </UButton>
      </div>
    </div>

    <!-- 汇总小卡 -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">支付笔数</p>
            <p class="text-2xl font-bold">
              {{ filteredRows.length?.toLocaleString() || '0' }}
            </p>
          </div>
          <UIcon
            name="i-heroicons-queue-list"
            class="w-7 h-7 text-primary-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">成功金额</p>
            <p class="text-2xl font-bold text-success-600">
              {{ formatCNY(totalPaid) }}
            </p>
          </div>
          <UIcon
            name="i-heroicons-banknotes"
            class="w-7 h-7 text-success-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">退款金额</p>
            <p class="text-2xl font-bold text-error-600">
              {{ formatCNY(totalRefund) }}
            </p>
          </div>
          <UIcon
            name="i-heroicons-arrow-uturn-left"
            class="w-7 h-7 text-error-500"
          />
        </div>
      </UCard>
      <UCard>
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-muted">待支付</p>
            <p class="text-2xl font-bold text-warning-600">
              {{ formatCNY(totalPending) }}
            </p>
          </div>
          <UIcon name="i-heroicons-clock" class="w-7 h-7 text-warning-500" />
        </div>
      </UCard>
    </div>

    <!-- 支付单表格 -->
    <UCard>
      <template #header>
        <h3 class="text-lg font-semibold">支付单列表</h3>
      </template>

      <UTable
        :data="filteredRows"
        :columns="columns"
        :loading="loading"
        sticky
        class="flex-1 max-h-[70vh]"
      >
        <!-- 编号 -->
        <template #paymentNo-cell="{ getValue }">
          <code class="text-xs">{{ getValue() }}</code>
        </template>

        <!-- 订单号 -->
        <template #orderId-cell="{ getValue }">
          <code class="text-xs">{{ getValue() }}</code>
        </template>

        <!-- 客户 -->
        <template #customer-cell="{ row }">
          <div class="flex items-center gap-2">
            <UAvatar :src="row.original.avatar" size="xs" />
            <div>
              <div class="font-medium">{{ row.original.customer }}</div>
              <div class="text-xs text-muted">
                {{ row.original.customerEmail }}
              </div>
            </div>
          </div>
        </template>

        <!-- 支付方式 -->
        <template #method-cell="{ getValue }">
          <div class="flex items-center gap-2">
            <UIcon :name="methodIcon(String(getValue()))" class="w-4 h-4" />
            <span>{{ methodLabel(String(getValue())) }}</span>
          </div>
        </template>

        <!-- 金额 -->
        <template #amount-cell="{ getValue }">
          <div class="text-right font-medium">
            {{ formatCNY(Number(getValue())) }}
          </div>
        </template>

        <!-- 状态 -->
        <template #status-cell="{ getValue }">
          <UBadge :color="statusColor(String(getValue()))" variant="soft">
            {{ statusLabel(String(getValue())) }}
          </UBadge>
        </template>

        <!-- 创建/完成时间 -->
        <template #createdAt-cell="{ getValue }">{{
          fmtDT(String(getValue()))
        }}</template>
        <template #paidAt-cell="{ getValue }">{{
          getValue() ? fmtDT(String(getValue())) : "-"
        }}</template>

        <!-- 操作 -->
        <template #actions-cell="{ row }">
          <div class="flex gap-2 justify-end">
            <UButton
              size="xs"
              variant="outline"
              icon="i-heroicons-eye"
              @click="openDetail(row.original)"
            >
              查看
            </UButton>
            <UButton
              v-if="row.original.status === 'paid'"
              size="xs"
              color="error"
              variant="soft"
              icon="i-heroicons-arrow-uturn-left"
              @click="requestRefund(row.original)"
            >
              退款
            </UButton>
            <UButton
              v-else-if="row.original.status === 'pending'"
              size="xs"
              color="primary"
              variant="soft"
              icon="i-heroicons-arrow-path"
              @click="retryPay(row.original)"
            >
              重试
            </UButton>
          </div>
        </template>
      </UTable>

      <template #footer>
        <div
          class="flex justify-between items-center text-sm px-4 py-3.5 border-t border-accented"
        >
          <div>共 {{ filteredRows.length }} 条</div>
          <div class="flex gap-6">
            <span
              >成功：<strong>{{ formatCNY(totalPaid) }}</strong></span
            >
            <span
              >退款：<strong>{{ formatCNY(totalRefund) }}</strong></span
            >
            <span
              >待付：<strong>{{ formatCNY(totalPending) }}</strong></span
            >
          </div>
        </div>
      </template>
    </UCard>

    <!-- 详情 Drawer -->
    <USlideover v-model="detailOpen">
      <UCard class="h-full">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm text-muted">支付单号</div>
              <div class="font-semibold">{{ current?.paymentNo }}</div>
            </div>
            <UBadge :color="current ? statusColor(current.status) : 'neutral'">
              {{ current ? statusLabel(current.status) : "" }}
            </UBadge>
          </div>
        </template>

        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-muted">订单号</div>
              <div class="font-medium">{{ current?.orderId }}</div>
            </div>
            <div>
              <div class="text-xs text-muted">渠道</div>
              <div class="font-medium">{{ current?.channelName }}</div>
            </div>
            <div>
              <div class="text-xs text-muted">客户</div>
              <div class="font-medium">{{ current?.customer }}</div>
              <div class="text-xs text-muted">{{ current?.customerEmail }}</div>
            </div>
            <div>
              <div class="text-xs text-muted">支付方式</div>
              <div class="font-medium flex items-center gap-2">
                <UIcon
                  :name="methodIcon(current?.method || '')"
                  class="w-4 h-4"
                />
                <span>{{ methodLabel(current?.method || "") }}</span>
              </div>
            </div>
            <div>
              <div class="text-xs text-muted">金额</div>
              <div class="text-lg font-semibold">
                {{ formatCNY(current?.amount || 0) }}
              </div>
            </div>
            <div>
              <div class="text-xs text-muted">手续费</div>
              <div class="font-medium">{{ formatCNY(current?.fee || 0) }}</div>
            </div>
            <div>
              <div class="text-xs text-muted">创建时间</div>
              <div class="font-medium">
                {{ current?.createdAt ? fmtDT(current.createdAt) : "-" }}
              </div>
            </div>
            <div>
              <div class="text-xs text-muted">完成时间</div>
              <div class="font-medium">
                {{ current?.paidAt ? fmtDT(current.paidAt) : "-" }}
              </div>
            </div>
          </div>

          <USeparator />

          <div>
            <div class="text-sm font-semibold mb-2">分账/明细</div>
            <div v-if="current?.items?.length" class="space-y-2">
              <div
                v-for="it in current?.items"
                :key="it.id"
                class="flex justify-between text-sm"
              >
                <div class="text-muted">{{ it.name }}</div>
                <div class="font-medium">{{ formatCNY(it.amount) }}</div>
              </div>
            </div>
            <div v-else class="text-sm text-muted">无明细</div>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton variant="outline" @click="detailOpen = false"
              >关闭</UButton
            >
            <UButton
              v-if="current?.status === 'paid'"
              color="error"
              icon="i-heroicons-arrow-uturn-left"
              @click="requestRefund(current!)"
            >
              退款
            </UButton>
          </div>
        </template>
      </UCard>
    </USlideover>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";

const { t } = useI18n();

type PaymentRow = {
  id: string;
  paymentNo: string;
  orderId: string;
  channelId: string;
  channelName: string;
  customer: string;
  customerEmail: string;
  avatar?: string;
  method: "wechat" | "alipay" | "card" | "bank" | "cod";
  amount: number;
  fee: number;
  status: "pending" | "paid" | "failed" | "refunded" | "partial_refund";
  createdAt: string;
  paidAt?: string | null;
  items?: { id: string; name: string; amount: number }[];
};

const loading = ref(false);
const keyword = ref("");
const selectedStatus = ref<string | "">("");
const selectedMethod = ref<string | "">("");

// 选项
const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "待支付", value: "pending" },
  { label: "已支付", value: "paid" },
  { label: "失败", value: "failed" },
  { label: "已退款", value: "refunded" },
  { label: "部分退款", value: "partial_refund" },
];
const methodOptions = [
  { label: "全部方式", value: "" },
  { label: "微信支付", value: "wechat" },
  { label: "支付宝", value: "alipay" },
  { label: "银行卡", value: "card" },
  { label: "银行转账", value: "bank" },
  { label: "货到付款", value: "cod" },
];

// 列（v3 TanStack 风格）
const columns = computed<TableColumn<PaymentRow>[]>(() => [
  { accessorKey: "paymentNo", header: "支付单号" },
  { accessorKey: "orderId", header: "订单号" },
  { accessorKey: "customer", header: "客户" },
  { accessorKey: "channelName", header: "渠道" },
  { accessorKey: "method", header: "方式" },
  {
    id: "amount",
    accessorKey: "amount",
    header: "金额",
    meta: { class: { td: "text-right" } },
  },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
  { accessorKey: "paidAt", header: "完成时间" },
  {
    id: "actions",
    header: "操作",
    meta: { class: { td: "text-right" } },
  },
]);

// 示例数据
const rows = ref<PaymentRow[]>([
  {
    id: "1",
    paymentNo: "PAY202401150001",
    orderId: "ORD001",
    channelId: "ONLINE",
    channelName: "官网自营",
    customer: "张三",
    customerEmail: "zhangsan@example.com",
    avatar: "https://i.pravatar.cc/80?img=1",
    method: "wechat",
    amount: 8999,
    fee: 90,
    status: "paid",
    createdAt: "2024-01-15T10:10:00",
    paidAt: "2024-01-15T10:11:35",
    items: [{ id: "i1", name: "iPhone 15 Pro", amount: 8999 }],
  },
  {
    id: "2",
    paymentNo: "PAY202401150002",
    orderId: "ORD002",
    channelId: "JD",
    channelName: "京东旗舰店",
    customer: "李四",
    customerEmail: "lisi@example.com",
    avatar: "https://i.pravatar.cc/80?img=2",
    method: "alipay",
    amount: 12999,
    fee: 130,
    status: "pending",
    createdAt: "2024-01-15T09:15:00",
    paidAt: null,
  },
  {
    id: "3",
    paymentNo: "PAY202401140003",
    orderId: "ORD003",
    channelId: "TMALL",
    channelName: "天猫旗舰店",
    customer: "王五",
    customerEmail: "wangwu@example.com",
    avatar: "https://i.pravatar.cc/80?img=3",
    method: "card",
    amount: 1899,
    fee: 19,
    status: "refunded",
    createdAt: "2024-01-14T16:45:00",
    paidAt: "2024-01-14T16:45:50",
    items: [{ id: "i2", name: "AirPods Pro", amount: 1899 }],
  },
  {
    id: "4",
    paymentNo: "PAY202401130004",
    orderId: "ORD004",
    channelId: "OFFLINE",
    channelName: "线下门店",
    customer: "赵六",
    customerEmail: "zhaoliu@example.com",
    avatar: "https://i.pravatar.cc/80?img=4",
    method: "bank",
    amount: 4599,
    fee: 0,
    status: "failed",
    createdAt: "2024-01-13T14:20:00",
    paidAt: null,
  },
  {
    id: "5",
    paymentNo: "PAY202401120005",
    orderId: "ORD005",
    channelId: "ONLINE",
    channelName: "官网自营",
    customer: "钱七",
    customerEmail: "qianqi@example.com",
    avatar: "https://i.pravatar.cc/80?img=5",
    method: "wechat",
    amount: 9999,
    fee: 100,
    status: "partial_refund",
    createdAt: "2024-01-12T11:20:00",
    paidAt: "2024-01-12T11:20:50",
    items: [
      { id: "i3", name: "iPad Air", amount: 4599 },
      { id: "i4", name: "Apple Pencil", amount: 599 },
      { id: "i5", name: "保护壳", amount: 199 },
    ],
  },
]);

// 过滤
const filteredRows = computed(() => {
  const q = keyword.value.trim().toLowerCase();
  return rows.value.filter((r) => {
    const okQ =
      !q ||
      r.paymentNo.toLowerCase().includes(q) ||
      r.orderId.toLowerCase().includes(q) ||
      r.customer.toLowerCase().includes(q);
    const okS = !selectedStatus.value || r.status === selectedStatus.value;
    const okM = !selectedMethod.value || r.method === selectedMethod.value;
    return okQ && okS && okM;
  });
});

// 汇总
const totalPaid = computed(() =>
  filteredRows.value
    .filter((r) => r.status === "paid" || r.status === "partial_refund")
    .reduce((s, r) => s + r.amount, 0)
);
const totalRefund = computed(() =>
  filteredRows.value
    .filter((r) => r.status === "refunded" || r.status === "partial_refund")
    .reduce((s, r) => s + Math.min(r.amount * 0.5, r.amount), 0)
); // demo：假设部分退款50%
const totalPending = computed(() =>
  filteredRows.value
    .filter((r) => r.status === "pending")
    .reduce((s, r) => s + r.amount, 0)
);

// 工具
const formatCNY = (n: number) =>
  new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    maximumFractionDigits: 0,
  }).format(n || 0);
const fmtDT = (s: string) =>
  new Date(s).toLocaleString("zh-CN", { hour12: false });

function statusLabel(s: PaymentRow["status"]) {
  return (
    (
      {
        pending: "待支付",
        paid: "已支付",
        failed: "失败",
        refunded: "已退款",
        partial_refund: "部分退款",
      } as const
    )[s] || s
  );
}
function statusColor(s: PaymentRow["status"]) {
  return (
    (
      {
        pending: "warning",
        paid: "success",
        failed: "error",
        refunded: "neutral",
        partial_refund: "info",
      } as const
    )[s] || "neutral"
  );
}
function methodLabel(m: PaymentRow["method"]) {
  return (
    (
      {
        wechat: "微信支付",
        alipay: "支付宝",
        card: "银行卡",
        bank: "银行转账",
        cod: "货到付款",
      } as const
    )[m] || m
  );
}
function methodIcon(m: PaymentRow["method"] | "") {
  return (
    (
      {
        wechat: "i-simple-icons-wechat",
        alipay: "i-simple-icons-alipay",
        card: "i-heroicons-credit-card",
        bank: "i-heroicons-building-library",
        cod: "i-heroicons-truck",
      } as Record<string, string>
    )[m] || "i-heroicons-question-mark-circle"
  );
}

function resetFilters() {
  keyword.value = "";
  selectedStatus.value = "";
  selectedMethod.value = "";
}

function exportCsv() {
  const header = [
    "PaymentNo",
    "OrderId",
    "Customer",
    "Channel",
    "Method",
    "Amount",
    "Status",
    "CreatedAt",
    "PaidAt",
  ];
  const body = filteredRows.value.map((r) => [
    r.paymentNo,
    r.orderId,
    r.customer,
    r.channelName,
    methodLabel(r.method),
    r.amount,
    statusLabel(r.status),
    r.createdAt,
    r.paidAt || "",
  ]);
  const csv = [header, ...body]
    .map((row) =>
      row.map((v) => `"${String(v).replace(/"/g, '""')}"`).join(",")
    )
    .join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "payments.csv";
  a.click();
  URL.revokeObjectURL(url);
}

// 明细/退款/重试（示例）
const toast = useToastAlert();
const detailOpen = ref(false);
const current = ref<PaymentRow | null>(null);
function openDetail(row: PaymentRow) {
  current.value = row;
  detailOpen.value = true;
}
function requestRefund(row: PaymentRow) {
  // 这里调用后端退款接口；演示先提示
  toast.add({ title: `已提交退款申请：${row.paymentNo}`, color: "info" });
}
function retryPay(row: PaymentRow) {
  toast.add({ title: `已触发重试：${row.paymentNo}`, color: "primary" });
}
</script>
