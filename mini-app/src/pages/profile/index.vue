<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 120px;">
    <view :style="`padding-top:${topInset}px;`" class="sticky top-0 z-10 w-full bg-background-light">
      <view class="flex items-center justify-between px-4 pt-4 pb-2">
        <view class="text-lg font-extrabold">个人中心</view>
        <view class="flex items-center gap-3">
          <view class="h-9 w-9 rounded-full bg-white shadow-sm flex items-center justify-center" hover-class="opacity-80" @tap="noop">
            <text class="text-base font-black">✉</text>
          </view>
          <view class="h-9 w-9 rounded-full bg-white shadow-sm flex items-center justify-center" hover-class="opacity-80" @tap="noop">
            <text class="text-base font-black">⚙</text>
          </view>
        </view>
      </view>
    </view>

    <view class="px-4">
      <view class="relative rounded-3xl bg-white p-4 shadow-sm overflow-hidden">
        <view class="absolute -top-8 -right-12 h-40 w-40 rounded-full" style="background: rgba(79, 138, 126, 0.12);"></view>
        <view class="absolute top-14 -right-6 h-28 w-28 rounded-full" style="background: rgba(79, 138, 126, 0.10);"></view>

        <view v-if="loggedIn" class="relative flex items-center gap-4">
          <view class="relative">
            <image class="h-20 w-20 rounded-full bg-gray-100" mode="aspectFill" :src="avatarUrl" />
            <view class="absolute bottom-0 right-0 rounded-full border border-white px-2 py-1" style="background: #4F8A7E;">
              <text class="font-extrabold text-white" style="font-size: 10px;">LV.5</text>
            </view>
          </view>
          <view class="flex-1 min-w-0">
            <view class="flex items-center gap-2">
              <text class="truncate text-xl font-extrabold">{{ displayName }}</text>
              <text class="text-sm text-muted" @tap="noop">✎</text>
            </view>
            <view class="mt-1 text-sm text-muted">Gold Member Distributor</view>
          </view>
        </view>

        <view v-else class="relative flex items-center justify-between gap-4">
          <view class="flex items-center gap-3">
            <image class="h-14 w-14 rounded-full bg-gray-100" mode="aspectFill" src="/static/icons/image-placeholder.svg" />
            <view>
              <view class="text-base font-extrabold">未登录</view>
              <view class="mt-1 text-xs text-muted">登录后查看订单、地址与更多服务</view>
            </view>
          </view>
          <button class="rounded-full bg-primary px-4 py-2 text-xs font-extrabold text-white" @tap="goLogin">
            去登录
          </button>
        </view>
      </view>

      <view class="mt-3 rounded-3xl bg-white p-4 shadow-sm">
        <view class="flex justify-between gap-3">
          <view class="flex flex-1 flex-col items-center">
            <view class="text-xl font-extrabold">{{ tokenBalanceText }}</view>
            <view class="mt-1 text-xs text-muted">代币</view>
          </view>
          <view class="my-auto h-8 w-px bg-gray-100"></view>
          <view class="flex flex-1 flex-col items-center">
            <view class="text-xl font-extrabold">{{ entitlementsCount }}</view>
            <view class="mt-1 text-xs text-muted">权益</view>
          </view>
          <view class="my-auto h-8 w-px bg-gray-100"></view>
          <view class="flex flex-1 flex-col items-center">
            <view class="text-xl font-extrabold">{{ tokenTypeCount }}</view>
            <view class="mt-1 text-xs text-muted">代币种类</view>
          </view>
        </view>
      </view>

      <view class="mt-3 rounded-3xl bg-white p-4 shadow-sm">
        <view class="flex items-center justify-between">
          <view class="text-base font-extrabold">我的订单</view>
          <view class="text-xs text-muted" hover-class="opacity-70" @tap="toOrders('all')">查看全部 ›</view>
        </view>
        <view class="mt-4 flex items-start justify-between">
          <view class="flex flex-1 flex-col items-center gap-2" hover-class="opacity-80" @tap="toOrders('pending_payment')">
            <view class="relative">
              <text class="text-2xl text-gray-500">💳</text>
              <view v-if="orderBadges.pendingPayment" class="absolute -top-1 -right-1 h-2 w-2 rounded-full bg-red-500"></view>
            </view>
            <text class="text-muted" style="font-size: 11px;">待付款</text>
          </view>
          <view class="flex flex-1 flex-col items-center gap-2" hover-class="opacity-80" @tap="toOrders('paid')">
            <text class="text-2xl text-gray-500">📦</text>
            <text class="text-muted" style="font-size: 11px;">待发货</text>
          </view>
          <view class="flex flex-1 flex-col items-center gap-2" hover-class="opacity-80" @tap="toOrders('shipped')">
            <view class="relative">
              <text class="text-2xl text-gray-500">🚚</text>
              <view
                v-if="orderBadges.shippedCount > 0"
                class="absolute -top-2 -right-2 h-4 w-4 rounded-full bg-red-500 flex items-center justify-center"
              >
                <text class="font-extrabold text-white" style="font-size: 9px;">
                  {{ orderBadges.shippedCount > 9 ? "9+" : String(orderBadges.shippedCount) }}
                </text>
              </view>
            </view>
            <text class="text-muted" style="font-size: 11px;">待收货</text>
          </view>
          <view class="flex flex-1 flex-col items-center gap-2" hover-class="opacity-80" @tap="toOrders('completed')">
            <text class="text-2xl text-gray-500">💬</text>
            <text class="text-muted" style="font-size: 11px;">评价</text>
          </view>
          <view class="flex flex-1 flex-col items-center gap-2" hover-class="opacity-80" @tap="toOrders('refund')">
            <text class="text-2xl text-gray-500">↩</text>
            <text class="text-muted" style="font-size: 11px;">售后</text>
          </view>
        </view>
      </view>

      <view class="mt-4 px-1 text-base font-extrabold">更多服务</view>
      <view class="mt-2 rounded-3xl bg-white p-4 shadow-sm">
        <view class="grid grid-cols-4 gap-y-6">
          <view class="flex flex-col items-center gap-2" hover-class="opacity-80" @tap="toAddress">
            <view class="h-10 w-10 rounded-full flex items-center justify-center" style="background: rgba(79, 138, 126, 0.10);">
              <text class="text-xl text-primary">📍</text>
            </view>
            <text class="font-medium" style="font-size: 11px;">地址</text>
          </view>
          <view class="flex flex-col items-center gap-2" hover-class="opacity-80" @tap="noop">
            <view class="h-10 w-10 rounded-full flex items-center justify-center" style="background: rgba(79, 138, 126, 0.10);">
              <text class="text-xl text-primary">❤</text>
            </view>
            <text class="font-medium" style="font-size: 11px;">收藏</text>
          </view>
          <view class="flex flex-col items-center gap-2" hover-class="opacity-80" @tap="noop">
            <view class="h-10 w-10 rounded-full flex items-center justify-center" style="background: rgba(79, 138, 126, 0.10);">
              <text class="text-xl text-primary">🎁</text>
            </view>
            <text class="font-medium" style="font-size: 11px;">邀请</text>
          </view>
          <view class="flex flex-col items-center gap-2" hover-class="opacity-80" @tap="noop">
            <view class="h-10 w-10 rounded-full flex items-center justify-center" style="background: rgba(79, 138, 126, 0.10);">
              <text class="text-xl text-primary">🎧</text>
            </view>
            <text class="font-medium" style="font-size: 11px;">客服</text>
          </view>
          <view class="flex flex-col items-center gap-2" hover-class="opacity-80" @tap="noop">
            <view class="h-10 w-10 rounded-full flex items-center justify-center" style="background: rgba(79, 138, 126, 0.10);">
              <text class="text-xl text-primary">🕒</text>
            </view>
            <text class="font-medium" style="font-size: 11px;">浏览记录</text>
          </view>
          <view class="flex flex-col items-center gap-2" hover-class="opacity-80" @tap="noop">
            <view class="h-10 w-10 rounded-full flex items-center justify-center" style="background: rgba(79, 138, 126, 0.10);">
              <text class="text-xl text-primary">❓</text>
            </view>
            <text class="font-medium" style="font-size: 11px;">帮助</text>
          </view>
        </view>
      </view>

      <view v-if="loggedIn" class="mt-4 pb-6">
        <button class="w-full rounded-2xl bg-red-50 py-3 text-sm font-extrabold text-red-600" @tap="logout">
          退出登录
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { syncTabBarSelected } from "@/utils/tabbar";
import { clearSession, getCustomerIdentifier, getCustomerName, isLoggedIn } from "@/services/session";
import { miniAppListMyOrders } from "@/services/miniapp-order";
import { miniAppGetEntitlements, miniAppGetTokenBalances } from "@/services/miniapp-membership";

