<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 96px;">
    <view class="sticky top-0 z-50 bg-background-light" :style="`padding-top:${topInset}px;`">
      <view class="px-4 pt-4 pb-2">
        <view class="text-xl font-extrabold">下单成功</view>
      </view>
    </view>

    <view class="px-4 pt-6">
      <view class="rounded-2xl bg-white p-5 shadow-sm">
        <view class="flex items-center gap-3">
          <view class="h-12 w-12 rounded-full bg-primary-10 flex items-center justify-center">
            <text class="text-primary text-lg font-black">✓</text>
          </view>
          <view class="flex-1">
            <view class="text-base font-extrabold">订单已创建</view>
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
        </view>
      </view>

      <view class="mt-4 space-y-3">
        <button class="w-full rounded-full bg-primary py-3 text-sm font-extrabold text-white" @tap="toMall">继续逛逛</button>
        <button class="w-full rounded-full bg-white py-3 text-sm font-extrabold text-gray-900 shadow-sm" @tap="toProfile">
          查看我的订单
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";

const topInset = ref(44);

const orderNo = ref("");
const orderId = ref("");
const status = ref("pending_payment");
const currency = ref("CNY");
const totalMinor = ref(0);

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
  return s || "—";
});

function toMall() {
  uni.switchTab({ url: "/pages/mall/index" });
}

function toProfile() {
  uni.navigateTo({ url: "/pages/order/list" });
}

onLoad((options) => {
  ensureTopInset();
  orderNo.value = String(options?.orderNo || "").trim();
  orderId.value = String(options?.orderId || "").trim();
  currency.value = String(options?.currency || "CNY").trim() || "CNY";
  totalMinor.value = Number(options?.total || 0) || 0;
});

onMounted(() => ensureTopInset());
</script>
