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
          :items="statusItems"
          placeholder="全部状态"
          class="w-44"
        />
        <USelect
          v-model="selectedMethod"
          :items="methodItems"
          placeholder="全部方式"
          class="w-44"
        />
        <UInput
          v-model="keyword"
          placeholder="搜索订单号/支付单号"
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
        <template #orderNo-cell="{ getValue }">
          <code class="text-xs">{{ getValue() }}</code>
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
              v-else-if="row.original.status === 'pending_payment' || row.original.status === 'paying'"
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

    <!-- 详情 Modal -->
    <UModal
      v-model:open="detailOpen"
      :title="current ? `支付单 ${current.paymentNo}` : '支付单详情'"
      description="查看支付详情、退款记录与风控/分账信息"
      :ui="{ content: 'max-w-3xl w-full' }"
      :prevent-close="detailLoading"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm text-muted">支付单号</div>
              <div class="font-semibold">{{ current?.paymentNo }}</div>
            </div>
            <UBadge :color="current ? statusColor(current.status) : 'neutral'">
              {{ current ? statusLabel(current.status) : "" }}
            </UBadge>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-muted">订单号</div>
              <div class="font-medium">{{ current?.orderNo }}</div>
            </div>
            <div>
              <div class="text-xs text-muted">渠道</div>
              <div class="font-medium">{{ current?.providerName }}</div>
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
            <div v-if="current?.failureReason" class="col-span-2">
              <div class="text-xs text-muted">失败原因</div>
              <div class="font-medium text-error-600">
                {{ current?.failureReason }}
              </div>
            </div>
          </div>

          <USeparator />

          <div>
            <div class="text-sm font-semibold mb-2">退款记录</div>
            <div v-if="refundsForCurrent.length" class="space-y-2">
              <div
                v-for="refund in refundsForCurrent"
                :key="refund.id"
                class="flex items-center justify-between text-sm"
              >
                <div>
                  <div class="font-medium">{{ refund.refundNo }}</div>
                  <div class="text-xs text-muted">
                    {{ refund.createdAt ? fmtDT(refund.createdAt) : "-" }}
                  </div>
                </div>
                <div class="text-right">
                  <div class="font-medium">{{ formatCNY(refund.refundAmount) }}</div>
                  <div class="text-xs text-muted">{{ refund.status }}</div>
                </div>
              </div>
            </div>
            <div v-else class="text-sm text-muted">暂无退款记录</div>
          </div>

          <USeparator />

          <div>
            <div class="text-sm font-semibold mb-2">风险事件</div>
            <div v-if="detailLoading" class="text-sm text-muted">加载中...</div>
            <div v-else-if="riskEvents.length" class="space-y-2">
              <div
                v-for="event in riskEvents"
                :key="event.id"
                class="flex items-center justify-between text-sm"
              >
                <div>
                  <div class="font-medium">{{ event.riskType }}</div>
                  <div class="text-xs text-muted">
                    {{ event.createdAt ? fmtDT(event.createdAt) : "-" }}
                  </div>
                </div>
                <div class="text-right">
                  <div class="font-medium">评分 {{ event.riskScore }}</div>
                  <div class="text-xs text-muted">{{ event.action }}</div>
                </div>
              </div>
            </div>
            <div v-else class="text-sm text-muted">暂无风险事件</div>
          </div>

          <USeparator />

          <div>
            <div class="text-sm font-semibold mb-2">分账结果</div>
            <div v-if="detailLoading" class="text-sm text-muted">加载中...</div>
            <div v-else-if="splitResults.length" class="space-y-2">
              <div
                v-for="result in splitResults"
                :key="result.id"
                class="flex items-center justify-between text-sm"
              >
                <div>
                  <div class="font-medium">{{ result.participant }}</div>
                  <div class="text-xs text-muted">
                    {{ result.createdAt ? fmtDT(result.createdAt) : "-" }}
                  </div>
                </div>
                <div class="text-right">
                  <div class="font-medium">{{ formatCNY(result.amount) }}</div>
                  <div class="text-xs text-muted">{{ result.status }}</div>
                </div>
              </div>
            </div>
            <div v-else class="text-sm text-muted">暂无分账结果</div>
          </div>

          <USeparator />

          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm font-semibold">对账处理</div>
              <div class="text-xs text-muted">记录该笔支付的对账差异</div>
            </div>
            <UButton
              size="xs"
              variant="outline"
              icon="i-heroicons-adjustments-horizontal"
              @click="openReconcile"
            >
              记录对账
            </UButton>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" @click="closeDetail">
            关闭
          </UButton>
          <UButton
            v-if="current?.status === 'paid'"
            color="error"
            type="button"
            icon="i-heroicons-arrow-uturn-left"
            @click="requestRefund(current!)"
          >
            退款
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="reconcileOpen"
      title="对账处理"
      description="创建对账记录并提交差异说明"
      :prevent-close="reconcileSaving"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField label="对账周期" required>
              <USelect
                v-model="reconcileForm.periodType"
                :items="reconcilePeriodItems"
                class="w-full"
              />
            </UFormField>
            <UFormField label="差异类型" required>
              <USelect
                v-model="reconcileForm.diffType"
                :items="reconcileDiffTypeItems"
                class="w-full"
              />
            </UFormField>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <UFormField label="周期开始" required>
              <UInput v-model="reconcileForm.periodStart" type="date" />
            </UFormField>
            <UFormField label="周期结束" required>
              <UInput v-model="reconcileForm.periodEnd" type="date" />
            </UFormField>
          </div>
          <UFormField label="差异金额（分）" required>
            <UInput v-model.number="reconcileForm.diffAmount" type="number" />
          </UFormField>
          <UFormField label="处理说明">
            <UTextarea v-model.trim="reconcileForm.resolution" :rows="3" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="reconcileSaving" @click="reconcileOpen = false">
            取消
          </UButton>
          <UButton color="primary" type="button" :loading="reconcileSaving" @click="submitReconcile">
            提交对账
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { usePaymentsApi } from "~/composables/api";
import { useToastAlert } from "~/composables/useToastAlert";
import type {
  PaymentProvider,
  PaymentTransaction,
  PaymentRefund,
  PaymentRiskEvent,
  PaymentSplitResult,
} from "~/types/payments";

