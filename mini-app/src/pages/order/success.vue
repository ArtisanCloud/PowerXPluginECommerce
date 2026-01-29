<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 96px;">
    <view class="sticky top-0 z-50 bg-background-light" :style="`padding-top:${topInset}px;`">
      <view class="px-4 pt-4 pb-2">
        <view class="text-xl font-extrabold">{{ titleText }}</view>
      </view>
    </view>

    <view class="px-4 pt-6">
      <view class="rounded-2xl bg-white p-5 shadow-sm">
        <view class="flex items-center gap-3">
          <view class="h-12 w-12 rounded-full flex items-center justify-center" :style="badgeStyle">
            <text class="text-lg font-black" :style="badgeTextStyle">{{ badgeIcon }}</text>
          </view>
          <view class="flex-1">
            <view class="text-base font-extrabold">{{ headlineText }}</view>
            <view class="mt-1 text-xs text-muted">状态：{{ statusText }}</view>
          </view>
        </view>

        <view class="mt-4 space-y-2 text-sm">
          <view class="flex items-center justify-between">
            <text class="text-muted">订单号</text>
            <text class="font-semibold">{{ orderNo || "—" }}</text>
          </view>
          <view class="flex items-center justify-between">
            <text class="text-muted">应付金额</text>
            <text class="font-extrabold text-primary">{{ totalText }}</text>
          </view>
          <view v-if="failureReason" class="text-xs" style="color:#ef4444;">{{ failureReason }}</view>
        </view>
      </view>

      <view class="mt-4 space-y-3">
        <button v-if="canPay" class="w-full rounded-full bg-primary py-3 text-sm font-extrabold text-white" @tap="onPay">
          {{ isPaying ? "支付中..." : "立即支付" }}
        </button>
        <button v-else class="w-full rounded-full bg-primary py-3 text-sm font-extrabold text-white" @tap="toMall">继续逛逛</button>
        <button class="w-full rounded-full bg-white py-3 text-sm font-extrabold text-gray-900 shadow-sm" @tap="toProfile">
          查看我的订单
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onLoad, onUnload } from "@dcloudio/uni-app";
import {
  buildIdempotencyKey,
  createPaymentTransaction,
  getWechatProviderIdNumber,
  getPaymentTransactionStatus,
  requestMiniAppPayment,
} from "@/services/miniapp-payment";
import { miniAppGetMyOrder } from "@/services/miniapp-order";

const topInset = ref(44);

const orderNo = ref("");
const orderId = ref("");
const status = ref("pending_payment");
const failureReason = ref("");
const currency = ref("CNY");
const totalMinor = ref(0);
const transactionId = ref("");
const isPaying = ref(false);
let pollTimer: ReturnType<typeof setTimeout> | null = null;

function ensureTopInset() {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(44, statusBar + 12);
      return;
    }
    const sys = uni.getSystemInfoSync();
    const statusBar = Number((sys as any).statusBarHeight ?? 0);
    topInset.value = Math.max(44, statusBar + 12);
  } catch {
    topInset.value = 44;
  }
}

function formatMoneyFromMinor(cur: string, minor: number) {
  const c = String(cur || "CNY").trim().toUpperCase() || "CNY";
  const symbol = c === "CNY" || c === "RMB" ? "¥" : c === "USD" ? "$" : c === "EUR" ? "€" : `${c} `;
  const n = Number(minor);
  if (!Number.isFinite(n)) return `${symbol}--`;
  const major = n / 100;
  const text = major.toFixed(2);
  return `${symbol}${text}`;
}

const totalText = computed(() => formatMoneyFromMinor(currency.value, totalMinor.value));
const statusText = computed(() => {
  const s = String(status.value || "").trim();
  if (s === "pending_payment") return "待支付";
  if (s === "paying") return "支付处理中";
  if (s === "paid") return "支付成功";
  if (s === "failed") return "支付失败";
  if (s === "canceled") return "已取消";
  if (s === "timeout") return "支付超时";
  return s || "—";
});

const titleText = computed(() => {
  if (status.value === "paid") return "支付成功";
  if (status.value === "paying") return "支付处理中";
  if (status.value === "failed") return "支付失败";
  if (status.value === "canceled") return "支付已取消";
  if (status.value === "timeout") return "支付超时";
  return "下单成功";
});

const headlineText = computed(() => {
  if (status.value === "paid") return "支付成功";
  if (status.value === "paying") return "支付处理中";
  if (status.value === "failed") return "支付失败";
  if (status.value === "canceled") return "支付已取消";
  if (status.value === "timeout") return "支付超时";
  return "订单已创建";
});

