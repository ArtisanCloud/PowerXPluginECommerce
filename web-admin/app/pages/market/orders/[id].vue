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
          v-if="!isEditMode && canEdit"
          color="neutral"
          variant="soft"
          @click="enterEdit"
        >
          编辑
        </UButton>
        <UButton
          v-else-if="isEditMode"
          color="neutral"
          variant="soft"
          @click="exitEdit"
        >
          取消编辑
        </UButton>
        <UButton
          v-if="canManualPay"
          color="primary"
          variant="soft"
          @click="openManualPayment"
        >
          手动收款
        </UButton>
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

    <UAlert
      v-if="isEditMode && !canEdit"
      color="warning"
      variant="soft"
      title="当前状态不可编辑"
      description="仅待支付或草稿订单支持编辑收货信息。"
    />

    <UCard v-if="detail?.summary" :ui="{ body: { base: 'space-y-3' } }">
      <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
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

    <UCard title="收货信息">
      <div v-if="isEditMode" class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <UFormField label="收货人" required>
          <UInput v-model.trim="editForm.recipientName" placeholder="请输入收货人姓名" />
        </UFormField>
        <UFormField label="联系电话" required>
          <UInput v-model.trim="editForm.recipientPhone" placeholder="请输入联系电话" />
        </UFormField>
        <UFormField label="国家/地区">
          <UInput v-model.trim="editForm.countryCode" placeholder="CN" />
        </UFormField>
        <UFormField label="邮编">
          <UInput v-model.trim="editForm.postalCode" placeholder="请输入邮编" />
        </UFormField>
        <UFormField label="省/州">
          <UInput v-model.trim="editForm.province" placeholder="请输入省/州" />
        </UFormField>
        <UFormField label="市">
          <UInput v-model.trim="editForm.city" placeholder="请输入城市" />
        </UFormField>
        <UFormField label="区/县">
          <UInput v-model.trim="editForm.district" placeholder="请输入区/县" />
        </UFormField>
        <UFormField label="地址 1" required>
          <UInput v-model.trim="editForm.address1" placeholder="街道/门牌号" />
        </UFormField>
        <UFormField label="地址 2">
          <UInput v-model.trim="editForm.address2" placeholder="楼层/房间号" />
        </UFormField>
        <UFormField label="标签">
          <UInput v-model.trim="editForm.label" placeholder="例如：公司/家" />
        </UFormField>
      </div>
      <div v-else class="grid grid-cols-1 gap-3 md:grid-cols-2">
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">收货人</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.recipientName || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">联系电话</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.recipientPhone || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">国家/地区</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.countryCode || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">邮编</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.postalCode || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">省/州</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.province || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">市</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.city || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">区/县</div>
          <div class="text-sm">
            {{ detail?.summary?.shippingAddressSnapshot?.district || "-" }}
          </div>
        </div>
        <div>
          <div class="text-xs text-gray-500 dark:text-gray-400">地址</div>
          <div class="text-sm">
            {{ shippingAddressLine }}
          </div>
        </div>
      </div>
      <template v-if="isEditMode" #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="updating" @click="resetEditForm">
            重置
          </UButton>
          <UButton color="primary" type="button" :loading="updating" :disabled="!canEdit" @click="saveShippingAddress">
            保存收货信息
          </UButton>
        </div>
      </template>
    </UCard>

    <UCard title="订单明细">
      <UTable :columns="itemColumns" :data="detail?.items || []" :loading="loading">
        <template #skuId-cell="{ row }">
          <div class="text-sm text-gray-900 dark:text-white">
            {{ skuDisplayMap[row.original.skuId] || row.original.skuId }}
          </div>
        </template>
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

    <UCard title="优惠券与礼品卡">
      <div class="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div class="text-sm text-gray-500 dark:text-gray-400">
          申请后需审核通过，当前不会自动修改订单金额。
        </div>
        <div class="flex items-center gap-2">
          <UButton color="primary" variant="soft" @click="openBenefitModal('coupon')">
            添加优惠券
          </UButton>
          <UButton color="primary" variant="soft" @click="openBenefitModal('giftcard')">
            添加礼品卡
          </UButton>
          <UButton
            color="success"
            variant="soft"
            :disabled="!hasBenefitSelection"
            :loading="benefitSubmitting"
            @click="approveSelectedBenefits"
          >
            批量通过
          </UButton>
          <UButton
            color="warning"
            variant="soft"
            :disabled="!hasBenefitSelection"
            @click="openBenefitReject"
          >
            批量拒绝
          </UButton>
        </div>
      </div>
      <UTable :columns="benefitColumns" :data="benefitReviews" :loading="benefitLoading">
        <template #select-header>
          <UCheckbox :model-value="benefitAllSelected" @change="toggleBenefitSelectAll" />
        </template>
        <template #select-cell="{ row }">
          <UCheckbox
            :model-value="benefitSelection.has(row.original.id)"
            :disabled="row.original.status !== 'pending_review'"
            @change="() => toggleBenefitRow(row.original)"
          />
        </template>
        <template #benefitType-cell="{ row }">
          <span class="text-sm">
            {{ benefitTypeLabel(row.original.benefitType) }}
          </span>
        </template>
        <template #benefitCode-cell="{ row }">
          <span class="text-sm font-medium">
            {{ row.original.benefitCode }}
          </span>
        </template>
        <template #value-cell="{ row }">
          <span class="text-sm">
            {{ benefitValueLabel(row.original) }}
          </span>
        </template>
        <template #amountMinor-cell="{ row }">
          <div class="text-right tabular-nums">
            {{ formatMoney(row.original.currency || detail?.summary?.amounts?.currency || "CNY", -row.original.amountMinor) }}
          </div>
        </template>
        <template #submittedBy-cell="{ row }">
          <span class="text-sm">{{ resolveAdminName(row.original.submittedBy) }}</span>
        </template>
        <template #reviewedBy-cell="{ row }">
          <span class="text-sm">{{ resolveAdminName(row.original.reviewedBy) }}</span>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="benefitStatusColor(row.original.status)" variant="subtle">
            {{ benefitStatusLabel(row.original.status) }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center justify-end gap-2">
            <UButton
              v-if="row.original.status === 'pending_review'"
              size="xs"
              color="success"
              variant="soft"
              :disabled="benefitSubmitting"
              @click="approveBenefit(row.original)"
            >
              通过
            </UButton>
            <UButton
              v-if="row.original.status === 'pending_review'"
              size="xs"
              color="warning"
              variant="soft"
              :disabled="benefitSubmitting"
              @click="openBenefitReject([row.original.id])"
            >
              拒绝
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard title="支付交易记录">
      <UTable :columns="paymentColumns" :data="paymentTransactions" :loading="paymentLoading">
        <template #transactionNo-cell="{ row }">
          <code class="text-xs">{{ row.original.transactionNo }}</code>
        </template>
        <template #amountTotal-cell="{ row }">
          <div class="text-right tabular-nums">
            {{ formatMoney(row.original.amountCurrency || detail?.summary?.amounts?.currency || "CNY", row.original.amountTotal) }}
          </div>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="paymentStatusColor(row.original.status)" variant="subtle">
            {{ paymentStatusLabel(row.original.status) }}
          </UBadge>
        </template>
        <template #createdAt-cell="{ row }">
          {{ formatTime(row.original.createdAt) }}
        </template>
        <template #completedAt-cell="{ row }">
          {{ row.original.completedAt ? formatTime(row.original.completedAt) : "-" }}
        </template>
      </UTable>
    </UCard>

    <UCard title="手动收款记录">
      <UTable :columns="manualColumns" :data="manualLogs" :loading="manualLoading">
        <template #amountMinor-cell="{ row }">
          <div class="text-right tabular-nums">
            {{ formatMoney(row.original.currency || detail?.summary?.amounts?.currency || "CNY", row.original.amountMinor) }}
          </div>
        </template>
        <template #submittedBy-cell="{ row }">
          <span class="text-sm">{{ resolveAdminName(row.original.submittedBy) }}</span>
        </template>
        <template #reviewedBy-cell="{ row }">
          <span class="text-sm">{{ resolveAdminName(row.original.reviewedBy) }}</span>
        </template>
        <template #status-cell="{ row }">
          <UBadge :color="manualRowStatusColor(row.original)" variant="subtle">
            {{ manualRowStatusLabel(row.original) }}
          </UBadge>
        </template>
        <template #actions-cell="{ row }">
          <div class="flex items-center justify-end gap-2">
            <UButton
              v-if="canActOnManualLog(row.original)"
              size="xs"
              color="success"
              variant="soft"
              @click="approveManual(row.original.reviewId)"
            >
              通过
            </UButton>
            <UButton
              v-if="canActOnManualLog(row.original)"
              size="xs"
              color="warning"
              variant="soft"
              @click="openReject(row.original.reviewId)"
            >
              拒绝
            </UButton>
          </div>
        </template>
      </UTable>
    </UCard>

    <UCard title="事件日志">
      <div class="mb-3 text-sm text-gray-500 dark:text-gray-400">
        订单全量事件记录（包含创建、收款、审核等）。
      </div>
      <UTable :columns="eventColumns" :data="detail?.events || []" :loading="loading">
        <template #operator-cell="{ row }">
          <span class="text-sm">
            {{ formatOperatorLabel(row.original) }}
          </span>
        </template>
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

    <UModal
      v-model:open="manualOpen"
      title="手动收款"
      description="提交后需要另一位管理员审核通过，订单才会变为已支付。"
      :prevent-close="manualSubmitting"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <UFormField label="支付方式">
            <USelect v-model="manualForm.payMethod" :items="manualMethodItems" class="w-full" />
          </UFormField>
          <UFormField label="收款金额">
            <UInput :model-value="manualAmountLabel" readonly />
          </UFormField>
          <UFormField label="凭证号（可选）">
            <UInput v-model.trim="manualForm.proofNo" placeholder="银行流水号/收据编号" />
          </UFormField>
          <UFormField label="备注（可选）">
            <UTextarea v-model.trim="manualForm.note" :rows="3" placeholder="补充说明" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="manualSubmitting" @click="closeManualPayment">
            返回
          </UButton>
          <UButton color="primary" type="button" :loading="manualSubmitting" @click="submitManualPayment">
            提交审核
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="benefitOpen"
      title="添加优惠权益"
      description="提交后需审核，通过后才会生效。"
      :prevent-close="benefitSubmitting"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <UFormField label="类型">
            <USelect
              v-model="benefitForm.type"
              :items="[{ label: '优惠券', value: 'coupon' }, { label: '礼品卡', value: 'giftcard' }]"
              class="w-full"
            />
          </UFormField>
          <UFormField label="券码/卡号" required>
            <USelectMenu
              v-model="benefitForm.code"
              v-model:search-term="benefitSearchTerm"
              :items="benefitCodeOptions"
              value-key="value"
              label-key="label"
              :loading="benefitCodeLoading"
              :portal="false"
              searchable
              :ignore-filter="true"
              class="w-full"
              placeholder="输入券码/卡号搜索"
              @update:search-term="loadBenefitCodes"
            />
            <p v-if="benefitCodeHint" class="mt-1 text-xs text-amber-500">
              {{ benefitCodeHint }}
            </p>
          </UFormField>
          <UFormField v-if="benefitForm.type === 'coupon'" label="优惠类型" required>
            <USelect
              v-model="benefitForm.valueType"
              :items="[{ label: '固定金额', value: 'amount' }, { label: '折扣百分比', value: 'percent' }]"
              class="w-full"
            />
          </UFormField>
          <UFormField v-else label="抵扣类型">
            <UInput value="余额抵扣" readonly />
          </UFormField>
          <UFormField label="优惠值" required>
            <UInput
              v-model.trim="benefitForm.value"
              type="number"
              min="0"
              step="0.01"
              :placeholder="benefitForm.valueType === 'percent' ? '例如：10（%）' : '例如：100（元）'"
            />
          </UFormField>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            预计抵扣：{{ formatMoney(detail?.summary?.amounts?.currency || "CNY", -benefitAmountMinor) }}
          </div>
          <UFormField label="允许叠加">
            <USwitch v-model="benefitForm.stackingAllowed" />
          </UFormField>
          <UFormField label="备注（可选）">
            <UTextarea v-model.trim="benefitForm.note" :rows="3" placeholder="补充说明" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="benefitSubmitting" @click="closeBenefitModal">
            返回
          </UButton>
          <UButton color="primary" type="button" :loading="benefitSubmitting" @click="submitBenefitReview">
            提交审核
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="benefitRejectOpen"
      title="拒绝优惠申请"
      description="请填写拒绝原因。"
      :prevent-close="benefitSubmitting"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <UFormField label="拒绝原因" required>
            <UTextarea v-model.trim="benefitRejectReason" :rows="3" placeholder="例如：不符合使用规则" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="benefitSubmitting" @click="closeBenefitReject">
            返回
          </UButton>
          <UButton color="warning" type="button" :loading="benefitSubmitting" @click="confirmBenefitReject">
            确认拒绝
          </UButton>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="rejectOpen"
      title="拒绝收款申请"
      description="请填写拒绝原因。"
      :prevent-close="manualSubmitting"
    >
      <template #body>
        <div class="space-y-4 p-4 sm:p-5">
          <UFormField label="拒绝原因">
            <UTextarea v-model.trim="rejectReason" :rows="3" placeholder="例如：金额不一致" />
          </UFormField>
        </div>
      </template>
      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 p-4 sm:p-5 sm:flex-row sm:justify-end">
          <UButton color="neutral" variant="subtle" type="button" :disabled="manualSubmitting" @click="closeReject">
            返回
          </UButton>
          <UButton color="warning" type="button" :loading="manualSubmitting" @click="confirmReject">
            确认拒绝
          </UButton>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { TableColumn } from "@nuxt/ui";
import { useOrderApi } from "~/composables/api/useOrder";
import { useSkuApi } from "~/composables/api/useSku";
import { usePaymentsApi } from "~/composables/api/usePayments";
import { useAuthService } from "~/composables/api/services/authService";
import type { ManualPaymentReview, ManualPaymentReviewLog, PaymentTransaction } from "~/types/payments";
import type { OrderBenefitReview, OrderDetail, OrderEvent, OrderItem, ShippingAddress } from "~/types/order";

definePageMeta({
  name: "order-detail",
});

const route = useRoute();
const router = useRouter();
const toast = useToastAlert();
const api = useOrderApi();
const skuApi = useSkuApi();
const paymentsApi = usePaymentsApi();
const authService = useAuthService();

const loading = ref(false);
const cancelling = ref(false);
const detail = ref<OrderDetail | null>(null);
const skuDisplayMap = ref<Record<string, string>>({});
const operatorMap = ref<Record<string, string>>({});

const cancelOpen = ref(false);
const cancelReason = ref("");
const manualOpen = ref(false);
const rejectOpen = ref(false);
const manualSubmitting = ref(false);
const manualLoading = ref(false);
const manualReviews = ref<ManualPaymentReview[]>([]);
const manualLogs = ref<ManualPaymentReviewLog[]>([]);
const paymentTransactions = ref<PaymentTransaction[]>([]);
const paymentLoading = ref(false);
const currentReviewId = ref<number | null>(null);
const rejectReason = ref("");
const updating = ref(false);
const manualForm = reactive({
  payMethod: "bank_transfer",
  proofNo: "",
  note: "",
});
const benefitOpen = ref(false);
const benefitRejectOpen = ref(false);
const benefitSubmitting = ref(false);
const benefitLoading = ref(false);
const benefitSearchTerm = ref("");
const benefitCodeLoading = ref(false);
const benefitCodeOptions = ref<Array<{ label: string; value: string }>>([]);
const benefitCodeHint = ref("");
const benefitReviews = ref<OrderBenefitReview[]>([]);
const benefitSelection = ref<Set<number>>(new Set());
const benefitRejectReason = ref("");
const benefitForm = reactive({
  type: "coupon",
  code: "",
  valueType: "amount",
  value: "",
  stackingAllowed: false,
  note: "",
});

const id = computed(() => String(route.params.id || ""));
const isEditMode = computed(() => route.query.mode === "edit");
const canEdit = computed(() => {
  const status = detail.value?.summary?.status || "";
  return status === "pending_payment" || status === "draft";
});

const emptyAddress = (): ShippingAddress => ({
  label: "",
  recipientName: "",
  recipientPhone: "",
  countryCode: "",
  province: "",
  city: "",
  district: "",
  address1: "",
  address2: "",
  postalCode: "",
});

const editForm = reactive<ShippingAddress>(emptyAddress());

const itemColumns = computed<TableColumn<OrderItem>[]>(() => [
  { accessorKey: "skuId", header: "SKU", meta: { class: { td: "text-left" } } },
  { accessorKey: "qty", header: "数量", meta: { class: { td: "text-center" } } },
  { accessorKey: "unitPrice", header: "单价", meta: { class: { td: "text-right tabular-nums" } } },
  { accessorKey: "lineAmount", header: "行金额", meta: { class: { td: "text-right tabular-nums" } } },
]);

const eventColumns = computed<TableColumn<OrderEvent>[]>(() => [
  { accessorKey: "eventType", header: "事件" },
  { accessorKey: "operatorType", header: "操作者类型" },
  { accessorKey: "operator", header: "操作者" },
  { accessorKey: "createdAt", header: "时间" },
]);

const manualColumns = computed<TableColumn<ManualPaymentReviewLog>[]>(() => [
  { accessorKey: "id", header: "编号" },
  { accessorKey: "payMethod", header: "支付方式" },
  { accessorKey: "amountMinor", header: "金额", meta: { class: { td: "text-right" } } },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "submittedBy", header: "提交人" },
  { accessorKey: "submittedAt", header: "提交时间" },
  { accessorKey: "reviewedBy", header: "审核人" },
  { accessorKey: "reviewedAt", header: "审核时间" },
  { id: "actions", header: "操作", meta: { class: { td: "text-right" } } },
]);

