<template>
  <view class="min-h-screen w-full bg-background-light font-display text-gray-900" style="padding-bottom: 96px;">
    <view
      class="sticky top-0 z-50 bg-white px-4 py-3"
      :style="`padding-top:${topInset}px; backdrop-filter: blur(12px); background: rgba(255,255,255,0.9);`"
    >
      <view class="flex items-center justify-between">
        <view class="flex items-center justify-center h-10 w-10 -ml-2" hover-class="opacity-80" @tap="goBack">
          <text class="text-2xl">‹</text>
        </view>
        <text class="text-base font-extrabold">权益详情</text>
        <view class="flex items-center justify-center h-10 w-10 -mr-2" hover-class="opacity-80" @tap="onShare">
          <text class="text-xl">↗</text>
        </view>
      </view>
    </view>

    <view class="px-4 pt-4 space-y-4">
      <view class="rounded-2xl bg-white p-5 shadow-sm">
        <view class="flex items-center gap-4">
          <image class="h-12 w-12 rounded-full bg-gray-100" mode="aspectFill" :src="profile.avatarUrl || '/static/logo.png'" />
          <view class="flex-1 min-w-0">
            <view class="text-base font-extrabold">{{ profileName }}</view>
            <view class="mt-1 text-xs text-muted">当前会籍：{{ tierName }}</view>
          </view>
        </view>
      </view>

      <view class="rounded-2xl bg-white p-5 shadow-sm">
        <view class="flex items-center justify-between mb-3">
          <text class="text-sm font-extrabold">权益清单</text>
          <text class="text-xs text-muted">{{ entitlements.length }} 项</text>
        </view>
        <view v-if="entitlements.length === 0" class="py-3 text-center text-xs text-muted">暂无权益信息</view>
        <view v-else class="space-y-3">
          <view v-for="it in entitlements" :key="it.id" class="flex items-start justify-between">
            <view>
              <text class="text-sm font-semibold">{{ it.serviceCode }}</text>
              <view class="mt-1 text-xs text-muted">{{ entitlementValidText(it) }}</view>
            </view>
            <text class="text-sm font-extrabold">x{{ it.quantity }}</text>
          </view>
        </view>
      </view>

      <view class="rounded-2xl bg-white p-5 shadow-sm">
        <view class="flex items-center justify-between mb-3">
          <text class="text-sm font-extrabold">代币余额</text>
          <text class="text-xs text-muted">{{ tokenBalances.length }} 项</text>
        </view>
        <view v-if="tokenBalances.length === 0" class="py-3 text-center text-xs text-muted">暂无代币余额</view>
        <view v-else class="space-y-3">
          <view v-for="it in tokenBalances" :key="it.tokenCode" class="flex items-center justify-between">
            <text class="text-sm">{{ it.tokenCode }}</text>
            <text class="text-sm font-extrabold">{{ it.balance }}</text>
          </view>
        </view>
      </view>

      <view class="rounded-2xl bg-white p-5 shadow-sm">
        <button class="w-full rounded-full bg-primary py-3 text-sm font-extrabold text-white" @tap="goUpgrade">
          升级/续费
        </button>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { onLoad } from "@dcloudio/uni-app";
import { miniAppGetEntitlements, miniAppGetMembershipProfile, miniAppGetTokenBalances, type EntitlementItem, type TokenBalanceItem, type MembershipProfile } from "@/services/miniapp-membership";

const topInset = ref(44);
const profile = ref<MembershipProfile>({});
const entitlements = ref<EntitlementItem[]>([]);
const tokenBalances = ref<TokenBalanceItem[]>([]);

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

function goBack() {
  uni.navigateBack();
}

function onShare() {
  uni.showToast({ title: "分享待接入", icon: "none" });
}

function goUpgrade() {
  uni.navigateTo({ url: "/pages/membership/upgrade" });
}

function entitlementValidText(item: EntitlementItem) {
  const from = String(item.validFrom || "").trim();
  const to = String(item.validTo || "").trim();
  if (from && to) return `${from} - ${to}`;
  if (to) return `有效期至 ${to}`;
  return "长期有效";
}

const profileName = computed(() => String(profile.value?.customerName || "—").trim() || "—");
const tierName = computed(() => String(profile.value?.tierName || profile.value?.membershipTier || "—").trim() || "—");

async function load() {
  try {
    const p = await miniAppGetMembershipProfile();
    profile.value = p || {};
  } catch {
    profile.value = {};
  }
  try {
    const resp = await miniAppGetEntitlements();
    entitlements.value = Array.isArray(resp?.items) ? resp.items : [];
  } catch {
    entitlements.value = [];
  }
  try {
    const resp = await miniAppGetTokenBalances();
    tokenBalances.value = Array.isArray(resp?.items) ? resp.items : [];
  } catch {
    tokenBalances.value = [];
  }
}

onLoad(() => {
  ensureTopInset();
  void load();
});

onMounted(() => ensureTopInset());
</script>