const badgeIcon = computed(() => {
  if (status.value === "paid") return "✓";
  if (status.value === "paying") return "…";
  if (status.value === "failed") return "!";
  if (status.value === "canceled") return "×";
  if (status.value === "timeout") return "!";
  return "✓";
});

const badgeStyle = computed(() => {
  if (status.value === "paid") return "background:#e7f6ed;";
  if (status.value === "paying") return "background:#fff4e5;";
  if (status.value === "failed" || status.value === "timeout") return "background:#fee2e2;";
  if (status.value === "canceled") return "background:#f3f4f6;";
  return "background:#eef2ff;";
});

const badgeTextStyle = computed(() => {
  if (status.value === "paid") return "color:#16a34a;";
  if (status.value === "paying") return "color:#f97316;";
  if (status.value === "failed" || status.value === "timeout") return "color:#ef4444;";
  if (status.value === "canceled") return "color:#6b7280;";
  return "color:#4F6F52;";
});

const canPay = computed(() => {
  const s = String(status.value || "").trim();
  return s === "pending_payment" || s === "failed" || s === "canceled" || s === "timeout";
});

function toMall() {
  uni.switchTab({ url: "/pages/mall/index" });
}

function toProfile() {
  if (orderId.value) return uni.navigateTo({ url: `/pages/order/detail?id=${encodeURIComponent(orderId.value)}` });
  uni.navigateTo({ url: "/pages/order/list" });
}

function stopPolling() {
  if (pollTimer) clearTimeout(pollTimer);
  pollTimer = null;
}

function startPolling() {
  stopPolling();
  if (!transactionId.value) return;
  const startedAt = Date.now();
  const poll = async () => {
    if (Date.now() - startedAt > 10000) return;
    try {
      const res = await getPaymentTransactionStatus(transactionId.value);
      if (res?.status) status.value = String(res.status || "").trim() || status.value;
      if (res?.failureReason) failureReason.value = String(res.failureReason || "").trim();
      if (!canPay.value && status.value !== "paying") return;
    } catch {
      // ignore
    }
    pollTimer = setTimeout(poll, 2000);
  };
  void poll();
}

async function refreshOrder() {
  if (!orderId.value) return;
  try {
    const detail = await miniAppGetMyOrder(orderId.value);
    if (detail?.summary?.status) status.value = String(detail.summary.status || "").trim();
    if (detail?.summary?.amounts?.currency) currency.value = String(detail.summary.amounts.currency || "").trim();
    if (typeof detail?.summary?.amounts?.total === "number") totalMinor.value = Number(detail.summary.amounts.total || 0) || totalMinor.value;
  } catch {
    // ignore
  }
}

async function onPay() {
  if (isPaying.value) return;
  if (!orderId.value || !orderNo.value) return;
  const providerId = getWechatProviderIdNumber();
  if (!providerId) {
    uni.showToast({ title: "支付渠道未配置", icon: "none" });
    return;
  }
  isPaying.value = true;
  try {
    const resp = await createPaymentTransaction({
      orderId: orderId.value,
      payMethod: "wechat_jsapi",
      providerId,
      client: "miniapp",
      idempotencyKey: buildIdempotencyKey("pay"),
    });
    transactionId.value = String(resp?.transactionId || transactionId.value || "");
    let nextStatus = "paying";
    if (resp?.wechat?.appId) {
      const outcome = await requestMiniAppPayment(resp.wechat);
      if (outcome.status === "cancel") nextStatus = "canceled";
      if (outcome.status === "fail") nextStatus = "failed";
      if (outcome.message) uni.showToast({ title: outcome.message, icon: "none" });
    } else {
      nextStatus = "pending_payment";
      uni.showToast({ title: "支付渠道未配置", icon: "none" });
    }
    status.value = nextStatus;
    if (nextStatus === "paying") startPolling();
  } catch (e: any) {
    uni.showToast({ title: e?.message || "支付发起失败", icon: "none" });
  } finally {
    isPaying.value = false;
  }
}

onLoad((options) => {
  ensureTopInset();
  orderNo.value = String(options?.orderNo || "").trim();
  orderId.value = String(options?.orderId || "").trim();
  currency.value = String(options?.currency || "CNY").trim() || "CNY";
  totalMinor.value = Number(options?.total || 0) || 0;
  status.value = String(options?.status || status.value || "pending_payment").trim();
  transactionId.value = String(options?.transactionId || "").trim();
});

onMounted(() => {
  ensureTopInset();
  void refreshOrder();
  startPolling();
});

onUnload(() => stopPolling());
</script>