const { t } = useI18n();
const { listProviders, listTransactions, createRefund, listRiskEvents, listSplitResults, createReconciliation } =
  usePaymentsApi();
const toast = useToastAlert();

type PaymentRow = {
  id: number;
  paymentNo: string;
  orderNo: string;
  providerId: number;
  providerName: string;
  method: string;
  amount: number;
  fee: number;
  status: string;
  createdAt: string;
  paidAt?: string | null;
  failureReason?: string;
};

const loading = ref(false);
const detailLoading = ref(false);
const keyword = ref("");
const ALL_FILTER = "all";
const selectedStatus = ref<string>(ALL_FILTER);
const selectedMethod = ref<string>(ALL_FILTER);
const providers = ref<PaymentProvider[]>([]);
const transactions = ref<PaymentTransaction[]>([]);
const refundsByTransaction = ref<Record<string, PaymentRefund[]>>({});
const riskEvents = ref<PaymentRiskEvent[]>([]);
const splitResults = ref<PaymentSplitResult[]>([]);
const reconcileOpen = ref(false);
const reconcileSaving = ref(false);
const reconcileForm = reactive({
  periodType: "daily",
  periodStart: "",
  periodEnd: "",
  diffType: "",
  diffAmount: 0,
  resolution: "",
});

const reconcilePeriodItems = [
  { label: "日对账", value: "daily" },
  { label: "周对账", value: "weekly" },
];
const reconcileDiffTypeItems = [
  { label: "金额不一致", value: "amount_mismatch" },
  { label: "缺失记录", value: "missing" },
  { label: "重复记录", value: "duplicate" },
  { label: "其他", value: "other" },
];