const topInset = ref<number>(40);

const loggedIn = computed(() => isLoggedIn());
const displayName = computed(() => getCustomerName() || getCustomerIdentifier() || "用户");
const avatarUrl = "/static/icons/image-placeholder.svg";
const orderBadges = ref({ pendingPayment: false, shippedCount: 0 });
const entitlementsCount = ref(0);
const tokenTypeCount = ref(0);
const tokenBalance = ref(0);
const tokenBalanceText = computed(() => (tokenBalance.value > 0 ? String(tokenBalance.value) : "0"));

function goLogin() {
  uni.setStorageSync("miniapp.auth.redirect", "/pages/profile/index");
  uni.navigateTo({ url: "/pages/auth/index?tab=login" });
}

function noop() {
  uni.showToast({ title: "功能完善中", icon: "none" });
}

function toOrders(tab: string) {
  if (!loggedIn.value) return goLogin();
  const t = String(tab || "all").trim() || "all";
  uni.navigateTo({ url: `/pages/order/list?tab=${encodeURIComponent(t)}` });
}

function toAddress() {
  if (!loggedIn.value) return goLogin();
  uni.navigateTo({ url: "/pages/address/index" });
}

async function refreshOrderBadges() {
  if (!loggedIn.value) {
    orderBadges.value = { pendingPayment: false, shippedCount: 0 };
    return;
  }
  try {
    const resp = await miniAppListMyOrders({ page: 1, pageSize: 50 });
    const items = Array.isArray(resp?.items) ? resp.items : [];
    const pending = items.filter((x) => String((x as any)?.status || "").trim() === "pending_payment").length;
    const shipped = items.filter((x) => String((x as any)?.status || "").trim() === "shipped").length;
    orderBadges.value = { pendingPayment: pending > 0, shippedCount: shipped };
  } catch {
    // ignore
  }
}