const paymentColumns = computed<TableColumn<PaymentTransaction>[]>(() => [
  { accessorKey: "transactionNo", header: "支付单号" },
  { accessorKey: "payMethod", header: "支付方式" },
  { accessorKey: "amountTotal", header: "金额", meta: { class: { td: "text-right" } } },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "createdAt", header: "创建时间" },
  { accessorKey: "completedAt", header: "完成时间" },
]);

const benefitColumns = computed<TableColumn<OrderBenefitReview>[]>(() => [
  { id: "select", header: "", meta: { class: { td: "w-10" } } },
  { accessorKey: "benefitType", header: "类型" },
  { accessorKey: "benefitCode", header: "券码/卡号" },
  { accessorKey: "value", header: "面值/折扣" },
  { accessorKey: "amountMinor", header: "抵扣金额", meta: { class: { td: "text-right" } } },
  { accessorKey: "status", header: "状态" },
  { accessorKey: "submittedBy", header: "提交人" },
  { accessorKey: "submittedAt", header: "提交时间" },
  { accessorKey: "reviewedBy", header: "审核人" },
  { accessorKey: "reviewedAt", header: "审核时间" },
  { id: "actions", header: "操作", meta: { class: { td: "text-right" } } },
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

const paymentStatusLabel = (st: string) => {
  const map: Record<string, string> = {
    pending_payment: "待支付",
    paying: "支付中",
    paid: "已支付",
    failed: "失败",
    cancelled: "已取消",
  };
  return map[st] || st || "-";
};

const paymentStatusColor = (st: string) => {
  const map: Record<string, "warning" | "success" | "neutral" | "info" | "error" | "primary"> = {
    pending_payment: "warning",
    paying: "info",
    paid: "success",
    failed: "error",
    cancelled: "neutral",
  };
  return map[st] || "neutral";
};

const canCancel = computed(() => detail.value?.summary?.status === "pending_payment");
const canManualPay = computed(() => detail.value?.summary?.status === "pending_payment");

const manualMethodItems = [
  { label: "银行转账", value: "bank_transfer" },
  { label: "现金", value: "cash" },
  { label: "POS", value: "pos" },
  { label: "其他", value: "other" },
];

const manualAmountLabel = computed(() => {
  if (!detail.value?.summary) return "-";
  return formatMoney(detail.value.summary.amounts.currency, detail.value.summary.amounts.total);
});

const benefitAmountMinor = computed(() => {
  if (!detail.value?.summary) return 0;
  const raw = Number(benefitForm.value || 0);
  if (!Number.isFinite(raw) || raw <= 0) return 0;
  const total = detail.value.summary.amounts.total;
  if (benefitForm.valueType === "percent") {
    return Math.round((total * raw) / 100);
  }
  return Math.round(raw * 100);
});

const hasBenefitSelection = computed(() => benefitSelection.value.size > 0);
const benefitSelectableIds = computed(() =>
  benefitReviews.value.filter((row) => row.status === "pending_review").map((row) => row.id),
);
const benefitAllSelected = computed(() => {
  const ids = benefitSelectableIds.value;
  if (!ids.length) return false;
  return ids.every((id) => benefitSelection.value.has(id));
});

const shippingAddressLine = computed(() => {
  const snap = detail.value?.summary?.shippingAddressSnapshot;
  if (!snap) return "-";
  const region = [snap.province, snap.city, snap.district].filter(Boolean).join("");
  const address = [region, snap.address1, snap.address2].filter(Boolean).join(" ");
  return address || "-";
});

const operatorTypeLabel = (raw?: string) => {
  const type = String(raw || "").trim();
  const map: Record<string, string> = {
    admin: "管理员",
    customer: "客户",
    system: "系统",
  };
  return map[type] || type || "-";
};

const formatOperatorLabel = (event: OrderEvent) => {
  const id = String(event.operator || "").trim();
  const name = operatorMap.value[id];
  if (name) return name;
  if (!id) return operatorTypeLabel(event.operatorType);
  const prefix = operatorTypeLabel(event.operatorType);
  return prefix && prefix !== "-" ? `${prefix}#${id}` : id;
};

const resolveAdminName = (raw?: string) => {
  const id = String(raw || "").trim();
  if (!id) return "-";
  return operatorMap.value[id] || `管理员#${id}`;
};

const loadOperatorNames = async (ids: string[]) => {
  const unique = Array.from(new Set(ids.map((id) => String(id || "").trim()).filter(Boolean)));
  const pending = unique.filter((id) => !operatorMap.value[id]);
  if (!pending.length) return;
  const adminIds = pending.filter((id) => /^\d+$/.test(id));
  if (!adminIds.length) return;
  const results = await Promise.allSettled(adminIds.map((id) => authService.getUser(id)));
  const next = { ...operatorMap.value };
  results.forEach((res, idx) => {
    const id = adminIds[idx];
    if (res.status === "fulfilled") {
      const user = res.value?.data;
      const name = user?.display_name || user?.email || user?.phone || user?.id || id;
      next[id] = name;
    }
  });
  operatorMap.value = next;
};

const syncOperatorMap = async () => {
  const ids = new Set<string>();
  (detail.value?.events || []).forEach((event) => {
    if (event.operatorType === "admin") {
      const id = String(event.operator || "").trim();
      if (id) ids.add(id);
    }
  });
  (manualReviews.value || []).forEach((row) => {
    if (row.submittedBy) ids.add(String(row.submittedBy));
    if (row.reviewedBy) ids.add(String(row.reviewedBy));
  });
  (manualLogs.value || []).forEach((row) => {
    if (row.submittedBy) ids.add(String(row.submittedBy));
    if (row.reviewedBy) ids.add(String(row.reviewedBy));
  });
  (benefitReviews.value || []).forEach((row) => {
    if (row.submittedBy) ids.add(String(row.submittedBy));
    if (row.reviewedBy) ids.add(String(row.reviewedBy));
  });
  await loadOperatorNames(Array.from(ids));
};

const buildSkuLabel = (item: { spuName?: string; skuCode?: string; specDisplay?: string; id?: string }) => {
  const spuName = String(item.spuName || "").trim();
  const skuCode = String(item.skuCode || "").trim();
  const specDisplay = String(item.specDisplay || "").trim();
  if (spuName && skuCode) return `${spuName} · ${skuCode}`;
  if (spuName && specDisplay) return `${spuName} · ${specDisplay}`;
  if (skuCode) return skuCode;
  if (specDisplay) return specDisplay;
  return item.id || "-";
};

const loadSkuMetaByIds = async (ids: string[]) => {
  const uniqueIds = Array.from(new Set(ids.map((id) => String(id || "").trim()).filter(Boolean)));
  if (!uniqueIds.length) return;
  const missing = uniqueIds.filter((id) => !skuDisplayMap.value[id]);
  if (!missing.length) return;
  const results = await Promise.allSettled(missing.map((id) => skuApi.get(id)));
  const next = { ...skuDisplayMap.value };
  results.forEach((res, idx) => {
    const skuId = missing[idx];
    if (res.status === "fulfilled" && res.value) {
      next[skuId] = buildSkuLabel(res.value as any);
    }
  });
  skuDisplayMap.value = next;
};

const loadSkuMeta = async (items: OrderItem[]) => {
  await loadSkuMetaByIds(items.map((it) => String(it.skuId || "").trim()));
};

const fetchDetail = async () => {
  if (!id.value) return;
  loading.value = true;
  try {
    detail.value = await api.getOrder(id.value);
    await Promise.all([
      loadSkuMeta(detail.value?.items || []),
    ]);
    await fetchPaymentTransactions();
    await fetchManualReviews();
    await fetchBenefitReviews();
    await syncOperatorMap();
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

const fetchPaymentTransactions = async () => {
  if (!id.value) return;
  paymentLoading.value = true;
  try {
    paymentTransactions.value = await paymentsApi.listTransactions({ orderId: id.value });
  } catch (e: any) {
    toast.add({
      title: "加载支付单失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
    paymentTransactions.value = [];
  } finally {
    paymentLoading.value = false;
  }
};

const fetchManualReviews = async () => {
  if (!id.value) return;
  manualLoading.value = true;
  try {
    manualReviews.value = await paymentsApi.listManualPayments(id.value);
    manualLogs.value = await paymentsApi.listManualPaymentLogs(id.value);
    await syncOperatorMap();
  } catch (e: any) {
    toast.add({
      title: "加载收款记录失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    manualLoading.value = false;
  }
};

const fetchBenefitReviews = async () => {
  if (!id.value) return;
  benefitLoading.value = true;
  try {
    benefitReviews.value = await api.listBenefitReviews(id.value);
    benefitSelection.value = new Set();
    await syncOperatorMap();
  } catch (e: any) {
    toast.add({
      title: "加载优惠权益失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    benefitLoading.value = false;
  }
};


const openCancel = () => {
  cancelReason.value = "";
  cancelOpen.value = true;
};

const openManualPayment = () => {
  manualForm.payMethod = "bank_transfer";
  manualForm.proofNo = "";
  manualForm.note = "";
  manualOpen.value = true;
};

const closeManualPayment = () => {
  blurActiveElement();
  manualOpen.value = false;
};

const submitManualPayment = async () => {
  if (!detail.value?.summary) return;
  manualSubmitting.value = true;
  try {
    await paymentsApi.createManualPayment({
      orderId: detail.value.summary.orderId,
      amountMinor: detail.value.summary.amounts.total,
      currency: detail.value.summary.amounts.currency,
      payMethod: manualForm.payMethod,
      proofNo: manualForm.proofNo || undefined,
      note: manualForm.note || undefined,
    });
    toast.add({ title: "已提交手动收款申请", color: "success" });
    manualOpen.value = false;
    await fetchManualReviews();
  } catch (e: any) {
    toast.add({
      title: "提交失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    manualSubmitting.value = false;
  }
};

const approveManual = async (reviewId: number) => {
  manualSubmitting.value = true;
  try {
    await paymentsApi.approveManualPayment(reviewId, {});
    toast.add({ title: "审核通过", color: "success" });
    await fetchManualReviews();
    await fetchDetail();
  } catch (e: any) {
    toast.add({
      title: "审核失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    manualSubmitting.value = false;
  }
};

const openReject = (reviewId: number) => {
  currentReviewId.value = reviewId;
  rejectReason.value = "";
  rejectOpen.value = true;
};

const closeReject = () => {
  blurActiveElement();
  rejectOpen.value = false;
};

const confirmReject = async () => {
  if (!currentReviewId.value) return;
  manualSubmitting.value = true;
  try {
    await paymentsApi.rejectManualPayment(currentReviewId.value, {
      reason: rejectReason.value || "",
    });
    toast.add({ title: "已拒绝收款申请", color: "success" });
    rejectOpen.value = false;
    currentReviewId.value = null;
    await fetchManualReviews();
  } catch (e: any) {
    toast.add({
      title: "拒绝失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    manualSubmitting.value = false;
  }
};

const benefitTypeLabel = (raw: string) => {
  if (raw === "coupon") return "优惠券";
  if (raw === "giftcard") return "礼品卡";
  return raw || "-";
};

const benefitStatusLabel = (st: string) => {
  const map: Record<string, string> = {
    pending_review: "待审核",
    approved: "已通过",
    rejected: "已拒绝",
  };
  return map[st] || st || "-";
};

const benefitStatusColor = (st: string) => {
  const map: Record<string, "warning" | "success" | "neutral" | "info" | "error" | "primary"> = {
    pending_review: "warning",
    approved: "success",
    rejected: "error",
  };
  return map[st] || "neutral";
};

const benefitValueLabel = (review: OrderBenefitReview) => {
  if (review.valueType === "percent") {
    return `${Number(review.value || 0) / 100}%`;
  }
  if (review.valueType === "amount" || review.valueType === "balance") {
    return formatMoney(review.currency || detail.value?.summary?.amounts?.currency || "CNY", review.value);
  }
  return review.value || "-";
};

const toggleBenefitRow = (row: OrderBenefitReview) => {
  const set = new Set(benefitSelection.value);
  if (set.has(row.id)) {
    set.delete(row.id);
  } else if (row.status === "pending_review") {
    set.add(row.id);
  }
  benefitSelection.value = set;
};

const toggleBenefitSelectAll = () => {
  const next = new Set<number>();
  if (!benefitAllSelected.value) {
    benefitSelectableIds.value.forEach((id) => next.add(id));
  }
  benefitSelection.value = next;
};

const openBenefitModal = (type: "coupon" | "giftcard") => {
  benefitForm.type = type;
  benefitForm.code = "";
  benefitForm.valueType = type === "giftcard" ? "balance" : "amount";
  benefitForm.value = "";
  benefitForm.note = "";
  benefitForm.stackingAllowed = false;
  benefitSearchTerm.value = "";
  benefitCodeOptions.value = [];
  benefitCodeHint.value = "";
  benefitOpen.value = true;
};

const closeBenefitModal = () => {
  blurActiveElement();
  benefitOpen.value = false;
};

const loadBenefitCodes = async (keyword: string) => {
  const term = keyword.trim();
  if (!term) {
    benefitCodeOptions.value = [];
    benefitCodeHint.value = "";
    return;
  }
  benefitCodeHint.value = "";
  benefitCodeLoading.value = true;
  try {
    const items = await api.searchBenefitCodes(benefitForm.type as "coupon" | "giftcard", term);
    benefitCodeOptions.value = items;
    if (items.length === 0) {
      benefitCodeHint.value = "该权益码已被使用或不可用";
    }
  } catch (e: any) {
    benefitCodeHint.value = e?.message || "检索失败";
  } finally {
    benefitCodeLoading.value = false;
  }
};

const submitBenefitReview = async () => {
  if (!detail.value?.summary) return;
  if (!canEdit.value) {
    toast.add({ title: "当前状态不可编辑", color: "warning" });
    return;
  }
  const code = String(benefitForm.code || "").trim();
  if (!code) {
    toast.add({ title: "请选择权益码", color: "error" });
    return;
  }
  const value = Number(benefitForm.value || 0);
  if (!Number.isFinite(value) || value <= 0) {
    toast.add({ title: "请输入有效的优惠值", color: "error" });
    return;
  }
  const amountMinor = benefitAmountMinor.value;
  if (!amountMinor || amountMinor <= 0) {
    toast.add({ title: "折扣金额无效", color: "error" });
    return;
  }
  if (amountMinor > detail.value.summary.amounts.total) {
    toast.add({ title: "抵扣金额不能超过订单金额", color: "error" });
    return;
  }
  benefitSubmitting.value = true;
  try {
    await api.createBenefitReview(detail.value.summary.orderId, {
      benefitType: benefitForm.type as "coupon" | "giftcard",
      benefitCode: code,
      valueType: benefitForm.valueType as "amount" | "percent" | "balance",
      value,
      currency: detail.value.summary.amounts.currency,
      stackingAllowed: benefitForm.stackingAllowed,
      note: benefitForm.note?.trim() || undefined,
    });
    toast.add({ title: "已提交审核", color: "success" });
    benefitOpen.value = false;
    await fetchBenefitReviews();
  } catch (e: any) {
    toast.add({
      title: "提交失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    benefitSubmitting.value = false;
  }
};

const approveBenefit = async (review: OrderBenefitReview) => {
  benefitSubmitting.value = true;
  try {
    await api.approveBenefitReviews({ reviewIds: [review.id] });
    toast.add({ title: "审核通过", color: "success" });
    benefitSelection.value = new Set();
    await fetchBenefitReviews();
  } catch (e: any) {
    toast.add({
      title: "审核失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    benefitSubmitting.value = false;
  }
};

const approveSelectedBenefits = async () => {
  const ids = Array.from(benefitSelection.value);
  if (!ids.length) return;
  benefitSubmitting.value = true;
  try {
    await api.approveBenefitReviews({ reviewIds: ids });
    toast.add({ title: "批量审核通过", color: "success" });
    benefitSelection.value = new Set();
    await fetchBenefitReviews();
  } catch (e: any) {
    toast.add({
      title: "审核失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    benefitSubmitting.value = false;
  }
};

const openBenefitReject = (ids?: number[]) => {
  benefitRejectReason.value = "";
  if (ids && ids.length) {
    benefitSelection.value = new Set(ids);
  }
  benefitRejectOpen.value = true;
};

const closeBenefitReject = () => {
  blurActiveElement();
  benefitRejectOpen.value = false;
};

const confirmBenefitReject = async () => {
  const ids = Array.from(benefitSelection.value);
  if (!ids.length) return;
  const reason = benefitRejectReason.value.trim();
  if (!reason) {
    toast.add({ title: "请填写拒绝原因", color: "error" });
    return;
  }
  benefitSubmitting.value = true;
  try {
    await api.rejectBenefitReviews({ reviewIds: ids, reason });
    toast.add({ title: "已拒绝申请", color: "success" });
    benefitRejectOpen.value = false;
    benefitSelection.value = new Set();
    await fetchBenefitReviews();
  } catch (e: any) {
    toast.add({
      title: "拒绝失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    benefitSubmitting.value = false;
  }
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

const back = () => router.push("/market/orders");

const enterEdit = () => {
  if (!id.value) return;
  router.push({ path: `/market/orders/${id.value}`, query: { mode: "edit" } });
};

const exitEdit = () => {
  if (!id.value) return;
  router.push(`/market/orders/${id.value}`);
};

const resetEditForm = () => {
  const snap = detail.value?.summary?.shippingAddressSnapshot;
  Object.assign(editForm, emptyAddress(), snap || {});
};

const buildAddressPayload = (): ShippingAddress => ({
  label: editForm.label?.trim() || "",
  recipientName: editForm.recipientName?.trim() || "",
  recipientPhone: editForm.recipientPhone?.trim() || "",
  countryCode: editForm.countryCode?.trim() || "",
  province: editForm.province?.trim() || "",
  city: editForm.city?.trim() || "",
  district: editForm.district?.trim() || "",
  address1: editForm.address1?.trim() || "",
  address2: editForm.address2?.trim() || "",
  postalCode: editForm.postalCode?.trim() || "",
});

const saveShippingAddress = async () => {
  if (!detail.value?.summary) return;
  if (!canEdit.value) {
    toast.add({ title: "当前状态不可编辑", color: "warning" });
    return;
  }
  const payload = buildAddressPayload();
  if (!payload.recipientName || !payload.recipientPhone || !payload.address1) {
    toast.add({ title: "请补全收货人、联系电话和地址", color: "error" });
    return;
  }
  updating.value = true;
  try {
    const updated = await api.updateShippingAddress(detail.value.summary.orderId, payload);
    detail.value = {
      ...detail.value,
      summary: {
        ...detail.value.summary,
        ...updated,
        shippingAddressSnapshot: updated.shippingAddressSnapshot || payload,
      },
    };
    resetEditForm();
    toast.add({ title: "收货信息已更新", color: "success" });
  } catch (e: any) {
    toast.add({
      title: "更新失败",
      description: e?.message || "请稍后重试",
      color: "error",
    });
  } finally {
    updating.value = false;
  }
};

const blurActiveElement = () => {
  if (typeof document !== "undefined") {
    (document.activeElement as HTMLElement | null)?.blur();
  }
};

const manualStatusLabel = (st: string) => {
  const map: Record<string, string> = {
    pending_review: "待审核",
    approved: "已通过",
    rejected: "已拒绝",
  };
  return map[st] || st || "-";
};

const manualStatusColor = (st: string) => {
  const map: Record<string, "warning" | "success" | "neutral" | "info" | "error" | "primary"> = {
    pending_review: "warning",
    approved: "success",
    rejected: "error",
  };
  return map[st] || "neutral";
};

const manualRowStatusLabel = (row: ManualPaymentReviewLog) => {
  if (row.action === "submitted") return "已提交";
  return manualStatusLabel(row.status);
};

const manualRowStatusColor = (row: ManualPaymentReviewLog) => {
  if (row.action === "submitted") return "info";
  return manualStatusColor(row.status);
};

const manualReviewTerminalMap = computed(() => {
  const map = new Map<number, string>();
  manualLogs.value.forEach((row) => {
    if (row.action === "approved" || row.action === "rejected") {
      map.set(row.reviewId, row.action);
    }
  });
  return map;
});

const canActOnManualLog = (row: ManualPaymentReviewLog) => {
  if (row.action !== "submitted" || row.status !== "pending_review") return false;
  return !manualReviewTerminalMap.value.has(row.reviewId);
};

watch(id, () => fetchDetail(), { immediate: true });
watch(
  () => detail.value?.summary?.orderId,
  () => {
    if (isEditMode.value) {
      resetEditForm();
    }
  },
);
watch(isEditMode, (next) => {
  if (next) {
    resetEditForm();
  }
});
watch(manualOpen, (open) => {
  if (!open) {
    blurActiveElement();
  }
});
watch(rejectOpen, (open) => {
  if (!open) {
    blurActiveElement();
  }
});
watch(benefitOpen, (open) => {
  if (!open) {
    blurActiveElement();
  }
});
watch(benefitRejectOpen, (open) => {
  if (!open) {
    blurActiveElement();
  }
});
watch(
  () => benefitForm.type,
  (next) => {
    if (next === "giftcard") {
      benefitForm.valueType = "balance";
    } else if (benefitForm.valueType === "balance") {
      benefitForm.valueType = "amount";
    }
  },
);
</script>
