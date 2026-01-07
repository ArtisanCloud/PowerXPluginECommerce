<template>
  <view class="min-h-screen w-full bg-background-light font-display text-text-dark" style="padding-bottom: 96px;">
    <view :style="`padding-top:${topInset}px;`" class="sticky top-0 z-10 w-full bg-background-light">
      <view class="px-4 pt-4 pb-2">
        <view class="text-xl font-extrabold">我的</view>
      </view>
    </view>
    <view class="px-4 py-6 text-sm text-muted">占位页：后续接入订单、地址、设置。</view>
  </view>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { onShow } from "@dcloudio/uni-app";
import { syncTabBarSelected } from "@/utils/tabbar";

const topInset = ref<number>(40);

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
});
</script>
