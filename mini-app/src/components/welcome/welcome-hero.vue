<template>
  <view class="w-full px-4">
    <view
      class="relative mx-auto w-full max-w-md overflow-hidden rounded-2xl bg-secondary-accent shadow-card"
      style="height: 560rpx;"
    >
      <image class="h-full w-full" mode="aspectFill" :src="imageSrc" />

      <view class="absolute inset-0 bg-gradient-to-t from-background-light via-transparent to-transparent opacity-60" />

      <view
        class="absolute right-4 top-4 flex items-center gap-2 rounded-full border border-line-5 bg-surface-95 px-3 py-2 shadow-sm"
      >
        <view class="h-2 w-2 rounded-full bg-primary" />
        <text class="text-xs font-extrabold uppercase tracking-wider text-text-dark">
          {{ badge }}
        </text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
defineOptions({
  options: {
    addGlobalClass: true,
    styleIsolation: "apply-shared",
  },
});

defineProps<{
  imageSrc: string;
  badge: string;
}>();
</script>