async function refreshMembershipSummary() {
  if (!loggedIn.value) {
    entitlementsCount.value = 0;
    tokenTypeCount.value = 0;
    tokenBalance.value = 0;
    return;
  }
  try {
    const [entRes, tokenRes] = await Promise.all([miniAppGetEntitlements(), miniAppGetTokenBalances()]);
    const entItems = Array.isArray(entRes?.items) ? entRes.items : [];
    const tokenItems = Array.isArray(tokenRes?.items) ? tokenRes.items : [];
    entitlementsCount.value = entItems.length;
    tokenTypeCount.value = tokenItems.length;
    tokenBalance.value = tokenItems.reduce((sum, item) => sum + Number(item?.balance || 0), 0);
  } catch {
    // ignore
  }
}

async function logout() {
  const ok = await new Promise<boolean>((resolve) => {
    uni.showModal({
      title: "退出登录",
      content: "确认退出当前账号？",
      confirmText: "退出",
      confirmColor: "#ef4444",
      cancelText: "取消",
      success: (res) => resolve(Boolean(res?.confirm)),
      fail: () => resolve(false),
    });
  });
  if (!ok) return;
  clearSession();
  uni.showToast({ title: "已退出", icon: "none" });
  goLogin();
}

onMounted(() => {
  try {
    const wxAny = (globalThis as any).wx;
    if (wxAny && typeof wxAny.getWindowInfo === "function") {
      const info = wxAny.getWindowInfo();
      const statusBar = Number(info?.statusBarHeight ?? 0);
      topInset.value = Math.max(40, statusBar + 12);
      return;
    }
    const sys = uni.getSystemInfoSync();
    const statusBar = Number((sys as any).statusBarHeight ?? 0);
    topInset.value = Math.max(40, statusBar + 12);
  } catch {
    topInset.value = 40;
  }
});

onShow(() => {
  syncTabBarSelected("pages/profile/index");
  void refreshOrderBadges();
  void refreshMembershipSummary();
  if (!loggedIn.value) {
    goLogin();
  }
});
</script>