// 选项
const statusItems = [
  { label: "全部状态", value: ALL_FILTER },
  { label: "待支付", value: "pending_payment" },
  { label: "支付中", value: "paying" },
  { label: "已支付", value: "paid" },
  { label: "失败", value: "failed" },
  { label: "已取消", value: "canceled" },
  { label: "超时", value: "timeout" },
  { label: "已退款", value: "refunded" },
];
const methodItems = [
  { label: "全部方式", value: ALL_FILTER },
  { label: "微信支付", value: "wechat" },
  { label: "微信 JSAPI", value: "wechat_jsapi" },
  { label: "微信小程序", value: "wechat_miniapp" },
  { label: "支付宝", value: "alipay" },
  { label: "银行卡", value: "card" },
  { label: "银行转账", value: "bank" },
  { label: "货到付款", value: "cod" },
];

// 列（v3 TanStack 风格）
const columns = computed<TableColumn<PaymentRow>[]>(() => [
  { accessorKey: "paymentNo", header: "支付单号" },
  { accessorKey: "orderNo", header: "订单号" },
  { accessorKey: "providerName", header: "渠道" },
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

const providerMap = computed(() => new Map(providers.value.map((p) => [p.id, p])));

const rows = computed<PaymentRow[]>(() =>
  transactions.value.map((tx) => ({
    id: tx.id,
    paymentNo: tx.transactionNo,
    orderNo: tx.orderNo || tx.orderId,
    providerId: tx.providerId,
    providerName: providerMap.value.get(tx.providerId)?.name || (tx.providerId ? `渠道#${tx.providerId}` : "-"),
    method: tx.payMethod,
    amount: tx.amountTotal,
    fee: tx.feeAmount,
    status: tx.status,
    createdAt: tx.createdAt,
    paidAt: tx.completedAt || null,
    failureReason: tx.failureReason,
  })),
);

// 过滤
const filteredRows = computed(() => {
  const q = keyword.value.trim().toLowerCase();
  return rows.value.filter((r) => {
    const okQ =
      !q ||
      r.paymentNo.toLowerCase().includes(q) ||
      r.orderNo.toLowerCase().includes(q);
    const okS =
      selectedStatus.value === ALL_FILTER || r.status === selectedStatus.value;
    const okM =
      selectedMethod.value === ALL_FILTER || r.method === selectedMethod.value;
    return okQ && okS && okM;
  });
});

// 汇总
const totalPaid = computed(() =>
  filteredRows.value
    .filter((r) => r.status === "paid" || r.status === "refunded")
    .reduce((s, r) => s + r.amount, 0)
);
const totalRefund = computed(() =>
  filteredRows.value
    .filter((r) => r.status === "refunded")
    .reduce((s, r) => s + r.amount, 0)
);
const totalPending = computed(() =>
  filteredRows.value
    .filter((r) => r.status === "pending_payment" || r.status === "paying")
    .reduce((s, r) => s + r.amount, 0)
);

// 工具
const formatCNY = (minor: number) => {
  const amount = Number(minor || 0) / 100;
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "CNY",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount);
};
const fmtDT = (s: string) =>
  s ? new Date(s).toLocaleString("zh-CN", { hour12: false }) : "-";

function statusLabel(s: PaymentRow["status"]) {
  return (
    (
      {
        pending_payment: "待支付",
        paying: "支付中",
        paid: "已支付",
        failed: "失败",
        canceled: "已取消",
        timeout: "超时",
        refunded: "已退款",
      } as const
    )[s] || s
  );
}
function statusColor(s: PaymentRow["status"]) {
  return (
    (
      {
        pending_payment: "warning",
        paying: "info",
        paid: "success",
        failed: "error",
        canceled: "neutral",
        timeout: "neutral",
        refunded: "neutral",
      } as const
    )[s] || "neutral"
  );
}
function methodLabel(m: PaymentRow["method"]) {
  return (
    (
      {
        wechat: "微信支付",
        wechat_jsapi: "微信 JSAPI",
        wechat_miniapp: "微信小程序",
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
        wechat_jsapi: "i-simple-icons-wechat",
        wechat_miniapp: "i-simple-icons-wechat",
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
  selectedStatus.value = ALL_FILTER;
  selectedMethod.value = ALL_FILTER;
}

function exportCsv() {
  const header = [
    "PaymentNo",
    "OrderNo",
    "Channel",
    "Method",
    "Amount",
    "Status",
    "CreatedAt",
    "PaidAt",
  ];
  const body = filteredRows.value.map((r) => [
    r.paymentNo,
    r.orderNo,
    r.providerName,
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

// 明细/退款/重试
const detailOpen = ref(false);
const current = ref<PaymentRow | null>(null);
function openDetail(row: PaymentRow) {
  current.value = row;
  detailOpen.value = true;
  loadDetailExtras(row);
}
const blurActiveElement = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
};
const closeDetail = () => {
  blurActiveElement();
  detailOpen.value = false;
};
function requestRefund(row: PaymentRow) {
  const amountMinor = Math.max(1, Math.round(row.amount || 0));
  const reason =
    typeof window !== "undefined"
      ? window.prompt("请输入退款原因（可选）", "后台退款")
      : "后台退款";
  createRefund(row.id, { amountMinor, reason: reason || undefined })
    .then((refund) => {
      const key = String(row.id);
      const existing = refundsByTransaction.value[key] || [];
      refundsByTransaction.value = {
        ...refundsByTransaction.value,
        [key]: [refund, ...existing],
      };
      toast.add({ title: `已提交退款申请：${row.paymentNo}`, color: "info" });
    })
    .catch((error: any) => {
      toast.add({
        title: "退款申请失败",
        description: error?.message || "请稍后重试",
        color: "red",
      });
    });
}
function retryPay(row: PaymentRow) {
  toast.add({ title: `已触发重试：${row.paymentNo}`, color: "primary" });
}

const openReconcile = () => {
  if (!current.value) return;
  const today = new Date().toISOString().slice(0, 10);
  if (!reconcileForm.periodStart) reconcileForm.periodStart = today;
  if (!reconcileForm.periodEnd) reconcileForm.periodEnd = today;
  if (!reconcileForm.diffType) reconcileForm.diffType = "amount_mismatch";
  reconcileOpen.value = true;
};

const submitReconcile = async () => {
  if (!current.value) return;
  const periodStart = reconcileForm.periodStart.trim();
  const periodEnd = reconcileForm.periodEnd.trim();
  const diffType = reconcileForm.diffType.trim();
  if (!periodStart || !periodEnd || !diffType) {
    toast.add({ title: "请补全对账信息", color: "error" });
    return;
  }
  reconcileSaving.value = true;
  try {
    await createReconciliation({
      periodType: reconcileForm.periodType,
      periodStart: new Date(`${periodStart}T00:00:00Z`).toISOString(),
      periodEnd: new Date(`${periodEnd}T23:59:59Z`).toISOString(),
      items: [
        {
          transactionId: current.value.id,
          diffType,
          diffAmount: Number(reconcileForm.diffAmount || 0),
        },
      ],
    });
    toast.add({ title: "对账记录已提交", color: "success" });
    reconcileOpen.value = false;
  } catch (error: any) {
    toast.add({
      title: "对账提交失败",
      description: error?.message || "请稍后重试",
      color: "red",
    });
  } finally {
    reconcileSaving.value = false;
  }
};

const refundsForCurrent = computed(() => {
  const id = current.value?.id;
  if (!id) return [];
  return refundsByTransaction.value[String(id)] || [];
});

watch(detailOpen, (open) => {
  if (!open) {
    blurActiveElement();
  }
});

const loadDetailExtras = async (row: PaymentRow) => {
  detailLoading.value = true;
  try {
    const [risk, splits] = await Promise.all([
      listRiskEvents(row.id),
      listSplitResults(row.id),
    ]);
    riskEvents.value = risk;
    splitResults.value = splits;
  } catch (error: any) {
    toast.add({
      title: "加载支付详情失败",
      description: error?.message || "请稍后重试",
      color: "red",
    });
  } finally {
    detailLoading.value = false;
  }
};

const loadData = async () => {
  loading.value = true;
  try {
    const [providerRows, transactionRows] = await Promise.all([
      listProviders(),
      listTransactions(),
    ]);
    providers.value = providerRows;
    transactions.value = transactionRows;
  } catch (error: any) {
    toast.add({
      title: "获取支付列表失败",
      description: error?.message || "请稍后重试",
      color: "red",
    });
  } finally {
    loading.value = false;
  }
};

onMounted(loadData);
</script>
