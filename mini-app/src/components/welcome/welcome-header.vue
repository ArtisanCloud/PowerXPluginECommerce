<template>
  <view class="relative w-full px-6 pt-10 pb-4">
    <view class="flex w-full items-center justify-center gap-3">
      <view class="flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-white shadow-md">
        <text class="text-xl font-extrabold leading-none">S</text>
      </view>
      <text class="text-xl font-extrabold tracking-tight text-text-dark">{{ brand }}</text>
    </view>

    <view
      class="absolute right-6 top-10 rounded-full bg-surface-90 px-3 py-2 text-xs font-extrabold text-text-dark shadow-sm border border-line-5"
      @tap="$emit('toggle-locale')"
    >
      {{ langSwitchLabel }}
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
  brand: string;
  langSwitchLabel: string;
}>();

defineEmits<{
  (e: "toggle-locale"): void;
}>();
</script>
